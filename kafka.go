package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"
)

var (
	kafkaProducer sarama.SyncProducer

	purgeTopic = os.Getenv("KAFKA_TOPIC")
)

func setupKafkaProducer() error {
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKER")}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	var err error
	kafkaProducer, err = sarama.NewSyncProducer(kafkaBrokers, config)
	return err
}

func sendPurgeRequestToKafka(urlPath string) {
	message := &sarama.ProducerMessage{
		Topic: purgeTopic,
		Key:   sarama.StringEncoder(urlPath),
		Value: sarama.StringEncoder(time.Now().UTC().Format(time.RFC3339)),
	}
	_, _, err := kafkaProducer.SendMessage(message)
	if err != nil {
		log.Printf("Error sending PURGE request to Kafka: %s\n", err)
	} else {
		log.Printf("Sent PURGE request to Kafka: %s\n", urlPath)
	}
}

func startKafkaConsumer() {
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKER")}
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(kafkaBrokers, config)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(purgeTopic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatal(err)
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			urlPath := string(msg.Key)
			purgeRequestTimestamp, err := time.Parse(time.RFC3339, string(msg.Value))
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
			updateCache(urlPath)
		case err := <-partitionConsumer.Errors():
			log.Printf("Kafka consumer error: %s\n", err)
		}
	}
}
