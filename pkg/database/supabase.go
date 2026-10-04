// api/database/supabase.go

package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// ErrEmailTaken is returned by CreateUser when the email is already registered
var ErrEmailTaken = errors.New("email already registered")

// SupabaseClient wraps database connections
type SupabaseClient struct {
	db *sql.DB
}

// NewSupabaseClient creates a new database connection
func NewSupabaseClient(config SupabaseConfig) (*SupabaseClient, error) {
	// PostgreSQL connection string
	var psqlInfo string
	if config.ConnString != "" {
		psqlInfo = config.ConnString
	} else {
		// For local development in Docker, URL should be the postgres service name
		psqlInfo = fmt.Sprintf(
			"host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
			config.URL,
			config.DBUser,
			config.DBPassword,
			config.DBName,
		)
	}

	psqlInfo = withPoolerSafeParams(psqlInfo)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Serverless-friendly pool: Supabase's transaction pooler does the heavy lifting
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(2 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &SupabaseClient{db: db}, nil
}

// withPoolerSafeParams makes lib/pq send each query in a single round trip.
// Supabase's transaction pooler (port 6543) can hand the second half of a
// two-step prepared statement to a different backend, which fails with
// "bind message supplies N parameters, but prepared statement requires M".
func withPoolerSafeParams(conn string) string {
	if strings.Contains(conn, "binary_parameters=") {
		return conn
	}
	if strings.HasPrefix(conn, "postgres://") || strings.HasPrefix(conn, "postgresql://") {
		if strings.Contains(conn, "?") {
			return conn + "&binary_parameters=yes"
		}
		return conn + "?binary_parameters=yes"
	}
	return conn + " binary_parameters=yes"
}

// GetUserByEmail retrieves a user by email (case-insensitive), including the password hash
func (sc *SupabaseClient) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := sc.db.QueryRowContext(
		ctx,
		`SELECT id, created_at, updated_at, email, full_name, COALESCE(encryption_pub_key, ''), COALESCE(privacy_mode, ''),
		        COALESCE(consent_therapeutic, false), COALESCE(consent_data_collection, false), COALESCE(consent_research, false),
		        last_login, COALESCE(is_active, true), password_hash
		 FROM public.users WHERE lower(email) = lower($1)`,
		email,
	).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Email, &user.FullName,
		&user.EncryptionPubKey, &user.PrivacyMode, &user.ConsentTherapeutic,
		&user.ConsentDataCollection, &user.ConsentResearch, &user.LastLogin, &user.IsActive,
		&user.PasswordHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("GetUserByEmail scan error: %w", err)
	}
	return &user, nil
}

// CreateUser inserts a new user with a bcrypt password hash.
// Returns ErrEmailTaken if the email is already registered.
func (sc *SupabaseClient) CreateUser(ctx context.Context, user *User) error {
	err := sc.db.QueryRowContext(
		ctx,
		`INSERT INTO public.users (id, email, full_name, privacy_mode, consent_therapeutic, consent_data_collection, consent_research, is_active, password_hash)
		 VALUES ($1, $2, $3, 'encrypted', false, false, false, true, $4)
		 RETURNING created_at, updated_at, privacy_mode`,
		user.ID, user.Email, user.FullName, user.PasswordHash,
	).Scan(&user.CreatedAt, &user.UpdatedAt, &user.PrivacyMode)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrEmailTaken
		}
		return fmt.Errorf("public.users insert failed: %w", err)
	}
	return nil
}

// TouchLastLogin records a successful login
func (sc *SupabaseClient) TouchLastLogin(ctx context.Context, userID string) error {
	_, err := sc.db.ExecContext(ctx, `UPDATE public.users SET last_login = NOW() WHERE id = $1`, userID)
	return err
}

// CreateConversation creates a new conversation owned by userID
func (sc *SupabaseClient) CreateConversation(ctx context.Context, userID, title string) (*Conversation, error) {
	var conv Conversation
	err := sc.db.QueryRowContext(
		ctx,
		`INSERT INTO public.conversations (user_id, status, ai_provider, title)
		 VALUES ($1, 'active', 'groq', $2)
		 RETURNING id, user_id, created_at, updated_at, status, ai_provider, title`,
		userID, title,
	).Scan(&conv.ID, &conv.UserID, &conv.CreatedAt, &conv.UpdatedAt, &conv.Status, &conv.AIProvider, &conv.Title)
	return &conv, err
}

// UserOwnsConversation reports whether conversationID exists and belongs to userID
func (sc *SupabaseClient) UserOwnsConversation(ctx context.Context, userID, conversationID string) (bool, error) {
	var exists bool
	err := sc.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM public.conversations WHERE id = $1 AND user_id = $2)`,
		conversationID, userID,
	).Scan(&exists)
	return exists, err
}

// StoreExchange saves the user's message and VANA's reply atomically and bumps the conversation
func (sc *SupabaseClient) StoreExchange(ctx context.Context, userMsg, aiMsg *Message) error {
	tx, err := sc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, m := range []*Message{userMsg, aiMsg} {
		if _, err := tx.ExecContext(
			ctx,
			// clock_timestamp(), not the column default now(): now() is fixed for the whole
			// transaction, which would give both messages the same time and an unstable order
			`INSERT INTO public.messages (conversation_id, user_id, sender, content, sentiment_score, sentiment_label,
			                              ai_model, ai_provider, tokens_used, response_time_ms, contains_crisis_keywords, flagged_for_review, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, clock_timestamp())`,
			m.ConversationID, m.UserID, m.Sender, m.Content, m.SentimentScore, m.SentimentLabel,
			m.AIModel, m.AIProvider, m.TokensUsed, m.ResponseTimeMs, m.ContainsCrisisKeywords, m.FlaggedForReview,
		); err != nil {
			return fmt.Errorf("insert %s message: %w", m.Sender, err)
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE public.conversations
		 SET message_count = COALESCE(message_count, 0) + 2,
		     ai_provider = COALESCE($2, ai_provider),
		     updated_at = NOW()
		 WHERE id = $1`,
		userMsg.ConversationID, aiMsg.AIProvider,
	); err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	return tx.Commit()
}

// ListConversations returns the user's conversations, most recently active first
func (sc *SupabaseClient) ListConversations(ctx context.Context, userID string, limit int) ([]ConversationSummary, error) {
	rows, err := sc.db.QueryContext(
		ctx,
		`SELECT id, COALESCE(title, 'Conversation'), created_at, updated_at, COALESCE(message_count, 0), COALESCE(crisis_detected, false)
		 FROM public.conversations
		 WHERE user_id = $1 AND status <> 'archived'
		 ORDER BY updated_at DESC
		 LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := []ConversationSummary{}
	for rows.Next() {
		var c ConversationSummary
		if err := rows.Scan(&c.ID, &c.Title, &c.CreatedAt, &c.UpdatedAt, &c.MessageCount, &c.CrisisDetected); err != nil {
			return nil, err
		}
		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
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

// CountClinicalKnowledge counts reviewed knowledge entries available for search
func (sc *SupabaseClient) CountClinicalKnowledge(ctx context.Context) (int, error) {
	var n int
	err := sc.db.QueryRowContext(ctx, `SELECT count(*) FROM public.clinical_knowledge WHERE is_approved_for_delivery = true`).Scan(&n)
	return n, err
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

// CheckRateLimit checks if a user has exceeded their message limits
func (sc *SupabaseClient) CheckRateLimit(ctx context.Context, userID string) error {
	var minuteCount, hourCount, dayCount int

	// Minute limit: 2
	err := sc.db.QueryRowContext(ctx, `SELECT count(*) FROM public.messages WHERE user_id = $1 AND sender = 'user' AND created_at >= NOW() - INTERVAL '1 minute'`, userID).Scan(&minuteCount)
	if err != nil {
		return fmt.Errorf("rate limit check error: %w", err)
	}
	if minuteCount >= 2 {
		return fmt.Errorf("rate limit exceeded: max 2 messages per minute")
	}

	// Hour limit: 20
	err = sc.db.QueryRowContext(ctx, `SELECT count(*) FROM public.messages WHERE user_id = $1 AND sender = 'user' AND created_at >= NOW() - INTERVAL '1 hour'`, userID).Scan(&hourCount)
	if err != nil {
		return fmt.Errorf("rate limit check error: %w", err)
	}
	if hourCount >= 20 {
		return fmt.Errorf("rate limit exceeded: max 20 messages per hour")
	}

	// Day limit: 50
	err = sc.db.QueryRowContext(ctx, `SELECT count(*) FROM public.messages WHERE user_id = $1 AND sender = 'user' AND created_at >= CURRENT_DATE`, userID).Scan(&dayCount)
	if err != nil {
		return fmt.Errorf("rate limit check error: %w", err)
	}
	if dayCount >= 50 {
		return fmt.Errorf("rate limit exceeded: max 50 messages per day")
	}

	return nil
}

// UserMemory is the remembered context about one person
type UserMemory struct {
	FirstName string
	Summary   string     // empty when nothing is remembered yet
	CoveredTo *time.Time // newest message already folded into Summary
}

// GetUserMemory returns the person's first name and remembered context
func (sc *SupabaseClient) GetUserMemory(ctx context.Context, userID string) (*UserMemory, error) {
	var fullName string
	var summary sql.NullString
	var covered sql.NullTime
	err := sc.db.QueryRowContext(ctx,
		`SELECT full_name, memory_summary, memory_updated_at FROM public.users WHERE id = $1`, userID,
	).Scan(&fullName, &summary, &covered)
	if err != nil {
		return nil, err
	}
	m := &UserMemory{Summary: summary.String}
	if f := strings.Fields(fullName); len(f) > 0 {
		m.FirstName = f[0]
	}
	if covered.Valid {
		m.CoveredTo = &covered.Time
	}
	return m, nil
}

// CountMessagesSince counts the user's messages newer than since (all messages when since is nil)
func (sc *SupabaseClient) CountMessagesSince(ctx context.Context, userID string, since *time.Time) (int, error) {
	var n int
	err := sc.db.QueryRowContext(ctx,
		`SELECT count(*) FROM public.messages WHERE user_id = $1 AND ($2::timestamptz IS NULL OR created_at > $2)`,
		userID, since,
	).Scan(&n)
	return n, err
}

// GetMessagesSince returns up to limit of the user's newest messages after since, oldest first
func (sc *SupabaseClient) GetMessagesSince(ctx context.Context, userID string, since *time.Time, limit int) ([]Message, error) {
	rows, err := sc.db.QueryContext(ctx,
		`SELECT id, conversation_id, created_at, sender, content FROM (
		   SELECT id, conversation_id, created_at, sender, content FROM public.messages
		   WHERE user_id = $1 AND ($2::timestamptz IS NULL OR created_at > $2)
		   ORDER BY created_at DESC LIMIT $3
		 ) recent ORDER BY created_at ASC`,
		userID, since, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.CreatedAt, &m.Sender, &m.Content); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// SaveUserMemory stores a new summary, unless another refresh already moved
// the memory past prevCoveredTo (then this one is stale and is dropped)
func (sc *SupabaseClient) SaveUserMemory(ctx context.Context, userID, summary string, prevCoveredTo *time.Time, coveredTo time.Time) (bool, error) {
	res, err := sc.db.ExecContext(ctx,
		`UPDATE public.users SET memory_summary = $2, memory_updated_at = $3
		 WHERE id = $1 AND memory_updated_at IS NOT DISTINCT FROM $4`,
		userID, summary, coveredTo, prevCoveredTo,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n == 1, nil
}

// ClearUserMemory forgets everything remembered about the user. Messages
// already written stay out of memory; only newer ones are remembered.
func (sc *SupabaseClient) ClearUserMemory(ctx context.Context, userID string) error {
	_, err := sc.db.ExecContext(ctx,
		`UPDATE public.users SET memory_summary = NULL,
		   memory_updated_at = COALESCE((SELECT max(created_at) FROM public.messages WHERE user_id = $1), now())
		 WHERE id = $1`, userID)
	return err
}

// Close closes the database connection
func (sc *SupabaseClient) Close() error {
	return sc.db.Close()
}
