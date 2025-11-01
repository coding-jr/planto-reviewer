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

	// Load AWS Bedrock config if using bedrock provider
	if cfg.AIProvider == "bedrock" {
		cfg.AWSRegion = getEnv("AWS_REGION", "us-east-1")
		cfg.AWSBearerToken = getEnv("AWS_BEARER_TOKEN_BEDROCK", "")
		cfg.AWSAccessKeyID = getEnv("AWS_ACCESS_KEY_ID", "")
		cfg.AWSSecretAccessKey = getEnv("AWS_SECRET_ACCESS_KEY", "")
		cfg.BedrockModel = getEnv("BEDROCK_MODEL", "anthropic.claude-3-5-sonnet-20241022-v2:0")
		cfg.BedrockModelArn = getEnv("BEDROCK_MODEL_ARN", "")
		if maxTokens, err := strconv.Atoi(getEnv("BEDROCK_MAX_TOKENS", "4096")); err == nil {
			cfg.BedrockMaxTokens = maxTokens
		}
		if temp, err := strconv.ParseFloat(getEnv("BEDROCK_TEMPERATURE", "0.3"), 64); err == nil {
			cfg.BedrockTemperature = temp
		}
	}

	// Validate required config
	if cfg.DatabaseURL == "" {
		log.Fatal("❌ DATABASE_URL is required")
	}
	// Only require AI_API_KEY for non-Bedrock providers
	if cfg.AIProvider != "bedrock" && cfg.AIAPIKey == "" {
		log.Fatal("❌ AI_API_KEY is required")
	}
	// For Bedrock, require either bearer token or IAM credentials
	if cfg.AIProvider == "bedrock" {
		if cfg.AWSBearerToken == "" && (cfg.AWSAccessKeyID == "" || cfg.AWSSecretAccessKey == "") {
			log.Fatal("❌ AWS_BEARER_TOKEN_BEDROCK or AWS credentials (AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY) are required for Bedrock")
		}
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