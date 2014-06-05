package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

func backendGetFileSize(file string) (int64, error) {
	log.Println("Retrieving size of", file, "from backend")

	resp, err := http.Head(baseUrl + file)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, errors.New(resp.Status)
	}

	return strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
}

func backendGetFileChunk(file string, chunk int64) ([]byte, error) {
	client := new(http.Client)

	log.Println("Retrieving chunk", chunk, "of", file, "from backend")

	req, err := http.NewRequest("GET", baseUrl+file, nil)
	if err != nil {
		return nil, err
	}

	rangeLow := chunk * chunkSize
	rangeHigh := (chunk+1)*chunkSize - 1
	rangeValue := fmt.Sprintf("bytes=%d-%d", rangeLow, rangeHigh)
	req.Header.Add("Range", rangeValue)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return nil, errors.New(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
