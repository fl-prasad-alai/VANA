// api/orchestration/groq.go

package orchestration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GroqClient implements AIProvider for Groq
type GroqClient struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGroqClient creates a new Groq API client
func NewGroqClient(apiKey, model string) *GroqClient {
	return &GroqClient{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 30 * 1e9}, // 30 seconds
	}
}

// groqRequest represents a Groq API request
type groqRequest struct {
	Messages            []groqMessage `json:"messages"`
	Model              string        `json:"model"`
	MaxTokens          int           `json:"max_tokens"`
	Temperature        float64       `json:"temperature"`
	TopP               float64       `json:"top_p"`
	Stream             bool          `json:"stream"`
}

// groqMessage represents a message in Groq's format
type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// groqResponse represents Groq's API response
type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateResponse calls Groq API to generate a response
func (gc *GroqClient) GenerateResponse(ctx context.Context, prompt string, conversationHistory []string) (string, int, error) {
	// Build message history
	messages := buildMessageHistory(prompt, conversationHistory)

	// Create request
	payload := groqRequest{
		Messages:     messages,
		Model:        gc.model,
		MaxTokens:    1024,
		Temperature:  0.7,
		TopP:         0.9,
		Stream:       false,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gc.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := gc.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("groq api error: %s (status: %d)", string(body), resp.StatusCode)
	}

	// Parse response
	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", 0, fmt.Errorf("no choices in response")
	}

	return groqResp.Choices[0].Message.Content, groqResp.Usage.TotalTokens, nil
}

// GetModel returns the model name
func (gc *GroqClient) GetModel() string {
	return gc.model
}

// GetProvider returns the provider name
func (gc *GroqClient) GetProvider() string {
	return "groq"
}

// buildMessageHistory constructs message history for the prompt
func buildMessageHistory(prompt string, conversationHistory []string) []groqMessage {
	messages := []groqMessage{
		{
			Role:    "system",
			Content: getSystemPrompt(),
		},
	}

	// Add conversation history (limit to last 10 messages for context)
	start := len(conversationHistory) - 10
	if start < 0 {
		start = 0
	}

	for i := start; i < len(conversationHistory); i++ {
		// Alternate between user and assistant
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		messages = append(messages, groqMessage{
			Role:    role,
			Content: conversationHistory[i],
		})
	}

	// Add current prompt
	messages = append(messages, groqMessage{
		Role:    "user",
		Content: prompt,
	})

	return messages
}

// getSystemPrompt returns the system prompt for the AI
func getSystemPrompt() string {
	return `You are Emerald Moss, a compassionate and empathetic mental health companion. 

Your purpose is to provide supportive, evidence-based guidance for mental wellness. Follow these principles:

1. EMPATHY & LISTENING: Always validate the user's feelings and listen deeply.
2. EVIDENCE-BASED: Use only scientifically-backed approaches.
3. SAFETY FIRST: Never provide medical advice. Always recommend professional help for serious concerns.
4. NON-JUDGMENTAL: Create a safe space free from judgment.
5. BIOPHILIC DESIGN: Remind users of nature's calming effects when appropriate.
6. CRISIS AWARE: If the user mentions self-harm or suicidal ideation, immediately:
   - Validate their feelings
   - Provide crisis resources (988 US Lifeline)
   - Strongly encourage professional help

When discussing mental health topics:
- Use accessible, non-clinical language
- Provide coping strategies and mindfulness techniques
- Suggest professional resources
- Never diagnose or prescribe
- Be honest about your limitations as an AI

Remember: You are here to support, not replace professional mental health care.`
}
