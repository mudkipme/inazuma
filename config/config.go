package config

import (
	"strings"

	"github.com/spf13/viper"
)

type CachePattern struct {
	Path         string
	ReplaceRules []struct {
		Match        int
		Replacements []struct {
			Replacement     string
			LanguageMatches []string
			Default         bool
		}
	}
	Purge bool
}

type Config struct {
	ListenAddr          string
	UpstreamURL         string
	CookieToBypassCache []string

	Storage struct {
		S3Endpoint        string
		S3Region          string
		S3AccessKeyID     string
		S3SecretAccessKey string
		S3Bucket          string
	}

	Queue struct {
		KafkaBrokers       []string
		KafkaConsumerGroup string
		KafkaTopic         string
		KafkaUpdateTopic   string
	}

	Redis struct {
		Addr     string
		Password string
		DB       int
	}

	CachePatterns []CachePattern
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/inazuma/")
	viper.AddConfigPath("$XDG_CONFIG_HOME/inazuma")
	viper.AddConfigPath("$HOME/.config/inazuma")
	viper.AddConfigPath(".")

	viper.SetDefault("ListenAddr", ":8080")
	viper.SetDefault("Queue.KafkaConsumerGroup", "inazuma")
	viper.SetDefault("Queue.KafkaTopic", "inazuma")
	viper.SetDefault("Queue.KafkaUpdateTopic", "inazuma-update")
	viper.SetDefault("Redis.DB", 0)

	viper.SetEnvPrefix("inazuma")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var conf Config
	if err = viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
