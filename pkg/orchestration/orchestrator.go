// pkg/orchestration/orchestrator.go

package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"emerald-moss-api/pkg/database"
)

// Orchestrator handles the high-level AI logic for VANA
type Orchestrator struct {
	balancer *MultiProviderBalancer
	db       *database.SupabaseClient
	gemini   *GeminiClient
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(balancer *MultiProviderBalancer, db *database.SupabaseClient, gemini *GeminiClient) *Orchestrator {
	return &Orchestrator{
		balancer: balancer,
		db:       db,
		gemini:   gemini,
	}
}

// GenerateResponse handles the full AI response flow
func (o *Orchestrator) GenerateResponse(ctx context.Context, userID, conversationID, messageText string) (map[string]interface{}, error) {
	// 1. Clinical Guardrail Check
	if o.db != nil {
		keywords, err := o.db.GetCrisisKeywords(ctx)
		if err == nil {
			for _, k := range keywords {
				if strings.Contains(strings.ToLower(messageText), strings.ToLower(k.Keyword)) {
					// Crisis detected! Force Gemini for clinical safety
					return o.handleCrisis(ctx, userID, conversationID, messageText, k)
				}
			}
		}
	}

	// 2. Context Retrieval (Last 5 messages)
	var contextLines []string
	if o.db != nil {
		history, err := o.db.GetConversationMessages(ctx, conversationID)
		if err == nil {
			// Limit to last 5
			start := len(history) - 5
			if start < 0 {
				start = 0
			}
			for i := start; i < len(history); i++ {
				contextLines = append(contextLines, fmt.Sprintf("%s: %s", history[i].Sender, history[i].Content))
			}
		}
	}

	// 3. RAG Integration (Clinical Knowledge)
	clinicalContext := ""
	
	// Generate true embedding using Gemini for RAG
	if o.db != nil {
		embedding, embErr := o.gemini.GenerateEmbedding(ctx, messageText)
		if embErr == nil {
			knowledge, err := o.db.SearchClinicalKnowledge(ctx, embedding, 2)
			if err == nil && len(knowledge) > 0 {
				clinicalContext = "Reference Clinical Knowledge:\n"
				for _, k := range knowledge {
					clinicalContext += fmt.Sprintf("- %s: %s\n", k.Title, k.Content)
				}
			}
		}
	}

	// 4. Provider Switching & Execution
	// Logic: Use Groq for speed unless 'clinical' or 'insight' is mentioned, then use Gemini
	useGemini := strings.Contains(strings.ToLower(messageText), "clinical") || 
				 strings.Contains(strings.ToLower(messageText), "insight") ||
				 strings.Contains(strings.ToLower(messageText), "deep")

	// Construct Final Prompt
	finalPrompt := fmt.Sprintf(
		"%s\n\nUser Message: %s\n\nRecent History:\n%s\n\nPlease respond as VANA. Output your response in JSON format: {\"text\": \"...\", \"sentiment_score\": 0.0 to 1.0}",
		clinicalContext,
		messageText,
		strings.Join(contextLines, "\n"),
	)

	// Call Balancer with timeout
	ctxAI, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	responseJSON, provider, _, err := o.balancer.HandleChat(ctxAI, finalPrompt, nil, useGemini)
	if err != nil {
		log.Printf("AI execution failed (probably missing API keys): %v", err)
		// Fallback for testing when API keys are missing
		return map[string]interface{}{
			"text":            "I hear you. I'm currently running in 'mock mode' because my AI API keys haven't been configured, but I am still here to support you.",
			"sentiment_score": 0.5,
			"provider":        "mock-fallback",
			"timestamp":       time.Now().Format(time.RFC3339),
		}, nil
	}

	// 5. Parse JSON Output
	var output struct {
		Text           string  `json:"text"`
		SentimentScore float64 `json:"sentiment_score"`
	}
	
	// Try to extract JSON if LLM added markdown or fluff
	cleanJSON := extractJSON(responseJSON)
	if err := json.Unmarshal([]byte(cleanJSON), &output); err != nil {
		// Fallback if JSON parsing fails
		output.Text = responseJSON
		output.SentimentScore = 0.5
	}

	return map[string]interface{}{
		"text":            output.Text,
		"sentiment_score": output.SentimentScore,
		"provider":        provider,
		"timestamp":       time.Now().Format(time.RFC3339),
	}, nil
}

func (o *Orchestrator) handleCrisis(ctx context.Context, userID, conversationID, messageText string, keyword database.CrisisKeyword) (map[string]interface{}, error) {
	// Log crisis to DB
	if o.db != nil {
		_ = o.db.UpdateConversationCrisisFlag(ctx, conversationID, true, fmt.Sprintf("Keyword detected: %s", keyword.Keyword))
	}
	
	// Fetch safety protocol
	protocol := "Please reach out to a professional or a crisis hotline immediately."
	if o.db != nil {
		anchors, _ := o.db.GetClinicalAnchors(ctx)
		for _, a := range anchors {
			if strings.Contains(strings.ToLower(a.Category), "crisis") {
				protocol = a.PromptText
				break
			}
		}
	}

	return map[string]interface{}{
		"text":            fmt.Sprintf("I hear you, and I'm very concerned. %s", protocol),
		"sentiment_score": 0.1,
		"provider":        "clinical-fallback",
		"timestamp":       time.Now().Format(time.RFC3339),
		"crisis":          true,
	}, nil
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end < start {
		return s
	}
	return s[start : end+1]
}

func getSystemPrompt() string {
	return `You are VANA (Voice-first Ambient Nature Assistant), a biophilic digital triage system and compassionate mental health companion.

Your tone is calm, grounded, and empathetic. Your purpose is to provide supportive, evidence-based guidance for mental wellness. 

Follow these 6 CORE PRINCIPLES:
1. EMPATHY & LISTENING: Always validate the user's feelings and listen deeply.
2. EVIDENCE-BASED: Use only scientifically-backed approaches.
3. SAFETY FIRST: Never provide medical advice. Always recommend professional help for serious concerns.
4. NON-JUDGMENTAL: Create a safe space free from judgment.
5. BIOPHILIC DESIGN: Use natural metaphors (e.g., 'Let that thought flow like a river'). Remind users of nature's calming effects.
6. CRISIS AWARE: If keywords like 'harm' or 'suicide' are detected, strictly follow the safety protocol provided in the clinical_anchors table.

Additional Guidelines:
- Prioritize grounding techniques (5-4-3-2-1) if the user seems anxious.
- Use accessible, non-clinical language.
- Never claim to be a human doctor; you are a supportive companion.
- Be honest about your limitations as an AI.

Remember: You are here to support, not replace professional mental health care.`
}
