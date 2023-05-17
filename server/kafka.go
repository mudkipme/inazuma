package server

import (
	"context"
	"log"
	"time"

	"github.com/mudkipme/inazuma/config"
	"github.com/segmentio/kafka-go"
)

var (
	kafkaWriter       *kafka.Writer
	kafkaUpdateWriter *kafka.Writer
)

func setupKafkaProducer(conf *config.Config) {
	kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(conf.Queue.KafkaBrokers...),
		Topic:    conf.Queue.KafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}

	kafkaUpdateWriter = &kafka.Writer{
		Addr:     kafka.TCP(conf.Queue.KafkaBrokers...),
		Topic:    conf.Queue.KafkaUpdateTopic,
		Balancer: &kafka.LeastBytes{},
	}
}

func produceMessageWithKafkaGo(urlPath string) {
	err := kafkaUpdateWriter.WriteMessages(context.Background(),
		kafka.Message{Value: []byte(urlPath)},
	)

	if err != nil {
		log.Printf("Error producing kafka message: %v", err)
	}
}

func sendPurgeRequestToKafka(urlPath string) {
	err := kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(urlPath),
			Value: []byte(time.Now().UTC().Format(time.RFC3339)),
		},
	)

	if err != nil {
		log.Printf("Error sending PURGE request to Kafka: %s\n", err)
	} else {
		log.Printf("Sent PURGE request to Kafka: %s\n", urlPath)
	}
}

func startKafkaConsumer(conf *config.Config) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   conf.Queue.KafkaBrokers,
		Topic:     conf.Queue.KafkaUpdateTopic,
		Partition: 0,
		MinBytes:  10e3,
		MaxBytes:  10e6,
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading kafka message: %v", err)
			continue
		}

		// Handle PURGE request
		urlPath := string(m.Key)

		purgeRequestTimestamp, err := time.Parse(time.RFC3339, string(m.Value))
		if err != nil {
			log.Printf("Error parsing PURGE request timestamp: %s\n", err)
			continue
		}

		cacheUpdateTimestampStr, err := redisClient.Get(context.Background(), urlPath).Result()
		if err == nil {
			cacheUpdateTimestamp, err := time.Parse(time.RFC3339, cacheUpdateTimestampStr)
			if err == nil && cacheUpdateTimestamp.After(purgeRequestTimestamp) {
				log.Printf("Skipping PURGE request from Kafka: %s (cache updated after request)\n", urlPath)
				continue
			}
		}

		log.Printf("Received PURGE request from Kafka: %s\n", urlPath)
		updateCache(context.Background(), urlPath, conf)

		produceMessageWithKafkaGo(urlPath)
	}
}
