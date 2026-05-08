// api/orchestration/gemini.go

package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GeminiClient implements AIProvider for Google Gemini
type GeminiClient struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * 1e9},
	}
}

// geminiRequest represents a Gemini API request
type geminiRequest struct {
	Contents        []geminiContent `json:"contents"`
	GenerationConfig struct {
		MaxOutputTokens int     `json:"maxOutputTokens"`
		Temperature     float64 `json:"temperature"`
		TopP            float64 `json:"topP"`
	} `json:"generationConfig"`
	SystemInstruction *struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"systemInstruction,omitempty"`
}

// geminiContent represents content in Gemini's format
type geminiContent struct {
	Role  string `json:"role"`
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
}

// geminiResponse represents Gemini's API response
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// GenerateResponse calls Gemini API to generate a response
func (gc *GeminiClient) GenerateResponse(ctx context.Context, prompt string, conversationHistory []string) (string, int, error) {
	// Build message contents
	contents := buildGeminiContents(prompt, conversationHistory)

	// Create request
	payload := geminiRequest{
		Contents: contents,
	}

	payload.GenerationConfig.MaxOutputTokens = 1024
	payload.GenerationConfig.Temperature = 0.7
	payload.GenerationConfig.TopP = 0.9

	// Add system instruction
	payload.SystemInstruction = &struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}{
		Parts: []struct {
			Text string `json:"text"`
		}{
			{Text: getSystemPrompt()},
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL with API key
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", gc.model, gc.apiKey)

	// Send request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := gc.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("gemini api error: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", 0, fmt.Errorf("no content in gemini response")
	}

	responseText := geminiResp.Candidates[0].Content.Parts[0].Text
	totalTokens := geminiResp.UsageMetadata.TotalTokenCount

	return responseText, totalTokens, nil
}

// GetModel returns the model name
func (gc *GeminiClient) GetModel() string {
	return gc.model
}

// GetProvider returns the provider name
func (gc *GeminiClient) GetProvider() string {
	return "gemini"
}

// buildGeminiContents constructs content for Gemini API
func buildGeminiContents(prompt string, conversationHistory []string) []geminiContent {
	contents := []geminiContent{}

	// Add conversation history (limit to last 10 messages)
	start := len(conversationHistory) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(conversationHistory); i++ {
		role := "user"
		if i%2 == 1 {
			role = "model"
		}

		contents = append(contents, geminiContent{
			Role: role,
			Parts: []struct {
				Text string `json:"text"`
			}{
				{Text: conversationHistory[i]},
			},
		})
	}

	// Add current prompt
	contents = append(contents, geminiContent{
		Role: "user",
		Parts: []struct {
			Text string `json:"text"`
		}{
			{Text: prompt},
		},
	})

	return contents
}
