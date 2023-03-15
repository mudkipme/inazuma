package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	s3Client    *minio.Client
	redisClient *redis.Client

	upstreamURL         = os.Getenv("UPSTREAM_SERVER")
	cacheBucketName     = os.Getenv("S3_BUCKET")
	cookieToBypassCache = os.Getenv("COOKIE_TO_BYPASS")
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{path:.*}", proxyHandler).Methods("GET", "PURGE")

	err := setupS3Client()
	if err != nil {
		log.Fatal(err)
	}

	ensureBucketExists(cacheBucketName)

	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})

	// Test Redis connection
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	err = setupKafkaProducer()
	if err != nil {
		log.Fatal(err)
	}
	defer kafkaProducer.Close()

	go startKafkaConsumer()

	log.Println("Starting proxy server on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func setupS3Client() error {
	endpoint := os.Getenv("S3_ENDPOINT")          // Set the S3-like object storage endpoint
	accessKeyID := os.Getenv("S3_ACCESS_KEY")     // Set the S3-like object storage access key ID
	secretAccessKey := os.Getenv("S3_SECRET_KEY") // Set the S3-like object storage secret access key

	var err error
	s3Client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	return err
}

func ensureBucketExists(bucketName string) {
	exists, err := s3Client.BucketExists(context.Background(), bucketName)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		err = s3Client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatal(err)
		}
	}
}
