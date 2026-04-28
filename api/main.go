package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"emerald-moss-api/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	dbClient  *database.SupabaseClient
	jwtSecret = []byte("your-super-secret-key-change-in-production")
)

func main() {
	_ = godotenv.Load()

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		jwtSecret = []byte(secret)
	}

	config := database.SupabaseConfig{
		URL:        os.Getenv("SUPABASE_URL"),
		DBUser:     os.Getenv("SUPABASE_DB_USER"),
		DBPassword: os.Getenv("SUPABASE_DB_PASSWORD"),
		DBName:     os.Getenv("SUPABASE_DB_NAME"),
	}

	// Default for docker-compose
	if config.URL == "" { config.URL = "postgres" }
	if config.DBUser == "" { config.DBUser = "postgres" }
	if config.DBPassword == "" { config.DBPassword = "postgres" }
	if config.DBName == "" { config.DBName = "emerald_moss" }

	client, err := database.NewSupabaseClient(config)
	if err != nil {
		log.Printf("Warning: Database connection failed (checking for health later): %v", err)
	}
	dbClient = client

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("POST /api/auth/register", handleRegister)
	mux.HandleFunc("POST /api/auth/login", handleLogin)
	mux.HandleFunc("POST /api/chat", handleChat)
	mux.HandleFunc("GET /api/health", handleHealth)

	// CORS Middleware
	handler := corsMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" { port = "3000" }

	fmt.Printf("VANA Backend running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
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
		// Mock success if DB is not available
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

	// Check if user already exists
	existingUser, err := dbClient.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		log.Printf("Registration Error: User already exists: %s", req.Email)
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
		log.Printf("DB Error during registration for %s: %v", req.Email, err)
		http.Error(w, "Registration failed. Please try again later.", http.StatusInternalServerError)
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
		// Mock success for admin
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
	if err != nil {
		log.Printf("Login Error: GetUserByEmail failed for %s: %v", req.Email, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if user == nil {
		log.Printf("Login Error: User not found: %s", req.Email)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Printf("Login Success: UserID=%s", user.ID)

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

	// Simple mock response
	response := "I'm here to support you. That sounds like something we should explore. How does that make you feel?"
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        uuid.New().String(),
		"response":  response,
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
