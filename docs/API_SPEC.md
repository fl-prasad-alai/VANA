# Project Emerald Moss - API Specification

## Base URL
- **Development**: `http://localhost:3000/api`
- **Production**: `https://your-domain.vercel.app/api`

## Authentication
All endpoints (except `/auth/*`) require JWT token in Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

---

## Authentication Endpoints

### POST /auth/register
Register a new user account.

**Request**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe"
}
```

**Response (201)**
```json
{
  "success": true,
  "message": "User registered successfully",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "John Doe",
    "created_at": "2026-04-28T10:00:00Z"
  }
}
```

**Error Responses**
- 400: Invalid email format
- 409: Email already exists
- 422: Password doesn't meet security requirements

---

### POST /auth/login
Authenticate user and receive JWT token.

**Request**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200)**
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "full_name": "John Doe"
  }
}
```

**Error Responses**
- 401: Invalid credentials
- 404: User not found

---

## Conversation Endpoints

### POST /conversations
Create a new conversation session.

**Request**
```json
{
  "title": "Anxiety Discussion"
}
```

**Response (201)**
```json
{
  "success": true,
  "conversation": {
    "id": "uuid",
    "user_id": "uuid",
    "status": "active",
    "created_at": "2026-04-28T10:00:00Z",
    "ai_provider": "groq"
  }
}
```

---

### GET /conversations/{id}
Retrieve conversation details.

**Response (200)**
```json
{
  "success": true,
  "conversation": {
    "id": "uuid",
    "user_id": "uuid",
    "status": "active",
    "gad7_score": 12,
    "gad7_severity": "mild",
    "message_count": 15,
    "mood_start": 2,
    "mood_end": 4,
    "sentiment_average": 0.15,
    "crisis_detected": false,
    "created_at": "2026-04-28T10:00:00Z",
    "updated_at": "2026-04-28T10:15:00Z"
  }
}
```

---

### GET /conversations/{id}/messages
Retrieve all messages in a conversation.

**Query Parameters**
- `limit`: Number of messages (default: 50)
- `offset`: Pagination offset (default: 0)

**Response (200)**
```json
{
  "success": true,
  "messages": [
    {
      "id": "uuid",
      "conversation_id": "uuid",
      "sender": "user",
      "content": "I've been feeling anxious lately",
      "sentiment_score": -0.3,
      "sentiment_label": "negative",
      "created_at": "2026-04-28T10:01:00Z"
    },
    {
      "id": "uuid",
      "conversation_id": "uuid",
      "sender": "ai",
      "content": "I'm sorry to hear that...",
      "sentiment_score": 0.2,
      "sentiment_label": "neutral",
      "ai_provider": "groq",
      "tokens_used": 50,
      "created_at": "2026-04-28T10:02:00Z"
    }
  ],
  "total": 45,
  "page": 1,
  "limit": 50
}
```

---

## Chat Endpoints

### POST /chat
Send a message and receive AI response.

**Request**
```json
{
  "conversation_id": "uuid",
  "message": "I'm feeling overwhelmed at work",
  "is_voice": false
}
```

**Response (200)**
```json
{
  "success": true,
  "is_crisis": false,
  "response": {
    "id": "uuid",
    "conversation_id": "uuid",
    "response": "That sounds really challenging. Work stress can be overwhelming. Let me help you think through some coping strategies.",
    "sentiment": 0.15,
    "provider_id": "groq",
    "tokens_used": 120,
    "timestamp": "2026-04-28T10:02:00Z"
  }
}
```

**Crisis Response (200)**
```json
{
  "success": true,
  "is_crisis": true,
  "message": "I hear that you're going through a very difficult time.",
  "crisis_resources": {
    "US Suicide Prevention Lifeline": "Call 988",
    "Crisis Text Line": "Text HOME to 741741",
    "Emergency": "Call 911"
  }
}
```

**Error Responses**
- 400: Invalid conversation ID
- 429: Rate limit exceeded
- 503: AI provider unavailable

---

## Triage Endpoints

### POST /triage
Submit GAD-7 intake assessment.

**Request**
```json
{
  "conversation_id": "uuid",
  "responses": [1, 0, 2, 1, 2, 1, 3]
}
```

Response values: 0-3 for each question
- 0: Not at all
- 1: Several days
- 2: More than half the days
- 3: Nearly every day

**Response (200)**
```json
{
  "success": true,
  "gad7_score": 10,
  "severity": "mild",
  "recommendation": "Your responses suggest mild anxiety. Consider self-help strategies or speaking with a professional.",
  "crisis_risk": false,
  "next_steps": [
    "Practice daily mindfulness",
    "Maintain regular sleep schedule",
    "Consider speaking with a therapist"
  ]
}
```

---

## Report Endpoints

### GET /reports/psychiatrist-bridge/{conversation_id}
Get shareable session summary for mental health professional.

**Response (200)**
```json
{
  "success": true,
  "report": {
    "id": "uuid",
    "conversation_id": "uuid",
    "session_summary": "User discussed work-related anxiety for 23 minutes...",
    "mood_trajectory": "Started at 2/10, improved to 5/10",
    "key_themes": ["work stress", "sleep deprivation", "low self-confidence"],
    "crisis_flags": [],
    "recommendations": [
      "Consider CBT for work anxiety",
      "Sleep hygiene recommendations"
    ],
    "is_shareable_with_professional": true,
    "generated_at": "2026-04-28T10:30:00Z"
  }
}
```

---

## Health & Admin Endpoints

### GET /health
Check API health status.

**Response (200)**
```json
{
  "status": "healthy",
  "timestamp": "2026-04-28T10:00:00Z",
  "services": {
    "database": "connected",
    "groq": "available",
    "gemini": "available"
  }
}
```

---

### GET /stats
Get API usage statistics (Admin only).

**Response (200)**
```json
{
  "success": true,
  "stats": {
    "total_users": 1250,
    "active_conversations": 87,
    "total_messages": 45230,
    "crisis_detections_today": 2,
    "avg_response_time_ms": 450,
    "groq_usage": {
      "requests": 3450,
      "tokens": 125000,
      "rpm_used": 18
    },
    "gemini_usage": {
      "requests": 890,
      "tokens": 45000,
      "rpm_used": 8
    }
  }
}
```

---

## Error Handling

All error responses follow this format:

```json
{
  "success": false,
  "message": "Error description",
  "code": "ERROR_CODE",
  "details": "Additional error details"
}
```

### HTTP Status Codes
- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 409: Conflict
- 429: Too Many Requests
- 500: Internal Server Error
- 503: Service Unavailable

### Error Codes
- `INVALID_INPUT`: Request body validation failed
- `INVALID_EMAIL`: Email format invalid
- `WEAK_PASSWORD`: Password doesn't meet requirements
- `USER_EXISTS`: User already registered
- `INVALID_CREDENTIALS`: Wrong email/password
- `UNAUTHORIZED`: Missing or invalid JWT
- `CONVERSATION_NOT_FOUND`: Conversation ID doesn't exist
- `CRISIS_DETECTED`: Crisis keywords detected
- `RATE_LIMIT_EXCEEDED`: Too many requests from this user/IP
- `PROVIDER_ERROR`: AI provider returned error
- `DATABASE_ERROR`: Database operation failed

---

## Rate Limiting

Rate limiting is per-user based on JWT:
- **Standard Tier**: 100 requests/minute
- **Premium Tier**: 1000 requests/minute

**Headers**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 87
X-RateLimit-Reset: 1619654400
```

---

## Data Models

### User
```typescript
{
  id: UUID;
  email: string;
  full_name: string;
  created_at: ISO8601;
  updated_at: ISO8601;
  privacy_mode: 'encrypted' | 'local-only' | 'supabase';
  consent_therapeutic: boolean;
  is_active: boolean;
}
```

### Conversation
```typescript
{
  id: UUID;
  user_id: UUID;
  status: 'active' | 'paused' | 'archived' | 'closed';
  gad7_score?: number; // 0-21
  gad7_severity?: 'minimal' | 'mild' | 'moderate' | 'severe';
  title?: string;
  summary?: string;
  mood_start?: number; // -5 to +5
  mood_end?: number;
  sentiment_average?: number;
  crisis_detected: boolean;
  escalation_triggered: boolean;
  message_count: number;
  ai_provider: 'groq' | 'gemini';
  created_at: ISO8601;
  updated_at: ISO8601;
}
```

### Message
```typescript
{
  id: UUID;
  conversation_id: UUID;
  sender: 'user' | 'ai';
  content: string;
  sentiment_score?: number; // -1.0 to 1.0
  sentiment_label?: 'positive' | 'neutral' | 'negative' | 'critical';
  emotion_detected?: string;
  ai_provider?: 'groq' | 'gemini';
  tokens_used?: number;
  response_time_ms?: number;
  contains_crisis_keywords: boolean;
  created_at: ISO8601;
}
```

---

## Webhook Events (Future)
- `conversation.created`
- `message.sent`
- `crisis.detected`
- `triage.completed`

---

**API Version**: 1.0.0
**Last Updated**: April 28, 2026
**Status**: Production
