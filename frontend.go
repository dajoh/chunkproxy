package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/groupcache"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func frontendRun() {
	http.HandleFunc("/", frontendHandleTop)
	http.HandleFunc("/stats", frontendHandleStats)
	http.ListenAndServe(webAddr, nil)
}

func frontendHandleTop(rw http.ResponseWriter, req *http.Request) {
	frontendAddHeaders(rw)

	var file string = req.URL.String()[1:]
	var hasRange bool
	var rangeLow int64
	var rangeHigh int64 = -1
	var rangeHeader string = req.Header.Get("Range")

	if file == "" {
		return
	}

	if rangeHeader != "" {
		hasRange = true
		rangeLow, rangeHigh = parseRangeHeader(rangeHeader)
	}

	if hasRange {
		log.Println("Range request for", file, "=>", rangeLow, "to", rangeHigh)
	} else {
		log.Println("Normal request for", file)
	}

	size, err := cacheGetFileSize(file)
	if err != nil {
		if strings.HasPrefix(err.Error(), "404") {
			http.NotFound(rw, req)
			return
		}

		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if hasRange && rangeLow > size {
		http.Error(rw, "", http.StatusRequestedRangeNotSatisfiable)
		return
	}

	if rangeHigh == -1 || (hasRange && rangeHigh >= size) {
		rangeHigh = size - 1
	}

	if hasRange {
		contentRange := fmt.Sprintf("bytes %d-%d/%d", rangeLow, rangeHigh, size)
		contentLength := fmt.Sprintf("%d", rangeHigh-rangeLow+1)

		rw.Header().Add("Content-Range", contentRange)
		rw.Header().Add("Content-Length", contentLength)

		rw.WriteHeader(http.StatusPartialContent)
	} else {
		rw.Header().Add("Content-Length", fmt.Sprintf("%d", size))
	}

	for offset := rangeLow; offset <= rangeHigh; {
		chunk := offset / chunkSize
		chunkView := groupcache.ByteView{}
		chunkOffset := offset % chunkSize

		chunkView, err := cacheGetFileChunk(file, chunk)
		if err != nil {
			log.Println("Error getting chunk:", err)
			return
		}

		reader := chunkView.Reader()
		written := int64(0)
		totalLeft := rangeHigh - offset + 1
		chunkLeft := int64(chunkView.Len()) - chunkOffset

		reader.Seek(chunkOffset, 0)

		if totalLeft < chunkLeft {
			written, err = io.CopyN(rw, reader, totalLeft)
		} else {
			written, err = io.Copy(rw, reader)
		}

		if err != nil {
			return
		}

		offset += written
	}
}

func parseRangeHeader(rangeHeader string) (int64, int64) {
	rangeRegex := regexp.MustCompile("bytes=([\\d]*)-([\\d]*)")
	rangeMatch := rangeRegex.FindAllStringSubmatch(rangeHeader, -1)

	rangeLowStr := rangeMatch[0][1]
	rangeHighStr := rangeMatch[0][2]

	rangeLow, err1 := strconv.ParseInt(rangeLowStr, 10, 64)
	rangeHigh, err2 := strconv.ParseInt(rangeHighStr, 10, 64)

	if err1 != nil {
		rangeLow = 0
	}

	if err2 != nil {
		rangeHigh = -1
	}

	return rangeLow, rangeHigh
}

func frontendHandleStats(rw http.ResponseWriter, req *http.Request) {
	frontendAddHeaders(rw)

	stats := map[string]interface{}{
		"SizeStats":      sizeCache.Stats,
		"SizeHotCache":   sizeCache.CacheStats(groupcache.HotCache),
		"SizeMainCache":  sizeCache.CacheStats(groupcache.MainCache),
		"ChunkStats":     chunkCache.Stats,
		"ChunkHotCache":  chunkCache.CacheStats(groupcache.HotCache),
		"ChunkMainCache": chunkCache.CacheStats(groupcache.MainCache),
	}

	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Write(data)
}

func frontendAddHeaders(rw http.ResponseWriter) {
	rw.Header().Add("Server", "chunkproxy/2.8")
	rw.Header().Add("X-Powered-By", "groupcache")
	rw.Header().Add("Accept-Ranges", "bytes")
}
