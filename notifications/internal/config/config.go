package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MessageBroker MessageBrokerConfig
}

type MessageBrokerConfig struct {
	Endpoint string
	Topic    string
	GroupID  string
}

func Load() *Config {
	// для development
	_ = godotenv.Load()

	return &Config{
		MessageBroker: MessageBrokerConfig{
			Endpoint: getEnv("MESSAGE_BROKER_ENDPOINT", "localhost:9094"),
			Topic:    getEnv("MESSAGE_BROKER_TOPIC", "product-events"),
			GroupID:  getEnv("CONSUMER_GROUP_ID", "notifications-group"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
