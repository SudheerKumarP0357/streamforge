package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port                   string
	Env                    string
	PostgresHost           string
	PostgresPort           string
	PostgresUser           string
	PostgresPassword       string
	PostgresDB             string
	PostgresSSLMode        string
	CosmosConnectionString string
	CosmosDatabase         string
	RedisURL               string
	RedisPassword          string
	RabbitMQURL            string
	AzureStorageName       string
	AzureStorageKey        string
	AzureStorageEndpoint   string
	AzureBlobRawContainer  string
	AzureBlobHLSContainer  string
	JWTSecret              string
	JWTExpiryHours         string
	RefreshTokenExpiryDays string
	CORSAllowedOrigin      string
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
		Port:                   getEnvOrDefault("PORT", "8080"),
		Env:                    getEnvOrDefault("ENV", "development"),
		PostgresHost:           getEnv("POSTGRES_HOST"),
		PostgresPort:           getEnv("POSTGRES_PORT"),
		PostgresUser:           getEnv("POSTGRES_USER"),
		PostgresPassword:       getEnv("POSTGRES_PASSWORD"),
		PostgresDB:             getEnv("POSTGRES_DB"),
		PostgresSSLMode:        getEnv("POSTGRES_SSL_MODE"),
		CosmosConnectionString: getEnv("COSMOS_CONNECTION_STRING"),
		CosmosDatabase:         getEnv("COSMOS_DATABASE"),
		RedisURL:               getEnv("REDIS_URL"),
		RedisPassword:          getEnvOrDefault("REDIS_PASSWORD", ""), // Redis password might be empty locally
		RabbitMQURL:            getEnv("RABBITMQ_URL"),
		AzureStorageName:       getEnv("AZURE_STORAGE_ACCOUNT_NAME"),
		AzureStorageKey:        getEnv("AZURE_STORAGE_ACCOUNT_KEY"),
		AzureStorageEndpoint:   getEnvOrDefault("AZURE_STORAGE_ENDPOINT", ""),
		AzureBlobRawContainer:  getEnv("AZURE_BLOB_RAW_CONTAINER"),
		AzureBlobHLSContainer:  getEnv("AZURE_BLOB_HLS_CONTAINER"),
		JWTSecret:              getEnv("JWT_SECRET"),
		JWTExpiryHours:         getEnv("JWT_EXPIRY_HOURS"),
		RefreshTokenExpiryDays: getEnv("REFRESH_TOKEN_EXPIRY_DAYS"),
		CORSAllowedOrigin:      getEnv("CORS_ALLOWED_ORIGIN"),
	}
}
