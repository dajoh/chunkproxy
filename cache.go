package main

import (
	"fmt"
	"github.com/golang/groupcache"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var pool *groupcache.HTTPPool
var sizeCache *groupcache.Group
var chunkCache *groupcache.Group

func cacheInit() {
	fixedAddr := "http://" + poolAddr
	fixedPeers := strings.Split(poolPeers, ",")

	for index, peer := range fixedPeers {
		fixedPeers[index] = "http://" + peer
	}

	log.Println("Cache pool peers:", fixedPeers)
	log.Println("Cache pool address:", fixedAddr)
	fmt.Println()

	pool = groupcache.NewHTTPPool(fixedAddr)
	pool.Set(fixedPeers...)

	sizeCache = groupcache.NewGroup("size-cache", sizeCacheSize, groupcache.GetterFunc(cacheSizeGetter))
	chunkCache = groupcache.NewGroup("chunk-cache", chunkCacheSize, groupcache.GetterFunc(cacheChunkGetter))

	go http.ListenAndServe(poolAddr, pool)
}

func cacheGetFileSize(file string) (int64, error) {
	var size string

	err := sizeCache.Get(nil, file, groupcache.StringSink(&size))
	if err != nil {
		return -1, err
	}

	return strconv.ParseInt(size, 16, 64)
}

func cacheGetFileChunk(file string, chunk int64) (groupcache.ByteView, error) {
	var key = fmt.Sprintf("%s#%d", file, chunk)
	var view = groupcache.ByteView{}

	err := chunkCache.Get(nil, key, groupcache.ByteViewSink(&view))
	if err != nil {
		return view, err
	}

	return view, nil
}

func cacheUpdatePeerList(peerList []string) {
	pool.Set(peerList...)
}

func cacheSizeGetter(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	size, err := backendGetFileSize(key)
	if err != nil {
		return err
	}

	dest.SetString(strconv.FormatInt(size, 16))
	return nil
}

func cacheChunkGetter(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	sep := strings.LastIndex(key, "#")
	file := key[:sep]
	chunk := key[sep+1:]

	chunkNum, err := strconv.ParseInt(chunk, 10, 64)
	if err != nil {
		return err
	}

	chunkData, err := diskGetFileChunk(file, chunkNum)
	if err != nil {
		return err
	}

	dest.SetBytes(chunkData)
	return nil
}
