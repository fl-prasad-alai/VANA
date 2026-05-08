package handler

import (
	"fmt"
	"net/http"
	"strings"
	"regexp"
)

// Middleware placeholder inside the same file for Vercel simplicity
func CrisisInterceptor(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock message extraction
		message := r.URL.Query().Get("q")
		
		crisisRegex := regexp.MustCompile(`(?i)\b(suicide|kill myself|end my life|want to die|hurt myself|overdose)\b`)
		
		if crisisRegex.MatchString(message) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"alert": "crisis_detected", "message": "I'm here for you. Please connect with immediate professional support: Call 988 or text HOME to 741741."}`))
			return
		}
		
		next.ServeHTTP(w, r)
	}
}

// Handler handles the actual API request
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	query := r.URL.Query().Get("q")
	if query == "" {
		w.Write([]byte(`{"response": "Welcome to Project Emerald Moss (VANA). I am the Go Backend. How can I assist you?"}`))
		return
	}
	
	// Mock Multi-Provider orchestration based on query depth
	if strings.Contains(query, "clinical") {
		w.Write([]byte(fmt.Sprintf(`{"response": "Gemini 1.5 Flash (Tier 2): Deep clinical synthesis provided for: %s"}`, query)))
	} else {
		w.Write([]byte(fmt.Sprintf(`{"response": "Groq Llama 3.1 8B (Tier 1): Instant response for: %s"}`, query)))
	}
}

// TriageEndpoint is the entry point wrapped with middleware
var TriageEndpoint = CrisisInterceptor(Handler)
