package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CompletionRequest struct {
	SystemPrompt string
	UserPrompt   string
}

type Client interface {
	Complete(ctx context.Context, req CompletionRequest) (string, error)
}

type OpenAIClient struct {
	apiKey     string
	httpClient *http.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

type openAIRequest struct {
	Model          string           `json:"model"`
	Messages       []openAIMessage  `json:"messages"`
	ResponseFormat openAIRespFormat `json:"response_format"`
	Temperature    float64          `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRespFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *OpenAIClient) Complete(ctx context.Context, req CompletionRequest) (string, error) {
	body := openAIRequest{
		Model: "gpt-4o-mini",
		Messages: []openAIMessage{
			{Role: "system", Content: req.SystemPrompt},
			{Role: "user", Content: req.UserPrompt},
		},
		ResponseFormat: openAIRespFormat{Type: "json_object"},
		Temperature:    1,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var oaiResp openAIResponse
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if oaiResp.Error != nil {
		return "", fmt.Errorf("openai error: %s", oaiResp.Error.Message)
	}

	if len(oaiResp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return oaiResp.Choices[0].Message.Content, nil
}
