package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env                     string
	PostgresHost            string
	PostgresPort            string
	PostgresUser            string
	PostgresPassword        string
	PostgresDB              string
	PostgresSSLMode         string
	RabbitMQURL             string
	RabbitMQQueue           string
	RabbitMQDeadLetterQueue string
	AzureStorageName        string
	AzureStorageKey         string
	AzureStorageEndpoint    string
	AzureBlobRawContainer   string
	AzureBlobHLSContainer   string
	FfmpegPath              string
	TempDir                 string
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Environment variable %s is required but missing", key))
	}
	return val
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func Load() Config {
	return Config{
		Env:                     getEnvOrDefault("ENV", "development"),
		PostgresHost:            getEnv("POSTGRES_HOST"),
		PostgresPort:            getEnv("POSTGRES_PORT"),
		PostgresUser:            getEnv("POSTGRES_USER"),
		PostgresPassword:        getEnv("POSTGRES_PASSWORD"),
		PostgresDB:              getEnv("POSTGRES_DB"),
		PostgresSSLMode:         getEnv("POSTGRES_SSL_MODE"),
		RabbitMQURL:             getEnv("RABBITMQ_URL"),
		RabbitMQQueue:           getEnv("RABBITMQ_QUEUE"),
		RabbitMQDeadLetterQueue: getEnv("RABBITMQ_DEAD_LETTER_QUEUE"),
		AzureStorageName:        getEnv("AZURE_STORAGE_ACCOUNT_NAME"),
		AzureStorageKey:         getEnv("AZURE_STORAGE_ACCOUNT_KEY"),
		AzureStorageEndpoint:    getEnvOrDefault("AZURE_STORAGE_ENDPOINT", ""),
		AzureBlobRawContainer:   getEnv("AZURE_BLOB_RAW_CONTAINER"),
		AzureBlobHLSContainer:   getEnv("AZURE_BLOB_HLS_CONTAINER"),
		FfmpegPath:              getEnv("FFMPEG_PATH"),
		TempDir:                 getEnv("TEMP_DIR"),
	}
}
