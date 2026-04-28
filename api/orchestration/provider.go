// api/orchestration/provider.go

package orchestration

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// AIProvider defines the interface for AI providers
type AIProvider interface {
	GenerateResponse(ctx context.Context, prompt string, conversationHistory []string) (string, int, error)
	GetModel() string
	GetProvider() string
}

// MultiProviderBalancer handles load balancing between Groq (Tier 1) and Gemini (Tier 2)
type MultiProviderBalancer struct {
	groqProvider   AIProvider
	geminiProvider AIProvider
	groqRPMUsed    int
	geminiRPMUsed  int
	groqRPMLimit   int
	geminiRPMLimit int
	lastResetTime  time.Time
}

// NewMultiProviderBalancer creates a new balancer
func NewMultiProviderBalancer(groq, gemini AIProvider, groqLimit, geminiLimit int) *MultiProviderBalancer {
	return &MultiProviderBalancer{
		groqProvider:   groq,
		geminiProvider: gemini,
		groqRPMUsed:    0,
		geminiRPMUsed:  0,
		groqRPMLimit:   groqLimit,
		geminiRPMLimit: geminiLimit,
		lastResetTime:  time.Now(),
	}
}

// HandleChat routes a message to the appropriate AI provider
func (b *MultiProviderBalancer) HandleChat(
	ctx context.Context,
	message string,
	conversationHistory []string,
	requireDeepLogic bool,
) (string, string, int, error) {
	// Reset RPM counters if minute has passed
	b.resetRPMIfNeeded()

	// If deep clinical synthesis or RAG is needed, route to Gemini (Tier 2)
	if requireDeepLogic || b.groqRPMUsed >= b.groqRPMLimit {
		return b.callGemini(ctx, message, conversationHistory)
	}

	// Try Groq first (Tier 1: Speed)
	ctxSpeed, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
	defer cancel()

	response, tokens, err := b.groqProvider.GenerateResponse(ctxSpeed, message, conversationHistory)
	if err == nil {
		b.groqRPMUsed++
		return response, "groq", tokens, nil
	}

	// Fallback to Gemini on Groq timeout or error
	fmt.Printf("Groq failed (%v), falling back to Gemini\n", err)
	return b.callGemini(ctx, message, conversationHistory)
}

// callGemini calls Gemini provider with rate limiting
func (b *MultiProviderBalancer) callGemini(ctx context.Context, message string, conversationHistory []string) (string, string, int, error) {
	if b.geminiRPMUsed >= b.geminiRPMLimit {
		return "", "", 0, fmt.Errorf("gemini rate limit reached (limit: %d)", b.geminiRPMLimit)
	}

	response, tokens, err := b.geminiProvider.GenerateResponse(ctx, message, conversationHistory)
	if err != nil {
		return "", "", 0, err
	}

	b.geminiRPMUsed++
	return response, "gemini", tokens, nil
}

// resetRPMIfNeeded resets RPM counters if a minute has passed
func (b *MultiProviderBalancer) resetRPMIfNeeded() {
	if time.Since(b.lastResetTime) >= time.Minute {
		b.groqRPMUsed = 0
		b.geminiRPMUsed = 0
		b.lastResetTime = time.Now()
	}
}

// GetRPMStatus returns current RPM usage
func (b *MultiProviderBalancer) GetRPMStatus() map[string]interface{} {
	b.resetRPMIfNeeded()
	return map[string]interface{}{
		"groq": map[string]int{
			"used":  b.groqRPMUsed,
			"limit": b.groqRPMLimit,
		},
		"gemini": map[string]int{
			"used":  b.geminiRPMUsed,
			"limit": b.geminiRPMLimit,
		},
	}
}

// ProviderStats holds statistics for load balancing
type ProviderStats struct {
	SuccessRate    float64
	AvgResponseMs  int
	TokensUsedHour int
	LastError      string
	LastErrorTime  time.Time
}
