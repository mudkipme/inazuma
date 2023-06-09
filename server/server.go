package server

import (
	"context"
	"log"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mudkipme/inazuma/config"
	"github.com/redis/go-redis/v9"
)

var (
	s3Client    *minio.Client
	redisClient *redis.Client
)

func StartServer(conf *config.Config) {
	err := setupS3Client(conf)
	if err != nil {
		log.Fatal(err)
	}

	ensureBucketExists(conf.Storage.S3Bucket)

	redisClient = redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})

	// Test Redis connection
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	setupKafkaProducer(conf)
	defer kafkaWriter.Close()
	defer kafkaUpdateWriter.Close()

	go startKafkaConsumer(conf)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getHandler(w, r, conf)
		case "PURGE":
			purgeHandler(w, r, conf)
		default:
			directProxy(w, r, conf)
		}
	})

	log.Printf("Listening on %s...\n", conf.ListenAddr)
	log.Fatal(http.ListenAndServe(conf.ListenAddr, nil))

}

func setupS3Client(conf *config.Config) error {
	endpoint := conf.Storage.S3Endpoint               // Set the S3-like object storage endpoint
	accessKeyID := conf.Storage.S3AccessKeyID         // Set the S3-like object storage access key ID
	secretAccessKey := conf.Storage.S3SecretAccessKey // Set the S3-like object storage secret access key

	var err error
	s3Client, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
		Region: conf.Storage.S3Region,
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
