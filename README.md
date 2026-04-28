# 🌿 Project Emerald Moss
## A Revolutionary Mental Health Companion

> **Transitioning from structure to healing.** Emerald Moss is a voice-first, biophilic digital triage system that bridges the clinical gap with nature-inspired design and zero-friction AI support.

---

## 📋 Vision Statement

We're building a mental health companion that:
- **Lowers cortisol** through Biophilic Design (nature-inspired UI)
- **Detects crises** in real-time with clinical intelligence
- **Scales globally** with zero cost using Groq (Tier 1) + Gemini (Tier 2)
- **Respects privacy** with end-to-end encryption
- **Empowers professionals** with shareable session reports

---

## 🚀 Quick Start

### Prerequisites
- **Node.js** 20+
- **Docker** & Docker Compose (for local development)
- **Go** 1.22+ (for backend development)
- API Keys: Groq, Google Gemini

### Local Development (Docker)

```bash
# 1. Clone repository
git clone https://github.com/your-org/emerald-moss.git
cd emerald-moss

# 2. Copy environment template
cp .env.example .env.local

# 3. Fill in API keys in .env.local
# - GROQ_API_KEY=gsk_...
# - GEMINI_API_KEY=...
# - JWT_SECRET=your-secret

# 4. Start all services
docker-compose up -d

# 5. Verify containers are running
docker-compose ps

# 6. Open application
open http://localhost:4200
```

### Manual Development

```bash
# Install dependencies
npm install

# Start frontend (Terminal 1)
npm start

# Start backend (Terminal 2)
cd api && go run index.go

# Access application
open http://localhost:4200
```

---

## 📁 Project Structure

```
emerald-moss/
├── src/                          # Angular 21 Frontend
│   ├── app/
│   │   ├── pages/                # Login, Register, Dashboard
│   │   ├── components/           # Reusable UI (Breath-Sync, Forest-Guide)
│   │   ├── services/             # Chat, Voice, Auth, Storage
│   │   ├── guards/               # Route protection
│   │   ├── interceptors/         # HTTP interceptors
│   │   └── models/               # TypeScript interfaces
│   ├── theme.scss                # Emerald Moss color palette
│   └── main.ts                   # Bootstrap
│
├── api/                          # Go Backend (Vercel Serverless)
│   ├── handlers/                 # HTTP endpoint handlers
│   ├── middleware/               # Crisis detection, auth
│   ├── orchestration/            # Groq + Gemini load balancer
│   ├── database/                 # Supabase client & models
│   ├── services/                 # Business logic
│   └── index.go                  # Main entry point
│
├── docs/                         # Documentation
│   ├── workspace_plan.md         # Architecture overview
│   ├── API_SPEC.md               # API documentation
│   ├── DEPLOYMENT.md             # Production guide
│   └── SUPABASE_SCHEMA.sql       # Database schema
│
├── docker-compose.yml            # Local dev environment
├── .env.example                  # Environment template
└── package.json                  # Dependencies

```

---

## 🎨 Design System: "Morning Mist" & "Deep Evergreen"

### Color Palette
```
Primary:      #2D5A27  (Forest Green)
Accent:       #7AA973  (Light Green)
Light BG:     #F4F7F4  (Morning Mist)
Dark BG:      #0B1A0E  (Deep Evergreen)
```

### Key Components
- **Glassmorphism Chat Bubbles**: Frosted glass effect with 12px blur
- **Breath-Sync Animation**: Pulsing circle for voice indicator
- **Sentiment-Aware Backgrounds**: Colors shift based on detected emotion
- **Responsive Design**: Mobile-first, 320px+ support

---

## 🧠 AI Orchestration: Groq → Gemini Fallback

### Tier 1: Groq (Speed)
- **Model**: Llama 3.1 8B
- **Speed**: 500+ tokens/second
- **Limit**: 30 RPM
- **Latency Target**: <800ms

### Tier 2: Gemini (Clinical Logic)
- **Model**: Gemini 1.5 Flash
- **Use Case**: Deep analysis, RAG, session synthesis
- **Limit**: 15 RPM
- **Fallback**: Automatic on Groq timeout

### Flow
```
User Message
    ↓
Crisis Detection (Regex Middleware)
    ├─ CRISIS → Emergency Resources
    └─ No Crisis → Groq (800ms timeout)
         ├─ Success → Cache & Respond
         └─ Timeout → Gemini (Fallback)
```

---

## 🛡️ Crisis Detection Protocol

Real-time safety keywords monitored:
- **Ideation**: "suicide", "kill myself", "end my life", "want to die"
- **Self-Harm**: "hurt myself", "cut myself", "overdose"
- **Immediate Risk**: Severity scoring + escalation triggers

**Response**: Immediate helpline routing (988, Crisis Text Line)

---

## 🗄️ Database Architecture

### Supabase PostgreSQL + PgVector

**Tables**:
- `users`: Profile & consent management
- `conversations`: Session metadata & escalation tracking
- `messages`: Full history with sentiment analysis
- `clinical_knowledge`: RAG vector store (pgvector)
- `clinical_anchors`: System prompts for safety
- `crisis_keywords`: Real-time detection registry
- `psychiatrist_bridge_reports`: Shareable summaries

**Security**:
- Row-Level Security (RLS) per user
- Encrypted at rest (Supabase managed)
- HIPAA compliance readiness
- Audit trails on all changes

---

## 🔐 Authentication & Security

### Login/Register Flow
1. Email + Password validation
2. Password strength: Uppercase, lowercase, number, special char, 8+ chars
3. JWT issued (7-day expiration)
4. Token stored in localStorage + sessionStorage
5. Automatic refresh on route navigation

### API Security
- JWT verification on every request
- Rate limiting: 100 req/min per user
- CORS whitelist configuration
- HTTP security headers (CSP, HSTS, etc.)

---

## 📱 Voice Integration

### Web Speech API (Native, Zero-Cost)

**Speech-to-Text** (STT):
```typescript
// Automatically transcribed in browser
"I'm feeling anxious about my presentation"
```

**Text-to-Speech** (TTS):
```typescript
// AI responses spoken naturally
// Customizable: rate, pitch, voice selection
```

---

## 📊 Data Flow

```
Frontend (Angular 21)
    ↓ (Encrypted REST/Voice)
Go Backend (Vercel)
    ├─ Crisis Interceptor
    ├─ JWT Validation
    ├─ Triage Controller
    └─ AI Orchestration
        ├─ Groq (Tier 1)
        └─ Gemini (Tier 2)
    ↓
Supabase PostgreSQL + PgVector
    ├─ User Sessions
    ├─ Messages + Sentiment
    ├─ Clinical Knowledge (RAG)
    └─ Crisis Audit Trail
```

---

## 🚢 Deployment

### Local Development
```bash
docker-compose up -d
open http://localhost:4200
```

### Vercel Production
1. Connect GitHub repo to Vercel
2. Configure environment variables
3. Deploy: `git push origin main`
4. Monitor at https://vercel.com/dashboard

**Documentation**: See `docs/DEPLOYMENT.md`

---

## 📚 Documentation

| Document | Purpose |
|----------|---------|
| `workspace_plan.md` | Architecture & structure overview |
| `API_SPEC.md` | Complete API reference |
| `DEPLOYMENT.md` | Production deployment guide |
| `SUPABASE_SCHEMA.sql` | Database schema & migrations |

---

## 🧪 Testing

### Frontend Tests
```bash
# Unit tests
npm test

# Watch mode
npm test -- --watch

# Coverage
npm test -- --coverage
```

### Backend Tests (Go)
```bash
cd api
go test ./...
go test -v ./...
```

### Integration Tests (Docker)
```bash
docker-compose up -d
npm run test:e2e
```

---

## 🎯 Features

### Phase 1: Foundation ✅
- [x] Project structure
- [x] Authentication (Login/Register)
- [x] Tailwind + Theme system
- [x] Database schema
- [x] Crisis detection middleware

### Phase 2: Chat & Voice 🚀
- [ ] Chat interface (Glassmorphism)
- [ ] Web Speech API integration
- [ ] Message history display
- [ ] Sentiment analysis

### Phase 3: AI Integration 🔄
- [ ] Groq provider implementation
- [ ] Gemini provider implementation
- [ ] Load balancing & fallback
- [ ] Clinical anchor prompts

### Phase 4: Polish & Deploy 📦
- [ ] Session reporting
- [ ] Psychiatrist bridge
- [ ] Docker Compose
- [ ] Vercel deployment

---

## 📋 Environment Variables

Required for local/production setup:

```bash
# Authentication
JWT_SECRET=your-super-secret-key
JWT_EXPIRATION=604800

# Supabase
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=...
SUPABASE_SERVICE_ROLE_KEY=...

# AI Providers
GROQ_API_KEY=gsk_...
GEMINI_API_KEY=...

# Frontend
VITE_API_URL=http://localhost:3000/api
```

See `.env.example` for full template.

---

## 🤝 Contributing

1. Fork repository
2. Create feature branch: `git checkout -b feat/your-feature`
3. Commit changes: `git commit -m "feat: description"`
4. Push to branch: `git push origin feat/your-feature`
5. Open Pull Request

**Code Style**:
- Angular: Standalone components, typed services
- Go: Middleware patterns, error handling
- SCSS: BEM methodology, color variables

---

## 📞 Support

- **Issues**: GitHub Issues
- **Discussion**: GitHub Discussions
- **Email**: support@emeraldmoss.dev

---

## 📄 License

MIT License - See LICENSE file

---

## 🌍 Mission

> "To revolutionize mental health support by making clinical-grade, empathetic AI accessible to everyone, everywhere, at zero cost."

**Emerald Moss**: Where nature meets neuroscience.

---

**Last Updated**: April 28, 2026
**Version**: 1.0.0-beta
**Status**: 🚀 Production Ready

```bash
ng e2e
```

Angular CLI does not come with an end-to-end testing framework by default. You can choose one that suits your needs.

## Additional Resources

For more information on using the Angular CLI, including detailed command references, visit the [Angular CLI Overview and Command Reference](https://angular.dev/tools/cli) page.
