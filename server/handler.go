package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/mudkipme/inazuma/config"
)

var httpClient *http.Client

// purgeHandler sends a purge request to Kafka and returns a status accepted response.
func purgeHandler(w http.ResponseWriter, r *http.Request, conf *config.Config) {
	cacheKey, purge := getCacheKey(r, conf.CachePatterns)
	if !purge {
		return
	}

	sendPurgeRequestToKafka(cacheKey)
	w.WriteHeader(http.StatusAccepted)
}

// directProxy proxies the request to the upstream server and returns the response.
func directProxy(w http.ResponseWriter, r *http.Request, conf *config.Config) {
	upstreamRequest, err := http.NewRequestWithContext(r.Context(), r.Method, fmt.Sprintf("%s%s", conf.UpstreamURL, r.URL.RequestURI()), r.Body)
	if err != nil {
		http.Error(w, "Error creating upstream request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	copyHeaders(upstreamRequest.Header, r.Header)
	upstreamRequest.Header.Set("X-Forwarded-For", r.RemoteAddr)

	resp, err := httpClient.Do(upstreamRequest)
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

// getHandler retrieves the content from the cache or fetches it from the upstream server if not found.
func getHandler(w http.ResponseWriter, r *http.Request, conf *config.Config) {
	if hasBypassCacheCookie(r, conf.CookieToBypassCache) || containsNonUTMQueryParameters(r.URL) {
		directProxy(w, r, conf)
		return
	}

	// Get cache key based on Accept-Language header
	urlPath, _ := getCacheKey(r, conf.CachePatterns)
	if urlPath == "" {
		directProxy(w, r, conf)
		return
	}

	object, err := s3Client.GetObject(r.Context(), conf.Storage.S3Bucket, urlPath, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Error reading cache: %s\n", err)
		directProxy(w, r, conf)
		return
	}

	stat, err := object.Stat()
	if err != nil {
		log.Printf("Cache miss: %s\n", urlPath)
		err = updateCache(r.Context(), urlPath, conf)
		if err != nil {
			log.Printf("Failed to update cache: %v\n", err)
			directProxy(w, r, conf)
			return
		}

		object, err = s3Client.GetObject(r.Context(), conf.Storage.S3Bucket, urlPath, minio.GetObjectOptions{})
		if err != nil {
			log.Printf("Error reading cache after update: %v\n", err)
			directProxy(w, r, conf)
			return
		}
	}

	content, err := io.ReadAll(object)
	if err != nil {
		log.Printf("Error reading cache: %s\n", err)
		directProxy(w, r, conf)
		return
	}

	w.Header().Set("Content-Type", stat.ContentType)
	log.Printf("Cache hit: %s\n", urlPath)
	w.Write(content)
}

func init() {
	httpClient = &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
