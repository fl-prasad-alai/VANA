package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"emerald-moss-api/pkg/database"
	"emerald-moss-api/pkg/orchestration"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	dbClient      *database.SupabaseClient
	aiOrchestrator *orchestration.Orchestrator
	jwtSecret     []byte
	once          sync.Once
)

func initApp() {
	once.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "your-super-secret-key-change-in-production"
		}
		jwtSecret = []byte(secret)

		config := database.SupabaseConfig{
			URL:        os.Getenv("SUPABASE_URL"),
			DBUser:     os.Getenv("SUPABASE_DB_USER"),
			DBPassword: os.Getenv("SUPABASE_DB_PASSWORD"),
			DBName:     os.Getenv("SUPABASE_DB_NAME"),
			ConnString: os.Getenv("DATABASE_URL"),
		}

		client, err := database.NewSupabaseClient(config)
		if err != nil {
			log.Printf("Warning: Database connection failed: %v", err)
		}
		dbClient = client

		// Initialize AI Balancer & Orchestrator
		groqKey := os.Getenv("GROQ_API_KEY")
		geminiKey := os.Getenv("GEMINI_API_KEY")
		groqModel := os.Getenv("GROQ_MODEL")
		if groqModel == "" {
			groqModel = "llama-3.1-8b-instant"
		}
		geminiModel := os.Getenv("GEMINI_MODEL")
		if geminiModel == "" {
			geminiModel = "gemini-2.0-flash"
		}

		log.Printf("[VANA] GROQ key: %s... len=%d model=%s", safePrefix(groqKey), len(groqKey), groqModel)
		log.Printf("[VANA] GEMINI key: %s... len=%d model=%s", safePrefix(geminiKey), len(geminiKey), geminiModel)

		groq := orchestration.NewGroqClient(groqKey, groqModel)
		gemini := orchestration.NewGeminiClient(geminiKey, geminiModel)
		balancer := orchestration.NewMultiProviderBalancer(groq, gemini, 30, 15)
		aiOrchestrator = orchestration.NewOrchestrator(balancer, dbClient, gemini)
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	initApp()

	mux := http.NewServeMux()

	// Routes (Standard patterns)
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/chat", handleChat)
	mux.HandleFunc("/api/health", handleHealth)

	// CORS Middleware
	handler := corsMiddleware(mux)
	handler.ServeHTTP(w, r)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if dbClient == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "User registered (Mocked)",
			"token":   "mock-token",
			"user": map[string]string{
				"email":    req.Email,
				"fullName": req.FullName,
			},
		})
		return
	}

	ctx := r.Context()
	existingUser, err := dbClient.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		http.Error(w, "User already exists with this email", http.StatusBadRequest)
		return
	}

	user := &database.User{
		ID:       uuid.New().String(),
		Email:    req.Email,
		FullName: req.FullName,
		IsActive: true,
	}

	if err := dbClient.CreateUser(ctx, user); err != nil {
		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	token := generateToken(user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"user":    user,
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if dbClient == nil {
		if req.Email == "admin@vana.com" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"token":   "mock-token",
				"user": map[string]string{
					"email":    req.Email,
					"fullName": "Admin User",
				},
			})
			return
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := dbClient.GetUserByEmail(ctx, req.Email)
	if err != nil || user == nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token := generateToken(user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"token":   token,
		"user":    user,
	})
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Message        string `json:"message"`
		ConversationID string `json:"conversationId"`
		UserID         string `json:"userId"`
		IsVoiceInput   bool   `json:"isVoiceInput"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Normalize IDs — non-UUID values from the frontend (e.g. "1") would cause DB errors
	if req.UserID == "" {
		req.UserID = "00000000-0000-0000-0000-000000000000"
	} else if _, err := uuid.Parse(req.UserID); err != nil {
		req.UserID = "00000000-0000-0000-0000-000000000000"
	}
	if req.ConversationID == "" {
		req.ConversationID = "00000000-0000-0000-0000-000000000000"
	} else if _, err := uuid.Parse(req.ConversationID); err != nil {
		req.ConversationID = "00000000-0000-0000-0000-000000000000"
	}

	ctx := r.Context()

	if dbClient != nil {
		if err := dbClient.CheckRateLimit(ctx, req.UserID); err != nil {
			if strings.Contains(err.Error(), "exceeded") {
				log.Printf("Rate limit exceeded for user %s: %v", req.UserID, err)
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
			log.Printf("Rate limit check warning: %v", err)
		}
	}

	result, err := aiOrchestrator.GenerateResponse(ctx, req.UserID, req.ConversationID, req.Message, req.IsVoiceInput)
	if err != nil {
		log.Printf("Orchestration error: %v", err)
		http.Error(w, "Failed to generate response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func safePrefix(s string) string {
	if len(s) < 8 {
		return "EMPTY"
	}
	return s[:8]
}

func generateToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}
