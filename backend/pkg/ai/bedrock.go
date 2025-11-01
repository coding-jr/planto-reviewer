package ai

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrockruntime"
)

// BedrockClient wraps AWS Bedrock client
type BedrockClient struct {
	client      *bedrockruntime.BedrockRuntime
	model       string
	maxTokens   int
	temperature float64
}

// BedrockConfig holds AWS Bedrock configuration
type BedrockConfig struct {
	Region          string
	BearerToken     string // For Bedrock API Key authentication
	AccessKeyID     string // For IAM authentication
	SecretAccessKey string // For IAM authentication
	Model           string
	ModelArn        string // Optional: full ARN for inference profile
	MaxTokens       int
	Temperature     float64
}

// NewBedrockClient creates a new Bedrock client
func NewBedrockClient(config BedrockConfig) (*BedrockClient, error) {
	var sess *session.Session
	var err error

	// Use bearer token authentication if provided, otherwise use IAM
	if config.BearerToken != "" {
		// Bedrock API Key authentication
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(config.Region),
			Credentials: credentials.NewStaticCredentials("", "", config.BearerToken),
		})
	} else {
		// IAM authentication
		sess, err = session.NewSession(&aws.Config{
			Region:      aws.String(config.Region),
			Credentials: credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create Bedrock Runtime client
	svc := bedrockruntime.New(sess)

	// Use ModelArn if provided, otherwise use Model
	modelID := config.Model
	if config.ModelArn != "" {
		modelID = config.ModelArn
	}

	return &BedrockClient{
		client:      svc,
		model:       modelID,
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

	// Invoke model
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.model),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        jsonBody,
	}

	result, err := c.client.InvokeModel(input)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke model: %w", err)
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

	if err := json.Unmarshal(result.Body, &response); err != nil {
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
	provider      Provider
	bedrockClient *BedrockClient
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
