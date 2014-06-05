package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"os"
	"strings"
	"time"
)

const coordExpireTime = 30
const coordUpdateInterval = 15 * time.Second
const coordBackoffInterval = 10 * time.Second

func discoveryInit() {
	if coordUrl == "" {
		return
	} else if coordCluster == "" {
		log.Fatalln("Error: Peer discovery cluster not set!")
	}

	log.Println("Doing peer discovery using URL:", coordUrl)
	log.Println("Doing peer discovery using cluster:", coordCluster)
	fmt.Println()

	client := etcd.NewClient([]string{coordUrl})
	peerKey := "/" + coordCluster + "/" + poolAddr
	peerValue := strings.Join(os.Args[1:], " ")
	clusterKey := "/" + coordCluster

	resp, err := client.Create(peerKey, peerValue, coordExpireTime)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	resp, err = client.Get(clusterKey, true, true)
	if err != nil {
		log.Fatalln("Error:", err)
	}

	for idx, node := range resp.Node.Nodes {
		sep := strings.LastIndex(node.Key, "/")
		addr := node.Key[sep+1:]

		if idx == 0 {
			poolPeers = addr
		} else {
			poolPeers = poolPeers + "," + addr
		}
	}

	go discoveryWatch(clusterKey, client)
	go discoveryUpdate(peerKey, peerValue, client)
}

func discoveryWatch(clusterKey string, client *etcd.Client) {
	var err error
	var list []string
	var resp *etcd.Response

	for {
		resp, err = client.Watch(clusterKey, 0, true, nil, nil)
		if err != nil {
			log.Println("Discovery watch error:", err)
			time.Sleep(coordBackoffInterval)
			continue
		}

		if resp.Action != "create" && resp.Action != "expire" {
			continue
		}

		resp, err = client.Get(clusterKey, true, true)
		if err != nil {
			log.Println("Discovery peer list get error:", err)
			time.Sleep(coordBackoffInterval)
			continue
		}

		list = make([]string, resp.Node.Nodes.Len())

		for idx, node := range resp.Node.Nodes {
			sep := strings.LastIndex(node.Key, "/")
			addr := node.Key[sep+1:]
			list[idx] = "http://" + addr
		}

		log.Println("List of peers updated:", list)
		cacheUpdatePeerList(list)
	}
}

func discoveryUpdate(peerKey, peerValue string, client *etcd.Client) {
	for {
		time.Sleep(coordUpdateInterval)

		_, err := client.Set(peerKey, peerValue, coordExpireTime)
		if err != nil {
			log.Println("Discovery update error:", err)
			continue
		}
	}
}
