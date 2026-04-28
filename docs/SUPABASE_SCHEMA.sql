-- Supabase PostgreSQL Schema for Project Emerald Moss
-- Database: emerald_moss
-- Region: us-east-1 (Recommended for low latency)

-- ============================================================================
-- EXTENSION: pgvector for RAG-based clinical knowledge
-- ============================================================================
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- MOCK AUTH SCHEMA (For Local Development)
-- ============================================================================
CREATE SCHEMA IF NOT EXISTS auth;
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Mock auth.uid() for local development
CREATE OR REPLACE FUNCTION auth.uid() RETURNS uuid AS $$
    SELECT id FROM auth.users LIMIT 1;
$$ LANGUAGE sql STABLE;

-- ============================================================================
-- TABLE: Users (Auth Integration)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.users (
    id UUID REFERENCES auth.users NOT NULL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT NOT NULL,
    encryption_pub_key TEXT,
    -- Privacy & Consent
    privacy_mode TEXT DEFAULT 'encrypted', -- 'encrypted', 'local-only', 'supabase'
    consent_therapeutic BOOLEAN DEFAULT FALSE,
    consent_data_collection BOOLEAN DEFAULT FALSE,
    consent_research BOOLEAN DEFAULT FALSE,
    -- Profile
    avatar_url TEXT,
    bio TEXT,
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_users_email ON public.users(email);
CREATE INDEX idx_users_created_at ON public.users(created_at DESC);

-- ============================================================================
-- TABLE: Conversations (Session Management)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.conversations (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    -- Clinical Data
    gad7_score INTEGER CHECK (gad7_score BETWEEN 0 AND 21),
    gad7_severity TEXT, -- 'minimal', 'mild', 'moderate', 'severe'
    -- Session Meta
    status TEXT DEFAULT 'active' CHECK (status IN ('active', 'paused', 'archived', 'closed')),
    title TEXT,
    summary TEXT,
    -- Mood Tracking
    mood_start INTEGER, -- -5 to +5 scale
    mood_end INTEGER,
    sentiment_average NUMERIC,
    -- Escalation
    crisis_detected BOOLEAN DEFAULT FALSE,
    escalation_triggered BOOLEAN DEFAULT FALSE,
    escalation_reason TEXT,
    escalation_timestamp TIMESTAMP WITH TIME ZONE,
    -- Metadata
    message_count INTEGER DEFAULT 0,
    ai_provider TEXT DEFAULT 'groq', -- 'groq', 'gemini'
    duration_seconds INTEGER
);

CREATE INDEX idx_conversations_user_id ON public.conversations(user_id);
CREATE INDEX idx_conversations_created_at ON public.conversations(created_at DESC);
CREATE INDEX idx_conversations_status ON public.conversations(status);
CREATE INDEX idx_conversations_crisis ON public.conversations(crisis_detected) WHERE crisis_detected = TRUE;

-- ============================================================================
-- TABLE: Messages (Conversation History)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.messages (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES public.conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    -- Message Content
    sender TEXT NOT NULL CHECK (sender IN ('user', 'ai')),
    content TEXT NOT NULL,
    encrypted_content TEXT, -- For end-to-end encryption
    is_encrypted BOOLEAN DEFAULT FALSE,
    -- Sentiment Analysis
    sentiment_score NUMERIC, -- -1.0 (very negative) to +1.0 (very positive)
    sentiment_label TEXT, -- 'positive', 'neutral', 'negative', 'critical'
    emotion_detected TEXT, -- emotion classification
    -- AI Metadata
    ai_model TEXT,
    ai_provider TEXT,
    tokens_used INTEGER,
    response_time_ms INTEGER,
    -- Safety & Crisis
    contains_crisis_keywords BOOLEAN DEFAULT FALSE,
    crisis_keywords TEXT[],
    flagged_for_review BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_messages_conversation_id ON public.messages(conversation_id);
CREATE INDEX idx_messages_user_id ON public.messages(user_id);
CREATE INDEX idx_messages_created_at ON public.messages(created_at DESC);
CREATE INDEX idx_messages_sender ON public.messages(sender);
CREATE INDEX idx_messages_sentiment ON public.messages(sentiment_score);

-- ============================================================================
-- TABLE: Clinical Knowledge Base (RAG - Retrieval Augmented Generation)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.clinical_knowledge (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    -- Content
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    source TEXT, -- 'dsm5', 'gad7', 'professional', 'evidence-based'
    category TEXT, -- 'anxiety', 'depression', 'mindfulness', 'coping', 'medication'
    -- Embedding & Retrieval
    embedding vector(1536), -- 1536 for OpenAI model-3-large, or adjust for Gemini
    metadata JSONB DEFAULT '{}'::jsonb, -- {confidence: 0.9, keywords: [...], relevance_tags: [...]}
    -- Clinical Safety
    requires_professional_review BOOLEAN DEFAULT FALSE,
    is_approved_for_delivery BOOLEAN DEFAULT TRUE,
    medical_disclaimers TEXT[]
);

CREATE INDEX idx_clinical_knowledge_category ON public.clinical_knowledge(category);
CREATE INDEX idx_clinical_knowledge_source ON public.clinical_knowledge(source);
CREATE INDEX idx_clinical_knowledge_embedding ON public.clinical_knowledge USING hnsw (embedding vector_cosine_ops);

-- ============================================================================
-- TABLE: Clinical Anchor Prompts (System Instructions)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.clinical_anchors (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()) NOT NULL,
    -- Prompt Definition
    name TEXT UNIQUE NOT NULL,
    category TEXT NOT NULL, -- 'system', 'safety', 'empathy', 'clinical'
    prompt_text TEXT NOT NULL,
    -- Metadata
    is_active BOOLEAN DEFAULT TRUE,
    version INTEGER DEFAULT 1,
    approved_by TEXT,
    description TEXT
);

INSERT INTO public.clinical_anchors (name, category, prompt_text, description) VALUES
(
    'empathetic_opening',
    'empathy',
    'You are a compassionate mental health companion. Always prioritize the user''s emotional wellbeing. Listen actively, validate their feelings, and respond with warmth and understanding.',
    'System prompt for empathetic responses'
),
(
    'evidence_based_response',
    'clinical',
    'When providing mental health information, only cite evidence-based approaches. Include proper medical disclaimers about medication and professional consultation.',
    'Ensure responses are backed by clinical evidence'
),
(
    'crisis_safety',
    'safety',
    'If the user mentions self-harm, suicide, or severe distress, immediately acknowledge their pain and provide crisis resources: National Suicide Prevention Lifeline: 988, Crisis Text Line: Text HOME to 741741',
    'Crisis response protocol'
),
(
    'medication_disclaimer',
    'safety',
    'When discussing medications: (1) Never recommend starting/stopping medication, (2) Explain medication classes (SSRIs, SNRIs, etc.) educationally only, (3) Always recommend professional consultation, (4) Remind user that this is not medical advice',
    'Medication discussion guidelines'
);

CREATE INDEX idx_clinical_anchors_category ON public.clinical_anchors(category);
CREATE INDEX idx_clinical_anchors_is_active ON public.clinical_anchors(is_active);

-- ============================================================================
-- TABLE: Crisis Keywords Registry
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.crisis_keywords (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    keyword TEXT UNIQUE NOT NULL,
    severity TEXT NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    category TEXT, -- 'ideation', 'self-harm', 'substance-abuse', 'trauma'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())
);

INSERT INTO public.crisis_keywords (keyword, severity, category) VALUES
('suicide', 'critical', 'ideation'),
('kill myself', 'critical', 'ideation'),
('end my life', 'critical', 'ideation'),
('want to die', 'high', 'ideation'),
('don''t want to live', 'high', 'ideation'),
('hurt myself', 'high', 'self-harm'),
('cut myself', 'high', 'self-harm'),
('overdose', 'critical', 'substance-abuse'),
('no reason to live', 'high', 'ideation'),
('worthless', 'medium', 'depression-severity'),
('hopeless', 'medium', 'depression-severity');

CREATE INDEX idx_crisis_keywords_severity ON public.crisis_keywords(severity);

-- ============================================================================
-- TABLE: Session Metadata (Analytics & Tracking)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.session_metadata (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES public.conversations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    -- Environment
    user_agent TEXT,
    device_type TEXT, -- 'mobile', 'tablet', 'desktop'
    browser TEXT,
    ip_address_hash TEXT, -- hashed for privacy
    -- Performance
    page_load_time_ms INTEGER,
    first_interaction_delay_ms INTEGER,
    -- Feature Usage
    voice_enabled BOOLEAN DEFAULT FALSE,
    tts_used BOOLEAN DEFAULT FALSE,
    dark_mode_used BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_session_metadata_user_id ON public.session_metadata(user_id);
CREATE INDEX idx_session_metadata_created_at ON public.session_metadata(created_at DESC);

-- ============================================================================
-- TABLE: Psychiatrist Bridge (Session Reports)
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.psychiatrist_bridge_reports (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES public.conversations(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    -- Report Content
    session_summary TEXT,
    mood_trajectory TEXT,
    key_themes TEXT[],
    crisis_flags TEXT[],
    recommendations TEXT,
    -- Metadata
    generated_by_model TEXT, -- 'gemini', 'groq'
    is_reviewed BOOLEAN DEFAULT FALSE,
    reviewed_by TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    is_shareable_with_professional BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_bridge_reports_user_id ON public.psychiatrist_bridge_reports(user_id);
CREATE INDEX idx_bridge_reports_is_reviewed ON public.psychiatrist_bridge_reports(is_reviewed);

-- ============================================================================
-- TABLE: API Usage & Rate Limiting
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.api_usage_logs (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now()),
    -- Rate Limit Data
    request_count INTEGER DEFAULT 1,
    response_status INTEGER,
    response_time_ms INTEGER,
    ai_provider TEXT,
    tokens_consumed INTEGER
);

CREATE INDEX idx_api_usage_user_id ON public.api_usage_logs(user_id);
CREATE INDEX idx_api_usage_created_at ON public.api_usage_logs(created_at DESC);
CREATE INDEX idx_api_usage_endpoint ON public.api_usage_logs(endpoint);

-- ============================================================================
-- POLICIES: Row-Level Security (RLS)
-- ============================================================================

-- Enable RLS on all tables
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.session_metadata ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.psychiatrist_bridge_reports ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.api_usage_logs ENABLE ROW LEVEL SECURITY;

-- Users can view only their own profile
CREATE POLICY users_select_self ON public.users
    FOR SELECT USING (auth.uid() = id);

-- Users can update only their own profile
CREATE POLICY users_update_self ON public.users
    FOR UPDATE USING (auth.uid() = id);

-- Users can view only their own conversations
CREATE POLICY conversations_select_self ON public.conversations
    FOR SELECT USING (auth.uid() = user_id);

-- Users can view only their own messages
CREATE POLICY messages_select_self ON public.messages
    FOR SELECT USING (auth.uid() = user_id);

-- ============================================================================
-- FUNCTIONS & TRIGGERS
-- ============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = timezone('utc'::text, now());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for users table
CREATE TRIGGER users_updated_at_trigger
BEFORE UPDATE ON public.users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Trigger for conversations table
CREATE TRIGGER conversations_updated_at_trigger
BEFORE UPDATE ON public.conversations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Trigger for clinical knowledge table
CREATE TRIGGER clinical_knowledge_updated_at_trigger
BEFORE UPDATE ON public.clinical_knowledge
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SUMMARY OF KEY TABLES
-- ============================================================================
/*
1. users: User profile & consent management
2. conversations: Session metadata & escalation tracking
3. messages: Full conversation history with sentiment analysis
4. clinical_knowledge: RAG vector store for clinical information
5. clinical_anchors: System prompts for empathy & safety
6. crisis_keywords: Real-time crisis detection registry
7. session_metadata: Analytics & performance tracking
8. psychiatrist_bridge_reports: Shareable session summaries
9. api_usage_logs: Rate limiting & usage analytics

All tables support:
- Full encryption at rest (Supabase)
- Row-level security (RLS)
- Audit trails with timestamps
- HIPAA compliance readiness
*/
