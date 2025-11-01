package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Settings represents the settings.json structure
type Settings struct {
	AWS struct {
		Region          string `json:"region"`
		BearerToken     string `json:"bearerToken"`
		AccessKeyID     string `json:"accessKeyId"`
		SecretAccessKey string `json:"secretAccessKey"`
	} `json:"aws"`
	Bedrock struct {
		Model       string  `json:"model"`
		ModelArn    string  `json:"modelArn"`
		MaxTokens   int     `json:"maxTokens"`
		Temperature float64 `json:"temperature"`
	} `json:"bedrock"`
	API struct {
		Port   int    `json:"port"`
		APIKey string `json:"apiKey"`
	} `json:"api"`
	Database struct {
		URL string `json:"url"`
	} `json:"database"`
	Worker struct {
		PollingIntervalSeconds int `json:"pollingIntervalSeconds"`
	} `json:"worker"`
}

// LoadSettings loads configuration from settings.json
func LoadSettings(settingsPath string) (*Settings, error) {
	// If no path provided, look in parent directory
	if settingsPath == "" {
		// Try to find settings.json in project root
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}

		// Go up directories to find settings.json
		for i := 0; i < 5; i++ {
			testPath := filepath.Join(cwd, "settings.json")
			if _, err := os.Stat(testPath); err == nil {
				settingsPath = testPath
				break
			}
			cwd = filepath.Dir(cwd)
		}

		if settingsPath == "" {
			return nil, fmt.Errorf("settings.json not found")
		}
	}

	// Read file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings.json: %w", err)
	}

	// Parse JSON
	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings.json: %w", err)
	}

	return &settings, nil
}

// LoadWithEnvOverride loads settings.json and allows env variables to override
func LoadWithEnvOverride() (*Config, error) {
	// First try to load from settings.json
	settings, err := LoadSettings("")
	if err != nil {
		fmt.Printf("⚠️  Could not load settings.json: %v\n", err)
		fmt.Println("ℹ️  Falling back to environment variables")
		return Load(), nil
	}

	fmt.Println("✅ Loaded configuration from settings.json")

	// Create config from settings
	cfg := &Config{
		Port:            fmt.Sprintf("%d", settings.API.Port),
		Env:             getEnv("ENV", "production"),
		DatabaseURL:     settings.Database.URL,
		AIProvider:      "bedrock", // Default to bedrock when using settings.json
		PollingInterval: settings.Worker.PollingIntervalSeconds,
		APIKey:          settings.API.APIKey,
	}

	// AWS/Bedrock config
	cfg.AWSRegion = settings.AWS.Region
	cfg.AWSBearerToken = settings.AWS.BearerToken
	cfg.AWSAccessKeyID = settings.AWS.AccessKeyID
	cfg.AWSSecretAccessKey = settings.AWS.SecretAccessKey
	cfg.BedrockModel = settings.Bedrock.Model
	cfg.BedrockModelArn = settings.Bedrock.ModelArn
	cfg.BedrockMaxTokens = settings.Bedrock.MaxTokens
	cfg.BedrockTemperature = settings.Bedrock.Temperature

	// Allow env variables to override
	if envProvider := os.Getenv("AI_PROVIDER"); envProvider != "" {
		cfg.AIProvider = envProvider
		cfg.AIAPIKey = os.Getenv("AI_API_KEY")
	}

	if envDB := os.Getenv("DATABASE_URL"); envDB != "" {
		cfg.DatabaseURL = envDB
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.Port = envPort
	}

	if envAPIKey := os.Getenv("API_KEY"); envAPIKey != "" {
		cfg.APIKey = envAPIKey
	}

	// Validate
	if cfg.DatabaseURL == "" {
		fmt.Println("❌ DATABASE_URL is required")
		os.Exit(1)
	}

	if cfg.AIProvider == "bedrock" {
		if cfg.AWSRegion == "" || cfg.BedrockModel == "" {
			fmt.Println("❌ AWS configuration is required for Bedrock provider")
			os.Exit(1)
		}
	}

	return cfg, nil
}
