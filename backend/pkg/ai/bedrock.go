package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// BedrockClient wraps AWS Bedrock client
type BedrockClient struct {
	httpClient  *http.Client
	bearerToken string
	region      string
	model       string
	modelArn    string
	maxTokens   int
	temperature float64
}

// BedrockConfig holds AWS Bedrock configuration
type BedrockConfig struct {
	Region          string
	BearerToken     string // For Bedrock API Key authentication
	AccessKeyID     string // For IAM authentication (not used with bearer token)
	SecretAccessKey string // For IAM authentication (not used with bearer token)
	Model           string
	ModelArn        string // Optional: full ARN for inference profile
	MaxTokens       int
	Temperature     float64
}

// NewBedrockClient creates a new Bedrock client using API Key (Bearer Token)
func NewBedrockClient(config BedrockConfig) (*BedrockClient, error) {
	if config.BearerToken == "" {
		return nil, fmt.Errorf("bearer token is required for Bedrock API Key authentication")
	}

	return &BedrockClient{
		httpClient:  &http.Client{},
		bearerToken: config.BearerToken,
		region:      config.Region,
		model:       config.Model,
		modelArn:    config.ModelArn,
		maxTokens:   config.MaxTokens,
		temperature: config.Temperature,
	}, nil
}

// ReviewCode sends code diff to AWS Bedrock Claude for review
func (c *BedrockClient) ReviewCode(diff string, context string) (*ReviewResult, error) {
	prompt := buildReviewPrompt(diff, context)

	// Prepare request for Claude 3.5 Sonnet
	requestBody := map[string]interface{}{
		"anthropic_version": "bedrock-2023-05-31",
		"max_tokens":        c.maxTokens,
		"temperature":       c.temperature,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]string{
					{
						"type": "text",
						"text": prompt,
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine which model ID to use
	modelID := c.model
	if c.modelArn != "" {
		modelID = c.modelArn
	}

	// Build Bedrock endpoint URL
	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/invoke", c.region, modelID)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bedrock API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no response from Bedrock")
	}

	// Parse AI response into structured format
	reviewResult, err := parseReviewResponse(response.Content[0].Text)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return reviewResult, nil
}

// Enhanced Client that supports both Bedrock and other providers
type EnhancedClient struct {
	provider       Provider
	bedrockClient  *BedrockClient
	standardClient *Client
}

// NewEnhancedClient creates a client that supports Bedrock and standard providers
func NewEnhancedClient(provider Provider, apiKey string, bedrockConfig *BedrockConfig) (*EnhancedClient, error) {
	if provider == "bedrock" && bedrockConfig != nil {
		bedrockClient, err := NewBedrockClient(*bedrockConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Bedrock client: %w", err)
		}
		return &EnhancedClient{
			provider:      provider,
			bedrockClient: bedrockClient,
		}, nil
	}

	// Use standard client for other providers
	return &EnhancedClient{
		provider:       provider,
		standardClient: NewClient(provider, apiKey),
	}, nil
}

// ReviewCode routes to appropriate client
func (c *EnhancedClient) ReviewCode(diff string, context string) (*ReviewResult, error) {
	if c.provider == "bedrock" && c.bedrockClient != nil {
		return c.bedrockClient.ReviewCode(diff, context)
	}

	if c.standardClient != nil {
		return c.standardClient.ReviewCode(diff, context)
	}

	return nil, fmt.Errorf("no valid client configured")
}
