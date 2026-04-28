// api/database/models.go

package database

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                    string    `json:"id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	Email                 string    `json:"email"`
	FullName              string    `json:"full_name"`
	EncryptionPubKey      string    `json:"encryption_pub_key"`
	PrivacyMode           string    `json:"privacy_mode"` // 'encrypted', 'local-only', 'supabase'
	ConsentTherapeutic    bool      `json:"consent_therapeutic"`
	ConsentDataCollection bool      `json:"consent_data_collection"`
	ConsentResearch       bool      `json:"consent_research"`
	LastLogin             *time.Time `json:"last_login"`
	IsActive              bool      `json:"is_active"`
}

// Conversation represents a chat session
type Conversation struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"user_id"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	GAD7Score            *int       `json:"gad7_score"`
	GAD7Severity         *string    `json:"gad7_severity"` // 'minimal', 'mild', 'moderate', 'severe'
	Status               string     `json:"status"`        // 'active', 'paused', 'archived', 'closed'
	Title                *string    `json:"title"`
	Summary              *string    `json:"summary"`
	MoodStart            *int       `json:"mood_start"` // -5 to +5
	MoodEnd              *int       `json:"mood_end"`
	SentimentAverage     *float64   `json:"sentiment_average"`
	CrisisDetected       bool       `json:"crisis_detected"`
	EscalationTriggered  bool       `json:"escalation_triggered"`
	EscalationReason     *string    `json:"escalation_reason"`
	EscalationTimestamp  *time.Time `json:"escalation_timestamp"`
	MessageCount         int        `json:"message_count"`
	AIProvider           string     `json:"ai_provider"` // 'groq', 'gemini'
	DurationSeconds      *int       `json:"duration_seconds"`
}

// Message represents a single message in a conversation
type Message struct {
	ID                     string    `json:"id"`
	ConversationID         string    `json:"conversation_id"`
	UserID                 string    `json:"user_id"`
	CreatedAt              time.Time `json:"created_at"`
	Sender                 string    `json:"sender"` // 'user' or 'ai'
	Content                string    `json:"content"`
	EncryptedContent       *string   `json:"encrypted_content"`
	IsEncrypted            bool      `json:"is_encrypted"`
	SentimentScore         *float64  `json:"sentiment_score"` // -1.0 to 1.0
	SentimentLabel         *string   `json:"sentiment_label"`
	EmotionDetected        *string   `json:"emotion_detected"`
	AIModel                *string   `json:"ai_model"`
	AIProvider             *string   `json:"ai_provider"`
	TokensUsed             *int      `json:"tokens_used"`
	ResponseTimeMs         *int      `json:"response_time_ms"`
	ContainsCrisisKeywords bool      `json:"contains_crisis_keywords"`
	FlaggedForReview       bool      `json:"flagged_for_review"`
}

// ClinicalKnowledge represents a RAG vector in the knowledge base
type ClinicalKnowledge struct {
	ID                      string    `json:"id"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	Title                   string    `json:"title"`
	Content                 string    `json:"content"`
	Source                  *string   `json:"source"` // 'dsm5', 'gad7', 'professional', 'evidence-based'
	Category                *string   `json:"category"`
	Embedding               []float64 `json:"embedding"`
	RequiresProfessionalReview bool   `json:"requires_professional_review"`
	IsApprovedForDelivery   bool      `json:"is_approved_for_delivery"`
}

// ClinicalAnchor represents system prompts for the AI
type ClinicalAnchor struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"` // 'system', 'safety', 'empathy', 'clinical'
	PromptText  string `json:"prompt_text"`
	IsActive    bool   `json:"is_active"`
	Version     int    `json:"version"`
	ApprovedBy  *string `json:"approved_by"`
	Description *string `json:"description"`
}

// CrisisKeyword represents a keyword for crisis detection
type CrisisKeyword struct {
	ID       string `json:"id"`
	Keyword  string `json:"keyword"`
	Severity string `json:"severity"` // 'low', 'medium', 'high', 'critical'
	Category *string `json:"category"`
}

// AuthRequest represents a login/register request
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the response after authentication
type AuthResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Token   *string `json:"token,omitempty"`
	User    *User   `json:"user,omitempty"`
}

// ChatRequest represents a user message request
type ChatRequest struct {
	ConversationID string `json:"conversation_id"`
	Message        string `json:"message"`
	IsVoice        bool   `json:"is_voice"`
}

// ChatResponse represents the AI response
type ChatResponse struct {
	ID             string  `json:"id"`
	ConversationID string  `json:"conversation_id"`
	Response       string  `json:"response"`
	Sentiment      float64 `json:"sentiment"`
	ProviderID     string  `json:"provider_id"`
	Timestamp      string  `json:"timestamp"`
	TokensUsed     int     `json:"tokens_used"`
}

// TriageRequest represents the GAD-7 intake form submission
type TriageRequest struct {
	ConversationID string `json:"conversation_id"`
	Responses      [5]int `json:"responses"` // 5 yes/no answers
}

// TriageResponse represents the computed GAD-7 score
type TriageResponse struct {
	GAD7Score      int    `json:"gad7_score"` // 0-21
	Severity       string `json:"severity"`   // minimal, mild, moderate, severe
	Recommendation string `json:"recommendation"`
	CrisisRisk     bool   `json:"crisis_risk"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse wraps paginated results
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int         `json:"total"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
}

// SupabaseConfig holds Supabase connection config
type SupabaseConfig struct {
	URL            string
	AnonKey        string
	ServiceRoleKey string
	DBPassword     string
	DBUser         string
	DBName         string
}

// AIProviderConfig holds API keys for AI providers
type AIProviderConfig struct {
	GroqAPIKey    string
	GroqModel     string
	GroqRPMLimit  int
	GeminiAPIKey  string
	GeminiModel   string
	GeminiRPMLimit int
}

// RequestContext holds request metadata
type RequestContext struct {
	UserID    string
	StartTime time.Time
	RequestID string
	Path      string
}
