package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
)

func updateCache(urlPath string) {
	cacheKeys := []string{urlPath, urlPath + "_zh-hans", urlPath + "_zh-hant"}
	acceptLanguages := []string{"", "zh-hans", "zh-hant"}

	for i, cacheKey := range cacheKeys {
		req, err := http.NewRequest("GET", upstreamURL+urlPath, nil)
		if err != nil {
			log.Printf("Failed to create request for cache update: %v\n", err)
			return
		}

		if acceptLanguages[i] != "" {
			req.Header.Set("Accept-Language", acceptLanguages[i])
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Failed to fetch content from upstream server for cache update: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if err := saveToCache(cacheKey, resp.Body); err != nil {
				log.Printf("Failed to save content to cache: %v\n", err)
			}
		} else {
			log.Printf("Upstream server returned an error status for cache update: %d\n", resp.StatusCode)
		}
	}
}

func saveToCache(cacheKey string, reader io.ReadCloser) error {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Printf("Error reading upstream response: %s\n", err)
		return err
	}
	_, err = s3Client.PutObject(context.Background(), cacheBucketName, cacheKey, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{})
	return err
}
