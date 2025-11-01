package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coding-jr/planto-reviewer/backend/internal/config"
	"github.com/coding-jr/planto-reviewer/backend/internal/database"
	"github.com/coding-jr/planto-reviewer/backend/internal/services"
	"github.com/coding-jr/planto-reviewer/backend/pkg/ai"
)

func main() {
	// Load configuration (tries settings.json first, then falls back to env)
	cfg, err := config.LoadWithEnvOverride()
	if err != nil {
		log.Fatalf("❌ Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL, cfg.Env == "development")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize AI client (supports Bedrock or standard providers)
	var aiClient ai.ReviewClient
	if cfg.AIProvider == "bedrock" {
		bedrockConfig := &ai.BedrockConfig{
			Region:          cfg.AWSRegion,
			BearerToken:     cfg.AWSBearerToken,
			AccessKeyID:     cfg.AWSAccessKeyID,
			SecretAccessKey: cfg.AWSSecretAccessKey,
			Model:           cfg.BedrockModel,
			ModelArn:        cfg.BedrockModelArn,
			MaxTokens:       cfg.BedrockMaxTokens,
			Temperature:     cfg.BedrockTemperature,
		}
		enhancedClient, err := ai.NewEnhancedClient(ai.Provider(cfg.AIProvider), "", bedrockConfig)
		if err != nil {
			log.Fatalf("❌ Failed to create Bedrock client: %v", err)
		}
		aiClient = enhancedClient

		// Log which auth method is being used
		if cfg.AWSBearerToken != "" {
			log.Printf("✅ Using AWS Bedrock with API Key authentication")
			log.Printf("ℹ️  Model: %s", cfg.BedrockModel)
			log.Printf("ℹ️  Region: %s", cfg.AWSRegion)
		} else {
			log.Printf("✅ Using AWS Bedrock with IAM authentication")
			log.Printf("ℹ️  Model: %s", cfg.BedrockModel)
		}
	} else {
		aiClient = ai.NewClient(ai.Provider(cfg.AIProvider), cfg.AIAPIKey)
		log.Printf("✅ Using %s provider", cfg.AIProvider)
	}

	// Initialize services
	githubService := services.NewGitHubService(db)
	reviewService := services.NewReviewService(db, aiClient)
	metricsService := services.NewMetricsService(db)

	log.Println("🚀 Background worker started")
	log.Printf("ℹ️  AI Provider: %s", cfg.AIProvider)
	log.Printf("ℹ️  Polling interval: %d seconds", cfg.PollingInterval)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Create ticker for polling
	ticker := time.NewTicker(time.Duration(cfg.PollingInterval) * time.Second)
	defer ticker.Stop()

	// Run initial cycle immediately
	runCycle(githubService, reviewService, metricsService)

	// Main loop
	for {
		select {
		case <-ticker.C:
			runCycle(githubService, reviewService, metricsService)
		case <-quit:
			log.Println("🛑 Shutting down worker...")
			return
		}
	}
}

func runCycle(githubService *services.GitHubService, reviewService *services.ReviewService, metricsService *services.MetricsService) {
	log.Println("⏰ Starting processing cycle...")

	// Step 1: Fetch new PRs from GitHub
	log.Println("📥 Fetching PRs from GitHub...")
	if err := githubService.FetchNewPRs(); err != nil {
		log.Printf("❌ Failed to fetch PRs: %v", err)
	}

	// Step 2: Process pending reviews (batch of 10)
	log.Println("🤖 Processing pending reviews...")
	if err := reviewService.ProcessPendingReviews(10); err != nil {
		log.Printf("❌ Failed to process reviews: %v", err)
	}

	// Optional: Recalculate all metrics (can be expensive, maybe run less frequently)
	// For now, metrics are updated per-developer/org after each review
	// Uncomment below to recalculate all metrics on each cycle:
	// log.Println("📊 Recalculating metrics...")
	// if err := metricsService.RecalculateAllMetrics(); err != nil {
	// 	log.Printf("❌ Failed to recalculate metrics: %v", err)
	// }

	log.Println("✅ Processing cycle completed")
	printSeparator()
}

func printSeparator() {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}
