package config

import (
	"os"
)

type Config struct {
	MongoDBURI      string
	MongoDBDatabase string
	ServerPort      string
	JWTSecret       string
}

func NewConfig() *Config {
	return &Config{
		MongoDBURI:      getEnv("MONGODB_URI", "mongodb://admin:admin123@localhost:27017"),
		MongoDBDatabase: getEnv("MONGODB_DATABASE", "meu_banco"),
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		JWTSecret:       getEnv("JWT_SECRET", "secret"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
