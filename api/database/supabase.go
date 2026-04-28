// api/database/supabase.go

package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// SupabaseClient wraps database connections
type SupabaseClient struct {
	db *sql.DB
}

// NewSupabaseClient creates a new database connection
func NewSupabaseClient(config SupabaseConfig) (*SupabaseClient, error) {
	// PostgreSQL connection string
	// For local development in Docker, URL should be the postgres service name
	psqlInfo := fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		config.URL,
		config.DBUser,
		config.DBPassword,
		config.DBName,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SupabaseClient{db: db}, nil
}

// GetUserByEmail retrieves a user by email
func (sc *SupabaseClient) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := sc.db.QueryRowContext(
		ctx,
		"SELECT id, created_at, updated_at, email, full_name, COALESCE(encryption_pub_key, ''), privacy_mode, consent_therapeutic, consent_data_collection, consent_research, is_active FROM public.users WHERE email = $1",
		email,
	).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Email, &user.FullName,
		&user.EncryptionPubKey, &user.PrivacyMode, &user.ConsentTherapeutic,
		&user.ConsentDataCollection, &user.ConsentResearch, &user.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("GetUserByEmail scan error: %w", err)
	}
	return &user, nil
}

// CreateUser creates a new user in both auth and public schemas
func (sc *SupabaseClient) CreateUser(ctx context.Context, user *User) error {
	tx, err := sc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert into auth.users (Mock Auth)
	_, err = tx.ExecContext(ctx, "INSERT INTO auth.users (id, email) VALUES ($1, $2)", user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("auth.users insert failed: %w", err)
	}

	// 2. Insert into public.users
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO public.users (id, email, full_name, privacy_mode, consent_therapeutic, consent_data_collection, consent_research, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.ID, user.Email, user.FullName, "encrypted", false, false, false, true,
	)
	if err != nil {
		return fmt.Errorf("public.users insert failed: %w", err)
	}

	return tx.Commit()
}

// CreateConversation creates a new conversation
func (sc *SupabaseClient) CreateConversation(ctx context.Context, userID string) (*Conversation, error) {
	var conv Conversation
	err := sc.db.QueryRowContext(
		ctx,
		`INSERT INTO public.conversations (user_id, status, ai_provider) 
		 VALUES ($1, 'active', 'groq')
		 RETURNING id, user_id, created_at, updated_at, status, ai_provider`,
		userID,
	).Scan(&conv.ID, &conv.UserID, &conv.CreatedAt, &conv.UpdatedAt, &conv.Status, &conv.AIProvider)
	return &conv, err
}

// StoreMessage stores a message in the database
func (sc *SupabaseClient) StoreMessage(ctx context.Context, msg *Message) error {
	_, err := sc.db.ExecContext(
		ctx,
		`INSERT INTO public.messages (conversation_id, user_id, sender, content, sentiment_score, sentiment_label, ai_provider)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		msg.ConversationID, msg.UserID, msg.Sender, msg.Content,
		msg.SentimentScore, msg.SentimentLabel, msg.AIProvider,
	)
	return err
}

// GetConversationMessages retrieves all messages in a conversation
func (sc *SupabaseClient) GetConversationMessages(ctx context.Context, conversationID string) ([]Message, error) {
	rows, err := sc.db.QueryContext(
		ctx,
		`SELECT id, conversation_id, user_id, created_at, sender, content, sentiment_score, sentiment_label, ai_provider
		 FROM public.messages WHERE conversation_id = $1 ORDER BY created_at ASC`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(
			&msg.ID, &msg.ConversationID, &msg.UserID, &msg.CreatedAt,
			&msg.Sender, &msg.Content, &msg.SentimentScore, &msg.SentimentLabel, &msg.AIProvider,
		); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

// SearchClinicalKnowledge performs RAG similarity search using pgvector
func (sc *SupabaseClient) SearchClinicalKnowledge(ctx context.Context, queryEmbedding []float64, limit int) ([]ClinicalKnowledge, error) {
	f32Embedding := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		f32Embedding[i] = float32(v)
	}
	pgvecEmbedding := pgvector.NewVector(f32Embedding)
	
	rows, err := sc.db.QueryContext(
		ctx,
		`SELECT id, title, content, source, category, embedding, is_approved_for_delivery
		 FROM public.clinical_knowledge
		 WHERE is_approved_for_delivery = true
		 ORDER BY embedding <-> $1
		 LIMIT $2`,
		pgvecEmbedding, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ClinicalKnowledge
	for rows.Next() {
		var k ClinicalKnowledge
		var embVector pgvector.Vector
		if err := rows.Scan(
			&k.ID, &k.Title, &k.Content, &k.Source, &k.Category,
			&embVector, &k.IsApprovedForDelivery,
		); err != nil {
			return nil, err
		}
		f32Slice := embVector.Slice()
		k.Embedding = make([]float64, len(f32Slice))
		for i, v := range f32Slice {
			k.Embedding[i] = float64(v)
		}
		results = append(results, k)
	}
	return results, rows.Err()
}

// GetCrisisKeywords retrieves all active crisis keywords
func (sc *SupabaseClient) GetCrisisKeywords(ctx context.Context) ([]CrisisKeyword, error) {
	rows, err := sc.db.QueryContext(
		ctx,
		`SELECT id, keyword, severity, category FROM public.crisis_keywords`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []CrisisKeyword
	for rows.Next() {
		var k CrisisKeyword
		if err := rows.Scan(&k.ID, &k.Keyword, &k.Severity, &k.Category); err != nil {
			return nil, err
		}
		keywords = append(keywords, k)
	}
	return keywords, rows.Err()
}

// GetClinicalAnchors retrieves all active clinical anchor prompts
func (sc *SupabaseClient) GetClinicalAnchors(ctx context.Context) ([]ClinicalAnchor, error) {
	rows, err := sc.db.QueryContext(
		ctx,
		`SELECT id, name, category, prompt_text, is_active, version FROM public.clinical_anchors WHERE is_active = true`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var anchors []ClinicalAnchor
	for rows.Next() {
		var a ClinicalAnchor
		if err := rows.Scan(&a.ID, &a.Name, &a.Category, &a.PromptText, &a.IsActive, &a.Version); err != nil {
			return nil, err
		}
		anchors = append(anchors, a)
	}
	return anchors, rows.Err()
}

// UpdateConversationCrisisFlag updates crisis status in conversation
func (sc *SupabaseClient) UpdateConversationCrisisFlag(ctx context.Context, conversationID string, detected bool, reason string) error {
	_, err := sc.db.ExecContext(
		ctx,
		`UPDATE public.conversations 
		 SET crisis_detected = $1, escalation_triggered = $2, escalation_reason = $3, escalation_timestamp = NOW()
		 WHERE id = $4`,
		detected, detected, reason, conversationID,
	)
	return err
}

// LogAPIUsage logs API usage for rate limiting
func (sc *SupabaseClient) LogAPIUsage(ctx context.Context, userID, endpoint string, provider string, tokensUsed int) error {
	_, err := sc.db.ExecContext(
		ctx,
		`INSERT INTO public.api_usage_logs (user_id, endpoint, ai_provider, tokens_consumed)
		 VALUES ($1, $2, $3, $4)`,
		userID, endpoint, provider, tokensUsed,
	)
	return err
}

// Close closes the database connection
func (sc *SupabaseClient) Close() error {
	return sc.db.Close()
}
