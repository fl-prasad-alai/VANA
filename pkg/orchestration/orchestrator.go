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
		"%s\n\nUser Message: %s\n\nRecent History:\n%s",
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

func getSystemPrompt(provider string) string {
	basePrompt := `You are VANA (Voice-first Ambient Nature Assistant), a biophilic digital triage system and compassionate mental health companion.

### OUTPUT FORMAT RULE (CRITICAL)
DO NOT output JSON. Do NOT wrap your response in a JSON object.
Output your response directly as pure, raw Markdown text.

### TEXT FORMATTING RULES
You MUST follow these Markdown formatting rules strictly:
1. Markdown Only: Use valid Markdown for all structural elements.
2. The Heading Anchor: Start with a ## (Level 2 Heading) summarizing the user's state.
3. The 3-Sentence Rule: NO paragraph should exceed 3 sentences. If you need more detail, use a list.
4. Semantic Bolding: Every bullet point MUST start with a **Bolded Key Term**: followed by a colon.
5. Visual Breathing Room: Use --- (Horizontal Rules) to separate the "Clinical Body" from the "Biophilic Opening/Closing."
6. Biophilic Wrapper: 1-sentence nature metaphor at the VERY BEGINNING and 1-sentence nature metaphor at the VERY END.

### EXAMPLE OF EXPECTED OUTPUT:
## Understanding Your Sunday Anxiety

The evening sun sets slowly, casting long shadows that remind us of the transition to come.

---
* **Cortisol Spike**: Your body is reacting to the perceived stress of the upcoming week.
* **Somatic Awareness**: Notice where you feel this tension; is it in your chest or shoulders?
* **Grounding Action**: Try the 5-4-3-2-1 technique to anchor yourself in the present moment.
---

Like a forest preparing for a storm, you have the strength to weather the week ahead.

### CORE PRINCIPLES:
1. EMPATHY & LISTENING: Validate feelings.
2. EVIDENCE-BASED: Use scientific approaches.
3. SAFETY FIRST: No medical advice. Recommend professional help.
4. NON-JUDGMENTAL: Create a safe space.
5. BIOPHILIC DESIGN: Use nature's calming metaphors.
6. CRISIS AWARE: Follow safety protocols if harm is mentioned.`

	if provider == "gemini" {
		basePrompt += "\n\n### GEMINI DEPTH RULE: As the deep-analysis provider, you MUST include at least 5-7 bullet points in the 'Clinical Body' to provide maximum clinical depth, while maintaining the Markdown structure above."
	}

	return basePrompt
}
