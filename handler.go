package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Get cache key suffix based on Accept-Language header
	cacheKeySuffix := getCacheKeySuffix(r)

	// Update the URL path to include the cache key suffix
	urlPath := r.URL.Path + cacheKeySuffix

	if r.Method != "GET" || hasBypassCacheCookie(r) || containsNonUTMQueryParameters(r.URL) {
		directProxy(w, r)
	} else if r.Method == "PURGE" {
		sendPurgeRequestToKafka(r.URL.Path)
		w.WriteHeader(http.StatusAccepted)
	} else {
		getFromCacheOrProxy(w, r, urlPath)
	}
}

func directProxy(w http.ResponseWriter, r *http.Request) {
	upstreamRequest, err := http.NewRequest(r.Method, fmt.Sprintf("%s%s", upstreamURL, r.URL.Path), r.Body)
	if err != nil {
		http.Error(w, "Error creating upstream request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	copyHeaders(upstreamRequest.Header, r.Header)
	upstreamRequest.Header.Set("X-Forwarded-For", r.RemoteAddr)

	resp, err := http.DefaultClient.Do(upstreamRequest)
	if err != nil {
		http.Error(w, "Error fetching from upstream", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
	log.Printf("Direct proxy: %s %s\n", r.Method, r.URL.Path)
}

func getFromCacheOrProxy(w http.ResponseWriter, r *http.Request, urlPath string) {
	object, err := s3Client.GetObject(r.Context(), cacheBucketName, urlPath, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Cache miss: %s\n", urlPath)
		resp, err := http.Get(fmt.Sprintf("%s%s", upstreamURL, r.URL.Path))
		if err != nil {
			http.Error(w, "Error fetching from upstream", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading upstream response", http.StatusInternalServerError)
			return
		}

		_, err = s3Client.PutObject(r.Context(), cacheBucketName, urlPath, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{})
		if err != nil {
			log.Printf("Error saving cache: %s\n", err)
		} else {
			log.Printf("Cache saved: %s\n", urlPath)
		}

		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	content, err := ioutil.ReadAll(object)
	if err != nil {
		log.Printf("Error reading cache: %s\n", err)
		directProxy(w, r)
		return
	}

	log.Printf("Cache hit: %s\n", urlPath)
	w.Write(content)
}
