# PROJECT EMERALD MOSS - Workspace Architecture & Implementation Plan

## Mission Statement
A revolutionary, voice-first Biophilic Digital Triage System bridging the "Clinical Chasm" by providing instant, nature-inspired mental health support at **zero cost** using high-efficiency Go/Angular/Gemini stack.

---

## 1. PROJECT STRUCTURE (Monorepo - Vercel Compatible)

```
emerald-moss/
├── angular.json                          # Angular configuration
├── package.json                          # Frontend dependencies + scripts
├── tailwind.config.js                    # Emerald Moss custom palette
├── tsconfig.json                         # TypeScript root config
├── tsconfig.app.json                     # Angular app config
├── tsconfig.spec.json                    # Testing config
├── vercel.json                           # Vercel serverless routing
├── docker-compose.yml                    # Local dev environment
├── Dockerfile                            # Multi-stage build for Go + Angular
├── .env.example                          # Environment variables template
├── .gitignore
│
├── api/                                  # Go Serverless Functions (Vercel /api routes)
│   ├── go.mod                           # Go module definition
│   ├── go.sum                           # Go dependencies lock
│   ├── index.go                         # Main Vercel entry point
│   ├── handlers/
│   │   ├── auth.go                      # Login/Register handlers
│   │   ├── chat.go                      # Chat message handler
│   │   ├── triage.go                    # Intake triage handler
│   │   └── health.go                    # Health check endpoint
│   ├── middleware/
│   │   ├── crisis.go                    # Crisis Interceptor (Safety Filter)
│   │   ├── auth.go                      # JWT authentication middleware
│   │   └── cors.go                      # CORS middleware
│   ├── orchestration/
│   │   ├── provider.go                  # Multi-Provider interface & balancer
│   │   ├── groq.go                      # Groq API client (Tier 1: Speed)
│   │   ├── gemini.go                    # Gemini API client (Tier 2: Logic)
│   │   └── fallback.go                  # Fallback & retry logic
│   ├── database/
│   │   ├── supabase.go                  # Supabase client + pgvector RAG
│   │   └── models.go                    # Data models
│   ├── services/
│   │   ├── triage.go                    # GAD-7 triage logic
│   │   ├── clinical.go                  # Clinical anchor prompts & safety
│   │   └── encryption.go                # Local-first encryption helpers
│   └── utils/
│       ├── logger.go                    # Structured logging
│       └── errors.go                    # Error handling
│
├── src/                                 # Angular 21 Application
│   ├── index.html                       # Main entry point
│   ├── main.ts                          # Bootstrap
│   ├── styles.scss                      # Global styles
│   ├── theme.scss                       # Emerald Moss palette & animations
│   │
│   ├── app/
│   │   ├── app.ts                       # Root component
│   │   ├── app.routes.ts                # Route definitions
│   │   ├── app.scss                     # App-level styles
│   │   │
│   │   ├── pages/
│   │   │   ├── login/
│   │   │   │   ├── login.component.ts
│   │   │   │   └── login.component.scss
│   │   │   ├── register/
│   │   │   │   ├── register.component.ts
│   │   │   │   └── register.component.scss
│   │   │   ├── dashboard/
│   │   │   │   ├── dashboard.component.ts
│   │   │   │   └── dashboard.component.scss
│   │   │   └── not-found/
│   │   │       └── not-found.component.ts
│   │   │
│   │   ├── components/
│   │   │   ├── breath-sync/             # Pulsing voice indicator
│   │   │   │   └── breath-sync.component.ts
│   │   │   ├── forest-guide/            # Main chat interface (Glassmorphism)
│   │   │   │   ├── forest-guide.component.ts
│   │   │   │   └── forest-guide.component.scss
│   │   │   ├── intake-form/             # GAD-7 triage form
│   │   │   │   ├── intake-form.component.ts
│   │   │   │   └── intake-form.component.scss
│   │   │   ├── sentiment-indicator/     # Mood-aware UI feedback
│   │   │   │   └── sentiment-indicator.component.ts
│   │   │   ├── navbar/
│   │   │   │   └── navbar.component.ts
│   │   │   └── loading-spinner/
│   │   │       └── loading-spinner.component.ts
│   │   │
│   │   ├── services/
│   │   │   ├── auth.service.ts          # Authentication logic
│   │   │   ├── chat.service.ts          # API communication with Go backend
│   │   │   ├── voice.service.ts         # Web Speech API (STT/TTS)
│   │   │   ├── storage.service.ts       # Local storage + encryption
│   │   │   ├── sentiment.service.ts     # Emotion detection & background shift
│   │   │   └── logger.service.ts        # Client-side logging
│   │   │
│   │   ├── guards/
│   │   │   ├── auth.guard.ts            # Route protection
│   │   │   └── unsaved-changes.guard.ts
│   │   │
│   │   ├── interceptors/
│   │   │   ├── auth.interceptor.ts      # Attach JWT to requests
│   │   │   ├── error.interceptor.ts     # Global error handling
│   │   │   └── loading.interceptor.ts   # Show/hide loading
│   │   │
│   │   ├── models/
│   │   │   ├── auth.models.ts
│   │   │   ├── chat.models.ts
│   │   │   ├── user.models.ts
│   │   │   └── api.models.ts
│   │   │
│   │   └── utils/
│   │       ├── validators.ts            # Custom validators
│   │       ├── emotion.ts               # Sentiment scoring
│   │       └── constants.ts             # App constants
│   │
│   └── assets/
│       ├── icons/
│       ├── images/
│       └── animations.json
│
├── public/                              # Static assets
│   └── favicon.ico
│
└── docs/
    ├── API_SPEC.md                      # Go API documentation
    ├── DEPLOYMENT.md                    # Vercel deployment guide
    └── SECURITY.md                      # Security & encryption specs
```

---

## 2. DATA FLOW ARCHITECTURE

### Frontend → Backend → AI Provider Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         USER INTERFACE                              │
│  (Angular 21 + Tailwind + Biophilic Design)                        │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │ Login/Register → Dashboard → Forest Guide Chat Interface    │  │
│  │                   ↓                                          │  │
│  │              Voice Service (Web Speech API)                 │  │
│  │                   ↓                                          │  │
│  │         Chat Service (HTTP + JWT Token)                    │  │
│  └──────────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────────┘
                             │ (Encrypted Request)
                             ↓
┌────────────────────────────────────────────────────────────────────┐
│                      GO SERVERLESS BACKEND                         │
│  (Vercel Functions)                                               │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │ 1. Auth Handlers (JWT Issue/Verify)                         │ │
│  │ 2. Crisis Middleware (Regex Interceptor)                    │ │
│  │    ↓ DETECT CRISIS KEYWORDS → ROUTE TO HELPLINE             │ │
│  │ 3. Triage Handler (GAD-7 Scoring)                           │ │
│  │ 4. Chat Handler (Message Processing)                        │ │
│  │    ↓ Route to AI Provider                                   │ │
│  │                                                              │ │
│  │ ┌─────────────────────────────────────────────────────────┐ │
│  │ │       MULTI-PROVIDER LOAD BALANCER                     │ │
│  │ │  ┌────────────────────────────────────────────────────┐ │ │
│  │ │  │ Tier 1 (Speed): Groq Llama 3.1 8B                │ │ │
│  │ │  │ • 500+ tokens/sec                                │ │ │
│  │ │  │ • 30 RPM limit                                   │ │ │
│  │ │  │ • Sub-second cold starts                         │ │ │
│  │ │  └────────────────────────────────────────────────────┘ │ │
│  │ │  ↓ (If timeout or error)                                 │ │
│  │ │  ┌────────────────────────────────────────────────────┐ │ │
│  │ │  │ Tier 2 (Logic): Gemini 1.5 Flash                 │ │ │
│  │ │  │ • Clinical synthesis                             │ │ │
│  │ │  │ • Deep memory (RAG)                              │ │ │
│  │ │  │ • 15 RPM limit                                   │ │ │
│  │ │  └────────────────────────────────────────────────────┘ │ │
│  │ │                                                          │ │
│  │ │ • Clinical Anchor Prompts Applied                       │ │
│  │ │ • Medication Class Disclaimers                          │ │
│  │ │ • Evidence-Based Responses                              │ │
│  │ └─────────────────────────────────────────────────────────┘ │
│  │                                                              │ │
│  │ 5. Database: Supabase PostgreSQL + PgVector               │ │
│  │    ├─ RAG Knowledge Base (Clinical Vectors)               │ │
│  │    ├─ User Sessions (Encrypted)                           │ │
│  │    └─ Conversation History + Sentiment Scores             │ │
│  │                                                              │ │
│  │ 6. Response Builder (Psychiatrist Bridge Report)          │ │
│  │    ├─ Session Summary                                     │ │
│  │    ├─ Mood Trends                                         │ │
│  │    └─ Escalation Flags                                    │ │
│  └──────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────┘
                             ↑
                    (JSON Response)
                             │
┌────────────────────────────┴───────────────────────────────────────┐
│                    RESPONSE HANDLING                               │
│  ├─ Parse & Validate                                             │
│  ├─ Update Sentiment-Aware Background                            │
│  ├─ Display Glassmorphism Chat Bubble                            │
│  ├─ Text-to-Speech (Optional, Native API)                       │
│  └─ Store Locally + Encrypt                                     │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 3. AUTHENTICATION FLOW

### Login/Register → JWT → Protected Routes

```
User → Login Component
    ↓ (Email + Password)
Go Auth Handler
    ↓ (Hash check in Supabase)
    ├─ Valid: Issue JWT (Exp: 7 days)
    └─ Invalid: Return 401 + Error message
    ↓
Store JWT in LocalStorage + SessionStorage
    ↓
Auth Guard validates on route navigation
    ↓
Auth Interceptor attaches JWT to all requests
    ↓
Go Auth Middleware verifies JWT on each request
```

---

## 4. CRISIS DETECTION PROTOCOL

### Real-Time Safety Interceptor

```go
Regex Patterns:
├─ Ideation: "suicide|kill myself|end my life|want to die"
├─ Self-Harm: "hurt myself|cut|overdose"
├─ Immediate: "right now|today|tonight"
└─ Severe: "noose|gun|pills"

Detection Flow:
User Message → Crisis Middleware (Go)
    ├─ Match Against Regex
    ├─ CRISIS DETECTED:
    │   ├─ Bypass AI entirely
    │   ├─ Log incident (Encrypted)
    │   ├─ Return IMMEDIATE helpline:
    │   │   "Call 988 (US): Suicide & Crisis Lifeline"
    │   │   "Text HOME to 741741"
    │   │   "International: findahelpline.com"
    │   └─ Optional: Escalate to mental health professional
    └─ No match: Continue to AI
```

---

## 5. TECHNICAL SPECIFICATIONS

### Frontend Stack
- **Framework**: Angular 21
- **Styling**: SCSS + Tailwind CSS 4
- **Component Architecture**: Standalone components
- **State Management**: RxJS + Angular Services
- **Voice**: Native Web Speech API (no external dependencies)
- **HTTP**: Angular HttpClient with Interceptors

### Backend Stack
- **Runtime**: Go 1.22+ (Vercel Serverless)
- **AI Providers**:
  - Groq (Tier 1): `groq-cloud/llama-3.1-8b-instant`
  - Gemini (Tier 2): `gemini-1.5-flash`
- **Database**: Supabase PostgreSQL + PgVector
- **Authentication**: JWT (HS256 or RS256)
- **Logging**: Structured JSON logging

### Deployment
- **Frontend**: Vercel (via `ng build` → `/dist`)
- **Backend**: Vercel Serverless Functions (`/api`)
- **Database**: Supabase Hosted (PostgreSQL)
- **Environment**: Docker for local development

---

## 6. IMPLEMENTATION PHASES

### Phase 1: Foundation (Week 1-2)
- [x] Project structure scaffolding
- [x] Tailwind + Emerald Moss theme finalization
- [ ] Login/Register UI components
- [ ] JWT authentication setup (Go)
- [ ] Supabase schema + migrations

### Phase 2: Voice & Chat (Week 3-4)
- [ ] Web Speech API integration
- [ ] Chat Service + WebSocket (optional)
- [ ] Forest Guide chat interface (Glassmorphism)
- [ ] Message display + animations

### Phase 3: AI Integration (Week 5-6)
- [ ] Multi-provider orchestration (Groq + Gemini)
- [ ] Crisis Interceptor middleware
- [ ] Clinical Anchor prompts
- [ ] RAG via PgVector

### Phase 4: Polish & Deployment (Week 7-8)
- [ ] Sentiment-aware background shifting
- [ ] Psychiatrist Bridge reports
- [ ] Docker Compose setup
- [ ] Vercel deployment
- [ ] Testing + security audit

---

## 7. ENVIRONMENT VARIABLES (REQUIRED)

```env
# Authentication
JWT_SECRET=your_super_secret_key_here
JWT_EXPIRATION=604800  # 7 days in seconds

# Supabase
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key

# Groq (Tier 1)
GROQ_API_KEY=gsk_your_groq_key
GROQ_MODEL=groq-cloud/llama-3.1-8b-instant
GROQ_RPM_LIMIT=30

# Gemini (Tier 2)
GEMINI_API_KEY=your_gemini_api_key
GEMINI_MODEL=gemini-1.5-flash
GEMINI_RPM_LIMIT=15

# Frontend
VITE_API_URL=http://localhost:3000/api
VITE_ENV=development

# Logging
LOG_LEVEL=info
```

---

## 8. SECURITY & PRIVACY

### Data Encryption
- **In Transit**: HTTPS (Vercel + Supabase TLS)
- **At Rest**: Supabase server-side encryption
- **Client-Side**: Optional AES-256 for sensitive messages

### Access Control
- JWT-based authentication
- Role-based access control (RBAC)
- Rate limiting on API endpoints
- Crisis data audit trail (encrypted)

### Compliance
- HIPAA-ready (encrypted conversations)
- GDPR compliance (user data deletion endpoints)
- No third-party analytics on sensitive data

---

## 9. DEPLOYMENT CHECKLIST

- [ ] Environment variables in Vercel dashboard
- [ ] Go functions deployed to `/api`
- [ ] Angular build output in `/public`
- [ ] Supabase tables migrated + indexes created
- [ ] Crisis keywords tested
- [ ] JWT token generation tested
- [ ] Groq + Gemini APIs verified working
- [ ] SSL certificate auto-renewal
- [ ] Monitoring + alerting setup
- [ ] Security headers configured
- [ ] Rate limiting configured
- [ ] CORS policies verified

---

## 10. KEY CONTACTS & RESOURCES

- **Groq Docs**: https://console.groq.com/docs
- **Google Gemini**: https://makersuite.google.com
- **Supabase**: https://supabase.com/docs
- **Angular 21**: https://angular.io/docs
- **Vercel Go**: https://vercel.com/docs/functions/serverless-functions

---

## MISSION STATUS: BEGINNING EXECUTION

**Next Steps:**
1. ✅ Review this architecture document
2. ⏳ Create authentication system (Login/Register)
3. ⏳ Scaffold Go backend functions
4. ⏳ Implement Supabase schema
5. ⏳ Build multi-provider AI orchestration
6. ⏳ Deploy to Vercel

*Let's bring Project Moss to life.*
