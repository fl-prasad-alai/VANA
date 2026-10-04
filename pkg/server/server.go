// Package server holds VANA's HTTP API. It is shared by the local server
// (main.go) and the Vercel serverless function (api/index.go) so the two can
// never drift apart again.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"emerald-moss-api/pkg/database"
	"emerald-moss-api/pkg/orchestration"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	tokenTTL          = 7 * 24 * time.Hour
	minPasswordLength = 8
	maxPasswordBytes  = 72 // bcrypt ignores anything longer
	maxMessageRunes   = 4000
	maxBodyBytes      = 64 << 10
	titleRunes        = 60
)

// Known placeholder secrets from .env.example and old defaults. Signing tokens
// with one of these would let anyone forge a login, so they are refused.
var placeholderSecrets = []string{
	"your-super-secret",
	"your_super_secret",
	"change-in-production",
	"change_in_production",
	"change-me",
	"change_me",
	"changeme",
}

// Server is VANA's HTTP API
type Server struct {
	jwtSecret []byte
	dbConfig  database.SupabaseConfig

	mu sync.Mutex
	db *database.SupabaseClient

	groq   *orchestration.GroqClient
	gemini *orchestration.GeminiClient
	orch   *orchestration.Orchestrator

	// dummyHash keeps login timing equal whether or not the email exists
	dummyHash []byte
}

// NewFromEnv builds the server from environment variables. A database that is
// unreachable at startup is retried on later requests (important on Vercel,
// where a cold start may race a waking Supabase project).
func NewFromEnv() *Server {
	s := &Server{
		dbConfig: database.SupabaseConfig{
			URL:        os.Getenv("SUPABASE_URL"),
			DBUser:     os.Getenv("SUPABASE_DB_USER"),
			DBPassword: os.Getenv("SUPABASE_DB_PASSWORD"),
			DBName:     os.Getenv("SUPABASE_DB_NAME"),
			ConnString: os.Getenv("DATABASE_URL"),
		},
	}

	secret := os.Getenv("JWT_SECRET")
	if isWeakSecret(secret) {
		log.Printf("[VANA] FATAL CONFIG: JWT_SECRET is missing, too short (<32 chars) or a placeholder. Auth endpoints are disabled until it is set.")
	} else {
		s.jwtSecret = []byte(secret)
	}

	s.dummyHash, _ = bcrypt.GenerateFromPassword([]byte("vana-timing-equaliser"), bcrypt.DefaultCost)

	groqModel := envOr("GROQ_MODEL", "openai/gpt-oss-20b")
	geminiModel := envOr("GEMINI_MODEL", "gemini-2.5-flash")
	groqKey, geminiKey := os.Getenv("GROQ_API_KEY"), os.Getenv("GEMINI_API_KEY")
	log.Printf("[VANA] GROQ key: %s... len=%d model=%s", safePrefix(groqKey), len(groqKey), groqModel)
	log.Printf("[VANA] GEMINI key: %s... len=%d model=%s", safePrefix(geminiKey), len(geminiKey), geminiModel)
	s.groq = orchestration.NewGroqClient(groqKey, groqModel)
	s.gemini = orchestration.NewGeminiClient(geminiKey, geminiModel)

	s.database() // first connection attempt
	return s
}

// database returns a live client, connecting (or reconnecting) if needed
func (s *Server) database() *database.SupabaseClient {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return s.db
	}
	client, err := database.NewSupabaseClient(s.dbConfig)
	if err != nil {
		log.Printf("[VANA] Database unavailable: %v", err)
		return nil
	}
	log.Println("[VANA] Database connected successfully.")
	s.db = client
	balancer := orchestration.NewMultiProviderBalancer(s.groq, s.gemini, 30, 15)
	s.orch = orchestration.NewOrchestrator(balancer, client, s.gemini)
	return s.db
}

// Handler returns the API routes wrapped in CORS
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/auth/register", s.handleRegister)
	mux.HandleFunc("/api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/auth/me", s.requireAuth(s.handleMe))
	mux.HandleFunc("/api/chat", s.requireAuth(s.handleChat))
	mux.HandleFunc("/api/conversations", s.requireAuth(s.handleListConversations))
	mux.HandleFunc("/api/conversations/", s.requireAuth(s.handleConversationMessages))
	return corsMiddleware(mux)
}

// ---------- auth ----------

type ctxKey struct{}

// requireAuth verifies the Bearer JWT and puts the user ID in the request context
func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.jwtSecret == nil {
			writeError(w, http.StatusServiceUnavailable, "Authentication is not configured on the server.")
			return
		}
		raw, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || raw == "" {
			writeError(w, http.StatusUnauthorized, "Please sign in to continue.")
			return
		}
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
			return s.jwtSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "Your session has expired. Please sign in again.")
			return
		}
		if _, err := uuid.Parse(claims.Subject); err != nil {
			writeError(w, http.StatusUnauthorized, "Your session has expired. Please sign in again.")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, claims.Subject)))
	}
}

func userIDFrom(r *http.Request) string {
	id, _ := r.Context().Value(ctxKey{}).(string)
	return id
}

func (s *Server) issueToken(userID string) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}).SignedString(s.jwtSecret)
}

// publicUser is the user shape the frontend expects (see src/hooks/useAuth.tsx)
type publicUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	CreatedAt time.Time `json:"createdAt"`
}

func toPublic(u *database.User) publicUser {
	return publicUser{ID: u.ID, Email: u.Email, FullName: u.FullName, CreatedAt: u.CreatedAt}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	fullName := strings.TrimSpace(req.FullName)
	switch {
	case fullName == "" || utf8.RuneCountInString(fullName) > 100:
		writeError(w, http.StatusBadRequest, "Please enter your name.")
		return
	case !validEmail(email):
		writeError(w, http.StatusBadRequest, "Please enter a valid email address.")
		return
	case utf8.RuneCountInString(req.Password) < minPasswordLength:
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters.")
		return
	case len(req.Password) > maxPasswordBytes:
		writeError(w, http.StatusBadRequest, "Password is too long (maximum 72 characters).")
		return
	}
	if s.jwtSecret == nil {
		writeError(w, http.StatusServiceUnavailable, "Authentication is not configured on the server.")
		return
	}
	db := s.database()
	if db == nil {
		writeError(w, http.StatusServiceUnavailable, "VANA is waking up. Please try again in a moment.")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Register: bcrypt failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Registration failed. Please try again.")
		return
	}
	hashStr := string(hash)
	user := &database.User{ID: uuid.New().String(), Email: email, FullName: fullName, IsActive: true, PasswordHash: &hashStr}
	if err := db.CreateUser(r.Context(), user); err != nil {
		if errors.Is(err, database.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "An account with this email already exists. Please sign in.")
			return
		}
		log.Printf("Register: DB error: %v", err)
		writeError(w, http.StatusInternalServerError, "Registration failed. Please try again.")
		return
	}
	s.respondWithSession(w, http.StatusCreated, user)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if s.jwtSecret == nil {
		writeError(w, http.StatusServiceUnavailable, "Authentication is not configured on the server.")
		return
	}
	db := s.database()
	if db == nil {
		writeError(w, http.StatusServiceUnavailable, "VANA is waking up. Please try again in a moment.")
		return
	}

	const invalid = "Invalid email or password."
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" || len(req.Password) > maxPasswordBytes {
		writeError(w, http.StatusUnauthorized, invalid)
		return
	}
	user, err := db.GetUserByEmail(r.Context(), email)
	if err != nil {
		log.Printf("Login: lookup failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Sign in failed. Please try again.")
		return
	}
	if user == nil || user.PasswordHash == nil || !user.IsActive {
		// Same bcrypt cost as a real check so response time doesn't reveal which emails exist
		_ = bcrypt.CompareHashAndPassword(s.dummyHash, []byte(req.Password))
		if user != nil && user.PasswordHash == nil {
			log.Printf("Login: legacy account without password: %s", user.ID)
		}
		writeError(w, http.StatusUnauthorized, invalid)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, invalid)
		return
	}
	if err := db.TouchLastLogin(r.Context(), user.ID); err != nil {
		log.Printf("Login: last_login update failed: %v", err)
	}
	log.Printf("Login success: UserID=%s", user.ID)
	s.respondWithSession(w, http.StatusOK, user)
}

func (s *Server) respondWithSession(w http.ResponseWriter, status int, user *database.User) {
	token, err := s.issueToken(user.ID)
	if err != nil {
		log.Printf("Token signing failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Sign in failed. Please try again.")
		return
	}
	writeJSON(w, status, map[string]interface{}{"success": true, "token": token, "user": toPublic(user)})
}

// handleMe lets the frontend confirm a stored token is still valid
func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "userId": userIDFrom(r)})
}

// ---------- chat ----------

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	var req struct {
		Message        string `json:"message"`
		ConversationID string `json:"conversationId"`
		IsVoiceInput   bool   `json:"isVoiceInput"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		writeError(w, http.StatusBadRequest, "Message cannot be empty.")
		return
	}
	if utf8.RuneCountInString(message) > maxMessageRunes {
		writeError(w, http.StatusBadRequest, "Message is too long. Please keep it under 4000 characters.")
		return
	}

	db := s.database()
	if db == nil {
		writeError(w, http.StatusServiceUnavailable, "VANA is waking up. Please try again in a moment.")
		return
	}
	ctx := r.Context()
	userID := userIDFrom(r) // identity comes only from the verified token, never the body

	if err := db.CheckRateLimit(ctx, userID); err != nil {
		if strings.Contains(err.Error(), "exceeded") {
			writeError(w, http.StatusTooManyRequests, "You're sending messages quickly. Take a slow breath and try again in a minute.")
			return
		}
		log.Printf("Rate limit check warning: %v", err)
	}

	// Continue an existing conversation the user owns, or start a new one
	conversationID := strings.TrimSpace(req.ConversationID)
	isNew := false
	if conversationID != "" {
		if _, err := uuid.Parse(conversationID); err != nil {
			conversationID = "" // legacy client-side ids (e.g. "1") start a fresh conversation
		} else if owned, err := db.UserOwnsConversation(ctx, userID, conversationID); err != nil {
			log.Printf("Chat: ownership check failed: %v", err)
			writeError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
			return
		} else if !owned {
			writeError(w, http.StatusNotFound, "Conversation not found.")
			return
		}
	}
	var title string
	if conversationID == "" {
		title = makeTitle(message)
		conv, err := db.CreateConversation(ctx, userID, title)
		if err != nil {
			log.Printf("Chat: create conversation failed: %v", err)
			writeError(w, http.StatusInternalServerError, "Something went wrong. Please try again.")
			return
		}
		conversationID, isNew = conv.ID, true
	}

	started := time.Now()
	result, err := s.orch.GenerateResponse(ctx, userID, conversationID, message, req.IsVoiceInput)
	if err != nil {
		log.Printf("Orchestration error: %v", err)
		writeError(w, http.StatusInternalServerError, "I couldn't respond just now. Please try again.")
		return
	}
	elapsedMs := int(time.Since(started).Milliseconds())

	// Persist the exchange. A storage failure is logged but doesn't hide VANA's reply.
	aiText, _ := result["text"].(string)
	provider, _ := result["provider"].(string)
	sentiment, _ := result["sentiment_score"].(float64)
	crisis, _ := result["crisis"].(bool)
	label := sentimentLabel(sentiment, crisis)
	model := s.modelFor(provider)
	userMsg := &database.Message{
		ConversationID: conversationID, UserID: userID, Sender: "user", Content: message,
		ContainsCrisisKeywords: crisis, FlaggedForReview: crisis,
	}
	aiMsg := &database.Message{
		ConversationID: conversationID, UserID: userID, Sender: "ai", Content: aiText,
		SentimentScore: &sentiment, SentimentLabel: &label, AIProvider: &provider, AIModel: model,
		ResponseTimeMs: &elapsedMs, ContainsCrisisKeywords: crisis,
	}
	if err := db.StoreExchange(ctx, userMsg, aiMsg); err != nil {
		log.Printf("Chat: storing messages failed for conversation %s: %v", conversationID, err)
	}
	if err := db.LogAPIUsage(ctx, userID, "/api/chat", provider, 0); err != nil {
		log.Printf("Chat: usage log failed: %v", err)
	}

	result["conversationId"] = conversationID
	if isNew {
		result["conversationTitle"] = title
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) modelFor(provider string) *string {
	var m string
	switch provider {
	case "groq":
		m = s.groq.GetModel()
	case "gemini":
		m = s.gemini.GetModel()
	default:
		return nil
	}
	return &m
}

// ---------- history ----------

func (s *Server) handleListConversations(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	db := s.database()
	if db == nil {
		writeError(w, http.StatusServiceUnavailable, "VANA is waking up. Please try again in a moment.")
		return
	}
	conversations, err := db.ListConversations(r.Context(), userIDFrom(r), 50)
	if err != nil {
		log.Printf("List conversations failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Couldn't load your conversations.")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "conversations": conversations})
}

// handleConversationMessages serves GET /api/conversations/{id}/messages
func (s *Server) handleConversationMessages(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/api/conversations/")
	conversationID, suffix, _ := strings.Cut(rest, "/")
	if suffix != "messages" {
		writeError(w, http.StatusNotFound, "Not found.")
		return
	}
	if _, err := uuid.Parse(conversationID); err != nil {
		writeError(w, http.StatusNotFound, "Conversation not found.")
		return
	}
	db := s.database()
	if db == nil {
		writeError(w, http.StatusServiceUnavailable, "VANA is waking up. Please try again in a moment.")
		return
	}
	ctx := r.Context()
	owned, err := db.UserOwnsConversation(ctx, userIDFrom(r), conversationID)
	if err != nil {
		log.Printf("Messages: ownership check failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Couldn't load this conversation.")
		return
	}
	if !owned {
		writeError(w, http.StatusNotFound, "Conversation not found.")
		return
	}
	messages, err := db.GetConversationMessages(ctx, conversationID)
	if err != nil {
		log.Printf("Messages: load failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Couldn't load this conversation.")
		return
	}
	type outMsg struct {
		ID             string    `json:"id"`
		Sender         string    `json:"sender"`
		Text           string    `json:"text"`
		CreatedAt      time.Time `json:"createdAt"`
		SentimentScore *float64  `json:"sentimentScore,omitempty"`
		SentimentLabel *string   `json:"sentimentLabel,omitempty"`
	}
	out := make([]outMsg, 0, len(messages))
	for _, m := range messages {
		out = append(out, outMsg{m.ID, m.Sender, m.Content, m.CreatedAt, m.SentimentScore, m.SentimentLabel})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true, "conversationId": conversationID, "messages": out})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// ---------- helpers ----------

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Auth uses Bearer tokens (not cookies), so a wildcard origin is safe here
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return false
	}
	return true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request.")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError returns {"success": false, "message": ...}, which the frontend shows to the user
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{"success": false, "message": message})
}

func validEmail(email string) bool {
	if len(email) > 254 || strings.ContainsAny(email, " <>") {
		return false
	}
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email && strings.Contains(email[strings.LastIndex(email, "@"):], ".")
}

func makeTitle(message string) string {
	t := strings.Join(strings.Fields(message), " ")
	if utf8.RuneCountInString(t) <= titleRunes {
		return t
	}
	runes := []rune(t)[:titleRunes]
	if i := strings.LastIndex(string(runes), " "); i > titleRunes/2 {
		return string(runes)[:i] + "…"
	}
	return string(runes) + "…"
}

func sentimentLabel(score float64, crisis bool) string {
	switch {
	case crisis:
		return "critical"
	case score > 0.6:
		return "positive"
	case score < 0.4:
		return "negative"
	default:
		return "neutral"
	}
}

func isWeakSecret(secret string) bool {
	if len(secret) < 32 {
		return true
	}
	lower := strings.ToLower(secret)
	for _, p := range placeholderSecrets {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func safePrefix(s string) string {
	if len(s) < 8 {
		return "EMPTY"
	}
	return s[:8]
}
