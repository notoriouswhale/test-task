package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTP          HTTPConfig
	DB            DBConfig
	MessageBroker MessageBrokerConfig
}

type HTTPConfig struct {
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type MessageBrokerConfig struct {
	Endpoint string
	Topic    string
	ClientID string
}

func Load() *Config {
	// для development
	_ = godotenv.Load()

	return &Config{
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8081"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "root"),
			DBName:   getEnv("DB_NAME", "products"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		MessageBroker: MessageBrokerConfig{
			Endpoint: getEnv("MESSAGE_BROKER_ENDPOINT", "localhost:9094"),
			Topic:    getEnv("MESSAGE_BROKER_TOPIC", "product-events"),
			ClientID: getEnv("MESSAGE_BROKER_CLIENT_ID", "product-service"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
