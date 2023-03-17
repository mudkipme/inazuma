package server

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/mudkipme/inazuma/config"
)

// purgeHandler sends a purge request to Kafka and returns a status accepted response.
func purgeHandler(w http.ResponseWriter, r *http.Request, conf *config.Configuration) {
	sendPurgeRequestToKafka(r.URL.Path)
	w.WriteHeader(http.StatusAccepted)
}

// directProxy proxies the request to the upstream server and returns the response.
func directProxy(w http.ResponseWriter, r *http.Request, conf *config.Configuration) {
	upstreamRequest, err := http.NewRequestWithContext(r.Context(), r.Method, fmt.Sprintf("%s%s", conf.UpstreamURL, r.URL.Path), r.Body)
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

// getHandler retrieves the content from the cache or fetches it from the upstream server if not found.
func getHandler(w http.ResponseWriter, r *http.Request, conf *config.Configuration) {
	if hasBypassCacheCookie(r, conf.CookieToBypassCache) || containsNonUTMQueryParameters(r.URL) {
		directProxy(w, r, conf)
		return
	}

	// Get cache key suffix based on Accept-Language header
	cacheKeySuffix := getCacheKeySuffix(r)

	// Update the URL path to include the cache key suffix
	urlPath := r.URL.Path + cacheKeySuffix

	object, err := s3Client.GetObject(r.Context(), conf.S3Bucket, urlPath, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Cache miss: %s\n", urlPath)
		err = updateCache(r.Context(), urlPath, conf)
		if err != nil {
			log.Printf("Failed to update cache: %v\n", err)
			directProxy(w, r, conf)
			return
		}

		object, err = s3Client.GetObject(r.Context(), conf.S3Bucket, urlPath, minio.GetObjectOptions{})
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

	stat, err := object.Stat()
	if err != nil {
		directProxy(w, r, conf)
		return
	}

	w.Header().Set("Content-Type", stat.ContentType)
	log.Printf("Cache hit: %s\n", urlPath)
	w.Write(content)
}
