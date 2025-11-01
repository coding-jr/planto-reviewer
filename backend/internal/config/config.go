package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	Env             string
	DatabaseURL     string
	AIProvider      string
	AIAPIKey        string
	AIModel         string
	EncryptionKey   string
	APIKey          string
	PollingInterval int
	// AWS Bedrock config
	AWSRegion          string
	AWSBearerToken     string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	BedrockModel       string
	BedrockModelArn    string
	BedrockMaxTokens   int
	BedrockTemperature float64
}

func Load() *Config {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	pollingInterval, _ := strconv.Atoi(getEnv("POLLING_INTERVAL", "30"))

	cfg := &Config{
		Port:            getEnv("PORT", "3000"),
		Env:             getEnv("ENV", "development"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		AIProvider:      getEnv("AI_PROVIDER", "openai"),
		AIAPIKey:        getEnv("AI_API_KEY", ""),
		AIModel:         getEnv("AI_MODEL", "gpt-4-turbo-preview"),
		EncryptionKey:   getEnv("ENCRYPTION_KEY", ""),
		APIKey:          getEnv("API_KEY", ""),
		PollingInterval: pollingInterval,
	}

	// Validate required config
	if cfg.DatabaseURL == "" {
		log.Fatal("❌ DATABASE_URL is required")
	}
	if cfg.AIAPIKey == "" {
		log.Fatal("❌ AI_API_KEY is required")
	}

	log.Printf("✅ Config loaded (env: %s)\n", cfg.Env)
	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}