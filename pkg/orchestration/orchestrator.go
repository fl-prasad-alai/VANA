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
func (o *Orchestrator) GenerateResponse(ctx context.Context, userID, conversationID, messageText string, isVoiceInput bool) (map[string]interface{}, error) {
	// If it's voice input, pass through the Refiner first
	if isVoiceInput {
		refinedText, err := o.RefineTranscription(ctx, messageText)
		if err == nil && refinedText != "" {
			messageText = refinedText
		} else {
			log.Printf("Refiner failed, falling back to raw STT: %v", err)
		}
	}

	// 1. Clinical Guardrail Check
	if o.db != nil {
		keywords, err := o.db.GetCrisisKeywords(ctx)
		if err == nil {
			for _, k := range keywords {
				if strings.Contains(strings.ToLower(messageText), strings.ToLower(k.Keyword)) {
					// Context Filter: keywords like 'stress', 'low', or 'help' must not trigger crisis
					// if accompanied by 'songs', 'movies', 'music', or 'recommendations'.
					lowerMsg := strings.ToLower(messageText)
					isCopingMechanism := strings.Contains(lowerMsg, "song") || 
									   strings.Contains(lowerMsg, "music") || 
									   strings.Contains(lowerMsg, "movie") || 
									   strings.Contains(lowerMsg, "film") || 
									   strings.Contains(lowerMsg, "recommend") ||
									   strings.Contains(lowerMsg, "list") ||
									   strings.Contains(lowerMsg, "doctor") || 
									   strings.Contains(lowerMsg, "address") || 
									   strings.Contains(lowerMsg, "near me")
					
					if !isCopingMechanism {
						// Crisis detected! Force Gemini for clinical safety
						return o.handleCrisis(ctx, userID, conversationID, messageText, k)
					}
				}
			}
		}
	}

	// 2. Context Retrieval (Last 5 messages)
	var contextLines []string
	hasHistory := false
	if o.db != nil {
		history, err := o.db.GetConversationMessages(ctx, conversationID)
		if err == nil && len(history) > 0 {
			hasHistory = true
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
		"SYSTEM CONTEXT: HasHistory=%v\n%s\n\nUser Message: %s\n\nRecent History:\n%s",
		hasHistory,
		clinicalContext,
		messageText,
		strings.Join(contextLines, "\n"),
	)

	// Call Balancer with timeout
	ctxAI, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	providerName := "groq"
	if useGemini {
		providerName = "gemini"
	}
	systemPrompt := getSystemPrompt(providerName)

	responseJSON, provider, _, err := o.balancer.HandleChat(ctxAI, systemPrompt, finalPrompt, nil, useGemini)
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

	// Determine suggested track
	suggestedTrack := "forest_morning"
	if output.SentimentScore < 0.3 {
		suggestedTrack = "monsoon_rain"
	} else if strings.Contains(strings.ToLower(messageText), "sleep") || strings.Contains(strings.ToLower(messageText), "insomnia") {
		suggestedTrack = "dusk_valley"
	}

	return map[string]interface{}{
		"text":                 output.Text,
		"sentiment_score":      output.SentimentScore,
		"provider":             provider,
		"timestamp":            time.Now().Format(time.RFC3339),
		"suggested_mood_audio": suggestedTrack,
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
		"text":                 fmt.Sprintf("I hear you, and I am very concerned. Your path is valuable, and right now, you need a human guide to help you find the light again. Please reach out to these professionals immediately: %s", protocol),
		"sentiment_score":      0.1,
		"provider":             "clinical-fallback",
		"timestamp":            time.Now().Format(time.RFC3339),
		"crisis":               true,
		"suggested_mood_audio": "monsoon_rain", // Heavy rain for crisis situations
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
	basePrompt := `You are VANA (Voice-first Ambient Nature Assistant), a biophilic digital triage system. You must strictly mirror the user's input language (English, Hindi, Marathi, or Hinglish).

### LOGIC CONSTRAINTS (MANDATORY):
1. ZERO-HALLUCINATION: If HasHistory=false, do NOT invent past events. Do not use phrases like "As we discussed before" or "Previously". 
2. INTENT-FIRST LISTS: If a user asks for a Count (e.g., "10 songs", "5 movies"), do NOT explain the meaning of the number or talk about culture. Provide the numbered Markdown list immediately after a 1-sentence biophilic opening.
3. MEDICAL FIREWALL: FORBIDDEN from providing specific drug names (e.g., Xanax, Zoloft) or dosages. 
   - Mandatory Medication Template: "I cannot provide specific medication names or quantities as that requires a professional clinical diagnosis. Generally, doctors explore classes like SSRIs or Anxiolytics, but only a licensed physician can determine what is safe for your body."
4. FACT-CHECK: Only provide real, verifiable song/movie titles. Do not invent titles.

### OUTPUT FORMAT:
- Use ## for the Heading.
- Use --- for the divider.
- Bolded Terms for bullet points.
- 1-sentence nature metaphor at the START and END.

### CORE PRINCIPLES:
1. MIRRORING: If user speaks Marathi, VANA speaks Marathi.
2. BIOPHILIC DESIGN: Nature metaphors are mandatory but must be brief (1 sentence).
3. SAFETY: Follow Warm Handoff protocol for crises.`

	if provider == "gemini" {
		basePrompt += "\n\n### GEMINI DEPTH RULE: As the deep-analysis provider, you MUST include at least 5-7 bullet points in the 'Clinical Body' to provide maximum clinical depth, while maintaining the Markdown structure above."
	}

	return basePrompt
}

// RefineTranscription cleans up messy STT inputs using Groq
func (o *Orchestrator) RefineTranscription(ctx context.Context, input string) (string, error) {
	systemPrompt := `You are a multilingual transcription editor for VANA-Mind. You will receive raw text from a STT engine. Your job:
1. Correct spelling, grammar, and phonetic errors.
2. Maintain the original language (English, Hindi, Marathi, or Hinglish).
3. SAFETY: Prioritize logical interpretations over phonetic noise (e.g., "meditation" vs "medicine").
4. Detect the language code (e.g., "en", "hi", "mr", "hinglish") and include it in your response.
Output ONLY a JSON object: {"text": "corrected text", "language": "code"}. No explanations.`

	finalPrompt := "Raw STT Input:\n" + input

	// Use Groq for speed (useGemini = false)
	ctxAI, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	responseJSON, _, _, err := o.balancer.HandleChat(ctxAI, systemPrompt, finalPrompt, nil, false)
	if err != nil {
		return "", err
	}

	// Extract text and language
	var output struct {
		Text     string `json:"text"`
		Language string `json:"language"`
	}
	cleanJSON := extractJSON(responseJSON)
	if err := json.Unmarshal([]byte(cleanJSON), &output); err == nil && output.Text != "" {
		return output.Text, nil
	}

	return strings.TrimSpace(responseJSON), nil
}
