package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"emerald-moss-api/internal/database"
	"emerald-moss-api/internal/orchestration"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	dbClient   *database.SupabaseClient
	aiBalancer *orchestration.MultiProviderBalancer
	jwtSecret  []byte
	once       sync.Once
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

		// Initialize AI Balancer
		groqKey := os.Getenv("GROQ_API_KEY")
		geminiKey := os.Getenv("GEMINI_API_KEY")

		groq := orchestration.NewGroqClient(groqKey, "llama-3.1-8b-instant")
		gemini := orchestration.NewGeminiClient(geminiKey, "gemini-1.5-flash")
		
		aiBalancer = orchestration.NewMultiProviderBalancer(groq, gemini, 30, 15)
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	initApp()

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("POST /api/auth/register", handleRegister)
	mux.HandleFunc("POST /api/auth/login", handleLogin)
	mux.HandleFunc("POST /api/chat", handleChat)
	mux.HandleFunc("GET /api/health", handleHealth)

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
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	response, provider, _, err := aiBalancer.HandleChat(ctx, req.Message, nil, false)
	if err != nil {
		response = "I'm here to support you. How does that make you feel?"
		provider = "mock"
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        uuid.New().String(),
		"response":  response,
		"provider":  provider,
		"sentiment": 0.5,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func generateToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)
	return tokenString
}
