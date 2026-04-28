// api/middleware/crisis.go

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"emerald-moss-api/database"
)

// CrisisInterceptor middleware for detecting crisis keywords
type CrisisInterceptor struct {
	db               *database.SupabaseClient
	crisisKeywords   map[string]string // keyword -> severity
	crisisRegex      *regexp.Regexp
}

// NewCrisisInterceptor creates a new crisis interceptor
func NewCrisisInterceptor(db *database.SupabaseClient) (*CrisisInterceptor, error) {
	ctx := context.Background()
	keywords, err := db.GetCrisisKeywords(ctx)
	if err != nil {
		return nil, err
	}

	crisisMap := make(map[string]string)
	for _, k := range keywords {
		crisisMap[strings.ToLower(k.Keyword)] = k.Severity
	}

	// Build regex pattern from keywords
	patterns := []string{
		`\b(suicide|kill myself|end my life|want to die|don't want to live)\b`,
		`\b(hurt myself|cut myself|self-harm|overdose|pills)\b`,
		`\b(noose|gun|hang myself|jump)\b`,
	}
	regex := regexp.MustCompile(strings.Join(patterns, "|"))

	return &CrisisInterceptor{
		db:             db,
		crisisKeywords: crisisMap,
		crisisRegex:    regex,
	}, nil
}

// Middleware returns the HTTP middleware function
func (ci *CrisisInterceptor) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check POST requests to chat endpoint
		if r.Method != http.MethodPost || !strings.Contains(r.URL.Path, "/api/chat") {
			next.ServeHTTP(w, r)
			return
		}

		// Parse request body
		var chatReq database.ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Detect crisis keywords
		if ci.DetectCrisis(chatReq.Message) {
			// Crisis detected - bypass AI and return immediate help
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK) // Return 200 but with crisis flag

			crisisResponse := map[string]interface{}{
				"success":         true,
				"is_crisis":       true,
				"message":         "I hear that you're going through a very difficult time. Your safety is the top priority.",
				"crisis_message": "If you're in immediate danger, please:",
				"crisis_resources": map[string]string{
					"US Suicide Prevention Lifeline": "Call 988",
					"Crisis Text Line":              "Text HOME to 741741",
					"International":                 "Visit findahelpline.com",
					"Emergency":                     "Call 911 or go to the nearest emergency room",
				},
				"next_steps": "A mental health professional should be contacted as soon as possible.",
			}

			json.NewEncoder(w).Encode(crisisResponse)

			// Log to database
			conversationID := r.Header.Get("X-Conversation-ID")
			userID := r.Header.Get("X-User-ID")
			if conversationID != "" && userID != "" {
				ci.db.UpdateConversationCrisisFlag(context.Background(), conversationID, true, "Crisis keywords detected")
			}

			return
		}

		// No crisis - continue to normal handler
		next.ServeHTTP(w, r)
	})
}

// DetectCrisis checks if a message contains crisis keywords
func (ci *CrisisInterceptor) DetectCrisis(message string) bool {
	// Check regex patterns (high sensitivity)
	if ci.crisisRegex.MatchString(strings.ToLower(message)) {
		return true
	}

	// Check individual keywords
	lowerMsg := strings.ToLower(message)
	for keyword := range ci.crisisKeywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}

	return false
}
