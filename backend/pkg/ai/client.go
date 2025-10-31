package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Provider string

const (
	ProviderOpenAI    Provider = "openai"
	ProviderAnthropic Provider = "anthropic"
	ProviderGoogle    Provider = "google"
)

type Client struct {
	provider Provider
	apiKey   string
	model    string
	client   *http.Client
}

type ReviewResult struct {
	Summary          string  `json:"summary"`
	CodeQualityScore float64 `json:"code_quality_score"`
	Issues           []Issue `json:"issues"`
}

type Issue struct {
	Type        string  `json:"type"`         // null_check, logic_error, scalability, etc
	Severity    string  `json:"severity"`     // critical, high, medium, low
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Suggestion  *string `json:"suggestion,omitempty"`
	CodeSnippet *string `json:"code_snippet,omitempty"`
	LineNumber  *int    `json:"line_number,omitempty"`
	FileName    *string `json:"file_name,omitempty"`
}

func NewClient(provider Provider, apiKey string) *Client {
	model := getDefaultModel(provider)
	return &Client{
		provider: provider,
		apiKey:   apiKey,
		model:    model,
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func getDefaultModel(provider Provider) string {
	switch provider {
	case ProviderOpenAI:
		return "gpt-4-turbo-preview"
	case ProviderAnthropic:
		return "claude-3-5-sonnet-20241022"
	case ProviderGoogle:
		return "gemini-pro"
	default:
		return "gpt-4-turbo-preview"
	}
}

// ReviewCode sends code diff to AI for review
func (c *Client) ReviewCode(diff string, context string) (*ReviewResult, error) {
	prompt := buildReviewPrompt(diff, context)

	var response string
	var err error

	switch c.provider {
	case ProviderOpenAI:
		response, err = c.callOpenAI(prompt)
	case ProviderAnthropic:
		response, err = c.callAnthropic(prompt)
	case ProviderGoogle:
		response, err = c.callGoogle(prompt)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", c.provider)
	}

	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// Parse AI response into structured format
	result, err := parseReviewResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return result, nil
}

func buildReviewPrompt(diff string, context string) string {
	return fmt.Sprintf(`You are a senior software engineer performing a code review. Analyze the following code changes and provide a detailed review.

%s

Code Changes (Git Diff):
%s

Please analyze the code and respond with a JSON object in this exact format:
{
  "summary": "Brief summary of the changes and overall quality",
  "code_quality_score": 85.5,
  "issues": [
    {
      "type": "null_check",
      "severity": "high",
      "title": "Missing null check",
      "description": "Variable 'user' could be null before accessing properties",
      "suggestion": "Add null check: if (user === null) return;",
      "code_snippet": "user.name",
      "line_number": 42,
      "file_name": "src/user.ts"
    }
  ]
}

Issue types you should look for:
- null_check: Missing null/undefined checks
- logic_error: Flawed logic, incorrect conditions, edge cases not handled
- scalability: Performance issues, N+1 queries, inefficient algorithms
- security: SQL injection, XSS, exposed secrets, authentication issues
- performance: Inefficient code, unnecessary loops, memory issues
- race_condition: Concurrency issues, race conditions
- memory_leak: Resource leaks, unclosed connections
- error_handling: Missing try-catch, unhandled errors

Severity levels: critical, high, medium, low

The code_quality_score should be 0-100, where:
- 90-100: Excellent code, minimal issues
- 70-89: Good code, minor issues
- 50-69: Average code, several issues
- 30-49: Below average, many issues
- 0-29: Poor code, critical issues

Respond ONLY with the JSON object, no additional text.`, context, diff)
}

// callOpenAI calls OpenAI API
func (c *Client) callOpenAI(prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3,
		"max_tokens":  4000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return result.Choices[0].Message.Content, nil
}

// callAnthropic calls Anthropic Claude API
func (c *Client) callAnthropic(prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	reqBody := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 4000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Anthropic API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no response from Anthropic")
	}

	return result.Content[0].Text, nil
}

// callGoogle calls Google Gemini API
func (c *Client) callGoogle(prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/%s:generateContent?key=%s", c.model, c.apiKey)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Google API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Google")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// parseReviewResponse parses AI response into ReviewResult
func parseReviewResponse(response string) (*ReviewResult, error) {
	// Try to extract JSON from response (in case AI added extra text)
	start := -1
	end := -1
	for i, c := range response {
		if c == '{' && start == -1 {
			start = i
		}
		if c == '}' {
			end = i + 1
		}
	}

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := response[start:end]

	var result ReviewResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate result
	if result.CodeQualityScore < 0 || result.CodeQualityScore > 100 {
		result.CodeQualityScore = 50 // Default to average
	}

	return &result, nil
}
