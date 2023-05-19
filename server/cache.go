package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/mudkipme/inazuma/config"
)

// updateCache fetches content from the upstream server and saves it to cache.
func updateCache(ctx context.Context, urlPath string, conf *config.Config) error {
	req, err := http.NewRequestWithContext(ctx, "GET", conf.UpstreamURL+urlPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for cache update: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch content from upstream server for cache update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		if err := saveToCache(ctx, conf.Storage.S3Bucket, urlPath, resp.Header.Get("Content-Type"), resp.Body); err != nil {
			return fmt.Errorf("failed to save content to cache: %w", err)
		}
	} else {
		return fmt.Errorf("upstream server returned an error status for cache update: %d", resp.StatusCode)
	}

	return redisClient.Set(ctx, urlPath, time.Now().UTC().Format(time.RFC3339), 0).Err()
}

// saveToCache saves the response body to the cache using the provided cacheBucketName and cacheKey.
func saveToCache(ctx context.Context, cacheBucketName, cacheKey, contentType string, reader io.ReadCloser) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading upstream response: %w", err)
	}
	_, err = s3Client.PutObject(ctx, cacheBucketName, cacheKey, bytes.NewReader(body), int64(len(body)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("error saving content to cache with cache key %s: %w", cacheKey, err)
	}
	return nil
}
