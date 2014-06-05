package main

import (
	"flag"
	"runtime/debug"
	"time"
)

var (
	baseUrlFlag      = flag.String("url", "http://dajoh.io/store/", "Backend store base URL.")
	cacheDirFlag     = flag.String("dir", "cache", "Disk cache directory.")
	coordUrlFlag     = flag.String("coord", "", "Peer discovery URL.")
	coordClusterFlag = flag.String("cluster", "", "Peer discovery cluster name.")

	webAddrFlag   = flag.String("webaddr", ":8080", "Frontend listen address.")
	poolAddrFlag  = flag.String("pooladdr", "127.0.0.1:7070", "Cache pool listen address.")
	poolPeersFlag = flag.String("poolpeers", "127.0.0.1:7070", "List of cache pool peers, comma separated.")

	chunkSizeFlag      = flag.Int64("size-c", 16, "Chunk size in MiB.")
	sizeCacheSizeFlag  = flag.Int64("size-s", 16, "Size cache size in MiB.")
	diskCacheSizeFlag  = flag.Int64("size-d", 768, "Disk cache size in MiB.")
	chunkCacheSizeFlag = flag.Int64("size-m", 512, "Memory cache size in MiB.")

	baseUrl      string
	cacheDir     string
	coordUrl     string
	coordCluster string

	webAddr   string
	poolAddr  string
	poolPeers string

	chunkSize      int64
	sizeCacheSize  int64
	diskCacheSize  int64
	chunkCacheSize int64
)

func main() {
	flag.Parse()

	baseUrl = *baseUrlFlag
	cacheDir = *cacheDirFlag
	coordUrl = *coordUrlFlag
	coordCluster = *coordClusterFlag

	webAddr = *webAddrFlag
	poolAddr = *poolAddrFlag
	poolPeers = *poolPeersFlag

	chunkSize = *chunkSizeFlag * 1024 * 1024
	diskCacheSize = *diskCacheSizeFlag * 1024 * 1024
	sizeCacheSize = *sizeCacheSizeFlag * 1024 * 1024
	chunkCacheSize = *chunkCacheSizeFlag * 1024 * 1024

	go freeOSMemory()

	discoveryInit()
	diskInit()
	cacheInit()
	frontendRun()
}

func freeOSMemory() {
	for {
		time.Sleep(15 * time.Second)
		debug.FreeOSMemory()
	}
}
