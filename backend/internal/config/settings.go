package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
// Environment variables take priority over settings.json
func LoadWithEnvOverride() (*Config, error) {
	// First try to load from settings.json
	settings, err := LoadSettings("")
	var cfg *Config

	if err != nil {
		fmt.Printf("⚠️  Could not load settings.json: %v\n", err)
		fmt.Println("ℹ️  Using environment variables only")
		cfg = &Config{}
	} else {
		fmt.Println("✅ Loaded base configuration from settings.json")

		// Create config from settings
		cfg = &Config{
			Port:            fmt.Sprintf("%d", settings.API.Port),
			Env:             getEnv("ENV", "production"),
			DatabaseURL:     settings.Database.URL,
			AIProvider:      "bedrock", // Default to bedrock when using settings.json
			PollingInterval: settings.Worker.PollingIntervalSeconds,
			APIKey:          settings.API.APIKey,
		}

		// AWS/Bedrock config from settings.json
		cfg.AWSRegion = settings.AWS.Region
		cfg.AWSBearerToken = settings.AWS.BearerToken
		cfg.AWSAccessKeyID = settings.AWS.AccessKeyID
		cfg.AWSSecretAccessKey = settings.AWS.SecretAccessKey
		cfg.BedrockModel = settings.Bedrock.Model
		cfg.BedrockModelArn = settings.Bedrock.ModelArn
		cfg.BedrockMaxTokens = settings.Bedrock.MaxTokens
		cfg.BedrockTemperature = settings.Bedrock.Temperature
	}

	// Environment variables override everything (priority order)
	if envProvider := os.Getenv("AI_PROVIDER"); envProvider != "" {
		cfg.AIProvider = envProvider
	}

	if envAPIKey := os.Getenv("AI_API_KEY"); envAPIKey != "" {
		cfg.AIAPIKey = envAPIKey
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

	// AWS Bedrock environment variables (priority: env vars > settings.json)
	if envRegion := os.Getenv("AWS_REGION"); envRegion != "" {
		cfg.AWSRegion = envRegion
	}

	// Support both AWS_BEARER_TOKEN_BEDROCK (VSCode style) and AWS_BEARER_TOKEN
	if envBearerToken := os.Getenv("AWS_BEARER_TOKEN_BEDROCK"); envBearerToken != "" {
		cfg.AWSBearerToken = envBearerToken
		fmt.Println("ℹ️  Using AWS_BEARER_TOKEN_BEDROCK from environment")
	} else if envBearerToken := os.Getenv("AWS_BEARER_TOKEN"); envBearerToken != "" {
		cfg.AWSBearerToken = envBearerToken
		fmt.Println("ℹ️  Using AWS_BEARER_TOKEN from environment")
	}

	if envAccessKey := os.Getenv("AWS_ACCESS_KEY_ID"); envAccessKey != "" {
		cfg.AWSAccessKeyID = envAccessKey
	}

	if envSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY"); envSecretKey != "" {
		cfg.AWSSecretAccessKey = envSecretKey
	}

	if envModel := os.Getenv("BEDROCK_MODEL"); envModel != "" {
		cfg.BedrockModel = envModel
	}

	if envModelArn := os.Getenv("BEDROCK_MODEL_ARN"); envModelArn != "" {
		cfg.BedrockModelArn = envModelArn
	}

	if envMaxTokens := os.Getenv("BEDROCK_MAX_TOKENS"); envMaxTokens != "" {
		if maxTokens, err := strconv.Atoi(envMaxTokens); err == nil {
			cfg.BedrockMaxTokens = maxTokens
		}
	}

	if envTemp := os.Getenv("BEDROCK_TEMPERATURE"); envTemp != "" {
		if temp, err := strconv.ParseFloat(envTemp, 64); err == nil {
			cfg.BedrockTemperature = temp
		}
	}

	if envPolling := os.Getenv("POLLING_INTERVAL"); envPolling != "" {
		if interval, err := strconv.Atoi(envPolling); err == nil {
			cfg.PollingInterval = interval
		}
	}

	// Validate required configuration
	if cfg.DatabaseURL == "" {
		fmt.Println("❌ DATABASE_URL is required (set via environment variable or settings.json)")
		os.Exit(1)
	}

	if cfg.AIProvider == "bedrock" {
		if cfg.AWSRegion == "" || cfg.BedrockModel == "" {
			fmt.Println("❌ AWS configuration is required for Bedrock provider")
			fmt.Println("   Required: AWS_REGION and BEDROCK_MODEL")
			fmt.Println("   Auth: AWS_BEARER_TOKEN_BEDROCK or AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY")
			os.Exit(1)
		}

		// Check authentication method
		hasAuth := cfg.AWSBearerToken != "" || (cfg.AWSAccessKeyID != "" && cfg.AWSSecretAccessKey != "")
		if !hasAuth {
			fmt.Println("❌ AWS authentication is required for Bedrock")
			fmt.Println("   Option 1: Set AWS_BEARER_TOKEN_BEDROCK (recommended)")
			fmt.Println("   Option 2: Set AWS_ACCESS_KEY_ID + AWS_SECRET_ACCESS_KEY")
			os.Exit(1)
		}

		if cfg.AWSBearerToken != "" {
			fmt.Println("✅ Using AWS Bedrock with Bearer Token authentication")
		} else {
			fmt.Println("✅ Using AWS Bedrock with IAM credentials authentication")
		}
		fmt.Printf("ℹ️  Model: %s\n", cfg.BedrockModel)
		fmt.Printf("ℹ️  Region: %s\n", cfg.AWSRegion)
	}

	return cfg, nil
}
