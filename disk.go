package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

var diskCache = map[string]time.Time{}
var diskCacheLock sync.RWMutex

func diskInit() {
	err := os.Mkdir(cacheDir, 0776)
	if err != nil {
		if !os.IsExist(err) {
			panic(err)
		}
	}

	files, err := ioutil.ReadDir(cacheDir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		diskCache[cacheDir+"/"+file.Name()] = time.Now()
	}
}

func diskGetFileChunk(file string, chunk int64) ([]byte, error) {
	var err error
	var path string = fmt.Sprintf("%s/%s#%d", cacheDir, file, chunk)
	var data []byte

	diskCacheLock.RLock()

	data, err = ioutil.ReadFile(path)
	if err == nil {
		diskCacheLock.RUnlock()
		return data, nil
	}

	diskCacheLock.RUnlock()

	data, err = backendGetFileChunk(file, chunk)
	if err != nil {
		return nil, err
	}

	diskCacheLock.Lock()
	diskRemoveIfNeeded()

	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		diskCacheLock.Unlock()
		return nil, err
	}

	diskCache[path] = time.Now()
	diskCacheLock.Unlock()

	return data, nil
}

func diskRemoveIfNeeded() {
	var entries int64 = int64(len(diskCache) + 1)
	var totalSize int64 = entries * chunkSize

	if (totalSize + chunkSize) > diskCacheSize {
		var oldestFile string
		var oldestTime time.Time

		for file, time := range diskCache {
			if oldestFile == "" {
				oldestFile = file
				oldestTime = time
			} else if time.Before(oldestTime) {
				oldestFile = file
				oldestTime = time
			}
		}

		log.Println("Removing disk cached chunk", oldestFile)
		os.Remove(oldestFile)
		delete(diskCache, oldestFile)
	}
}
