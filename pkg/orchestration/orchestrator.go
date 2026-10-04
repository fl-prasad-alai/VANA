// pkg/orchestration/orchestrator.go

package orchestration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
	"unicode"

	"emerald-moss-api/pkg/database"
)

// ErrAIUnavailable means no AI provider could answer (usually free-tier limits)
var ErrAIUnavailable = errors.New("no AI provider available")

// Orchestrator handles the high-level AI logic for VANA
type Orchestrator struct {
	balancer *MultiProviderBalancer
	db       *database.SupabaseClient
	gemini   *GeminiClient

	knowMu      sync.Mutex
	knowCount   int
	knowChecked time.Time
}

// hasKnowledge reports whether reviewed clinical knowledge exists (re-checked every 10 minutes)
func (o *Orchestrator) hasKnowledge(ctx context.Context) bool {
	o.knowMu.Lock()
	defer o.knowMu.Unlock()
	if time.Since(o.knowChecked) > 10*time.Minute {
		if n, err := o.db.CountClinicalKnowledge(ctx); err == nil {
			o.knowCount, o.knowChecked = n, time.Now()
		}
	}
	return o.knowCount > 0
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

	// 1. Conversation history (last 10 messages): used by the crisis check,
	// question pacing and the prompt
	var contextLines []string
	hasHistory, lastAIAsked, lastAICrisis := false, false, false
	if o.db != nil {
		history, err := o.db.GetConversationMessages(ctx, conversationID)
		if err == nil && len(history) > 0 {
			hasHistory = true
			start := len(history) - 10
			if start < 0 {
				start = 0
			}
			for i := start; i < len(history); i++ {
				contextLines = append(contextLines, fmt.Sprintf("%s: %s", history[i].Sender, history[i].Content))
			}
			for i := len(history) - 1; i >= 0; i-- {
				if history[i].Sender == "ai" {
					last := strings.TrimSpace(history[i].Content)
					lastAIAsked = strings.HasSuffix(last, "?")
					lastAICrisis = isCrisisTemplate(last)
					break
				}
			}
		}
	}

	// 2. Clinical Guardrail Check
	// Whole-word/phrase matching so "help" does not fire on "helpful".
	// High/critical keywords trigger the human handoff; medium/low ones keep the
	// conversation going but tell the model to respond with extra care.
	distressNote := ""
	keywords := builtinCrisisKeywords
	if o.db != nil {
		if dbKeywords, err := o.db.GetCrisisKeywords(ctx); err == nil {
			keywords = append(dbKeywords, builtinCrisisKeywords...)
		}
	}
	lowerMsg := normalizeIndic(strings.ToLower(messageText))
	// Context Filter: asking for songs, movies, doctors etc. is a coping request, not a crisis.
	isCopingMechanism := strings.Contains(lowerMsg, "song") ||
		strings.Contains(lowerMsg, "music") ||
		strings.Contains(lowerMsg, "movie") ||
		strings.Contains(lowerMsg, "film") ||
		strings.Contains(lowerMsg, "recommend") ||
		strings.Contains(lowerMsg, "list") ||
		strings.Contains(lowerMsg, "doctor") ||
		strings.Contains(lowerMsg, "address") ||
		strings.Contains(lowerMsg, "near me")
	for _, k := range keywords {
		if !containsPhrase(lowerMsg, normalizeIndic(strings.ToLower(k.Keyword))) {
			continue
		}
		severity := strings.ToLower(k.Severity)
		if (severity == "critical" || severity == "high") && !isCopingMechanism {
			return o.handleCrisis(ctx, userID, conversationID, messageText, k, lastAICrisis)
		}
		if severity == "medium" {
			distressNote = "\nSAFETY NOTE: The person may be feeling hopeless. Reflect their pain first, then gently and directly ask, in their language, whether they are having thoughts of ending their life, and mention Tele-MANAS 14416 (free, 24x7).\n"
		}
	}

	// 3. RAG Integration (Clinical Knowledge)
	clinicalContext := ""
	
	// Embedding search only when there is reviewed knowledge to search (saves a
	// Gemini call and its daily quota on every message while the table is empty)
	if o.db != nil && o.hasKnowledge(ctx) {
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

	// 4. Provider: Groq answers; Gemini is only the fallback (its free tier is
	// ~20 requests a day, so words like "deep" must not route to it)
	useGemini := false
	pacingNote := ""
	if lastAIAsked {
		pacingNote = "PACING: Your previous reply ended with a question. This time do NOT ask a question; reflect, validate, or offer a short line of presence.\n"
	}

	// Construct Final Prompt
	finalPrompt := fmt.Sprintf(
		"SYSTEM CONTEXT: HasHistory=%v\n%s%s%s\n\nUser Message: %s\n\nRecent History:\n%s\n\n%s%s",
		hasHistory,
		o.memoryContext(ctx, userID),
		distressNote,
		clinicalContext,
		messageText,
		strings.Join(contextLines, "\n"),
		pacingNote,
		scriptNote(messageText), // last, where models weight instructions most
	)

	// Call Balancer with timeout
	ctxAI, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	providerName := "groq"
	if useGemini {
		providerName = "gemini"
	}
	systemPrompt := getSystemPrompt(providerName)

	responseJSON, provider, tokens, err := o.balancer.HandleChat(ctxAI, systemPrompt, finalPrompt, nil, useGemini)
	// The model sometimes answers Hinglish in Devanagari despite the instruction; retry once, firmly.
	if err == nil && !hasDevanagari(messageText) && devanagariShare(responseJSON) > 0.2 {
		log.Printf("Reply script mismatch (Devanagari for Latin input); retrying once")
		retryPrompt := finalPrompt + "\nIMPORTANT: Your previous draft used Devanagari. Write the ENTIRE reply in English letters (Latin script) only."
		if again, p, t, rerr := o.balancer.HandleChat(ctxAI, systemPrompt, retryPrompt, nil, useGemini); rerr == nil && devanagariShare(again) <= 0.2 {
			responseJSON, provider = again, p
			tokens += t
		}
	}
	// Hindi/Marathi verbs carry gender; a reply that guesses the person's gender is regenerated once
	if err == nil && guessesGender(responseJSON) {
		log.Printf("Reply guessed the user's gender; retrying once")
		retryPrompt := finalPrompt + "\nIMPORTANT: Your previous draft guessed the person's gender in a verb (e.g. 'कर रही हैं', 'sochte ho'). Rewrite using only gender-neutral forms such as 'आपको कैसा लग रहा है', 'आपने...', 'तुम्हाला कसं वाटतंय'."
		if again, p, t, rerr := o.balancer.HandleChat(ctxAI, systemPrompt, retryPrompt, nil, useGemini); rerr == nil && !guessesGender(again) {
			responseJSON, provider = again, p
			tokens += t
		}
	}
	if err != nil {
		log.Printf("AI execution failed on every provider: %v", err)
		return nil, ErrAIUnavailable
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
		"_tokens":              tokens, // removed by the server before replying
		"timestamp":            time.Now().Format(time.RFC3339),
		"suggested_mood_audio": suggestedTrack,
	}, nil
}

func (o *Orchestrator) handleCrisis(ctx context.Context, userID, conversationID, messageText string, keyword database.CrisisKeyword, repeat bool) (map[string]interface{}, error) {
	// Log crisis to DB
	if o.db != nil {
		_ = o.db.UpdateConversationCrisisFlag(ctx, conversationID, true, fmt.Sprintf("Keyword detected: %s", keyword.Keyword))
	}
	
	return map[string]interface{}{
		"text":                 crisisReply(messageText, repeat),
		"sentiment_score":      0.1,
		"provider":             "clinical-fallback",
		"timestamp":            time.Now().Format(time.RFC3339),
		"crisis":               true,
		"suggested_mood_audio": "monsoon_rain", // Heavy rain for crisis situations
	}, nil
}

// scriptNote pins the reply to the script the user typed in. Hinglish typed in
// English letters otherwise often gets a Devanagari reply.
func scriptNote(message string) string {
	if hasDevanagari(message) {
		return "REPLY SCRIPT: The user wrote in Devanagari. Reply in Devanagari.\n"
	}
	return "REPLY SCRIPT: The user wrote in English letters. Reply ONLY in English letters (Latin script); if they wrote Hinglish, reply in Hinglish. Do not use Devanagari.\n"
}

func hasDevanagari(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Devanagari) {
			return true
		}
	}
	return false
}

// devanagariShare is the fraction of letters in s that are Devanagari
func devanagariShare(s string) float64 {
	letters, deva := 0, 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.Is(unicode.Mn, r) {
			letters++
			if unicode.In(r, unicode.Devanagari) {
				deva++
			}
		}
	}
	if letters == 0 {
		return 0
	}
	return float64(deva) / float64(letters)
}

// containsPhrase reports whether phrase appears in text starting at a word
// boundary and ending either at a word boundary or with a common English
// ending, so "suicides", "suicidal" and "overdosed" still match while "help"
// does not match "helpful" and "want to die" does not match "want to diet".
func containsPhrase(text, phrase string) bool {
	if phrase == "" {
		return false
	}
	isWordChar := func(b byte) bool {
		return b == '_' || b == '\'' || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
	}
	endings := map[string]bool{"": true, "s": true, "es": true, "d": true, "ed": true, "ing": true, "al": true, "ness": true}
	stems := []string{phrase}
	if strings.HasSuffix(phrase, "e") {
		stems = append(stems, strings.TrimSuffix(phrase, "e")) // suicide -> suicid(al), overdose -> overdos(ing)
	}
	for _, stem := range stems {
		for from := 0; ; {
			i := strings.Index(text[from:], stem)
			if i < 0 {
				break
			}
			start, end := from+i, from+i+len(stem)
			wordEnd := end
			for wordEnd < len(text) && isWordChar(text[wordEnd]) {
				wordEnd++
			}
			rest := text[end:wordEnd]
			if stem != phrase && rest == "" {
				rest = "-" // the bare stem ("suicid") is not a word on its own
			}
			if (start == 0 || !isWordChar(text[start-1])) && endings[rest] {
				return true
			}
			from = start + 1
		}
	}
	return false
}

func extractJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end < start {
		return s
	}
	return s[start : end+1]
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
