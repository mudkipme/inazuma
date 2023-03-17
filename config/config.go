package config

import (
	"github.com/spf13/viper"
)

type Configuration struct {
	ListenAddr          string
	UpstreamURL         string
	S3Endpoint          string
	S3Region            string
	S3AccessKeyID       string
	S3SecretAccessKey   string
	S3Bucket            string
	KafkaBrokers        []string
	KafkaConsumerGroup  string
	KafkaTopic          string
	KafkaUpdateTopic    string
	RedisAddr           string
	RedisPassword       string
	RedisDB             int
	CookieToBypassCache string
}

func LoadConfig() (*Configuration, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/inazuma/")

	viper.SetDefault("listenAddr", ":8080")
	viper.SetDefault("kafkaConsumerGroup", "inazuma")
	viper.SetDefault("kafkaTopic", "inazuma")
	viper.SetDefault("kafkaUpdateTopic", "inazuma-update")
	viper.SetDefault("cookieToBypass", "bypass-cookie")
	viper.SetDefault("redisDB", 0)

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	conf := &Configuration{
		ListenAddr:          viper.GetString("listenAddr"),
		UpstreamURL:         viper.GetString("upstreamURL"),
		S3Endpoint:          viper.GetString("s3.endpoint"),
		S3Region:            viper.GetString("s3.region"),
		S3AccessKeyID:       viper.GetString("s3.accessKeyID"),
		S3SecretAccessKey:   viper.GetString("s3.secretAccessKey"),
		S3Bucket:            viper.GetString("s3.bucket"),
		KafkaBrokers:        viper.GetStringSlice("kafka.brokers"),
		KafkaConsumerGroup:  viper.GetString("kafka.consumerGroup"),
		KafkaTopic:          viper.GetString("kafka.topic"),
		KafkaUpdateTopic:    viper.GetString("kafka.updateTopic"),
		RedisAddr:           viper.GetString("redis.addr"),
		RedisPassword:       viper.GetString("redis.password"),
		RedisDB:             viper.GetInt("redis.db"),
		CookieToBypassCache: viper.GetString("cookieToBypass"),
	}

	return conf, nil
}
