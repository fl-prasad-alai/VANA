# PROJECT EMERALD MOSS - Workspace Architecture & Implementation Plan

## Mission Statement
A revolutionary, voice-first Biophilic Digital Triage System bridging the "Clinical Chasm" by providing instant, nature-inspired mental health support at **zero cost** using high-efficiency Go/React/Gemini stack.

---

## 1. PROJECT STRUCTURE (Full-Stack Docker)

```
VANA/
├── package.json                          # Frontend dependencies + scripts
├── tailwind.config.js                    # Custom palette
├── tsconfig.json                         # TypeScript root config
├── tsconfig.node.json                    # Vite config
├── vite.config.ts                        # Vite configuration
├── vercel.json                           # Vercel deployment config
├── docker-compose.yml                    # Local dev orchestration
├── Dockerfile.backend                    # Multi-stage Go build
├── Dockerfile.frontend                   # React build
├── .env.example                          # Environment template
├── .gitignore
│
├── api/                                  # Go Backend
│   ├── go.mod                            # Go module
│   ├── main.go                           # Main entry point
│   ├── handlers/                         # HTTP handlers
│   ├── middleware/                       # Crisis detection & Auth
│   ├── orchestration/                    # AI Provider Logic (Groq + Gemini)
│   └── database/                         # PostgreSQL + pgvector Client
│
├── src/                                  # React Frontend
│   ├── main.tsx                          # Entry point
│   ├── App.tsx                           # Root component
│   ├── components/                       # UI Components (Cinematic, Glassmorphism)
│   ├── pages/                            # Dashboard, Login, Signup
│   ├── contexts/                         # Auth & Theme State
│   └── hooks/                            # Custom Hooks (useAuth)
│
└── docs/                                 # Documentation
    ├── API_SPEC.md                       # API documentation
    ├── DEPLOYMENT.md                     # Deployment guide
    ├── TESTING_GUIDE.md                  # Comprehensive test plan
    └── SEED_DATA.sql                     # Test data
```

---

## 2. DATA FLOW ARCHITECTURE

### Frontend → Backend → AI Provider Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│                         USER INTERFACE                              │
│  (React 18 + Vite + Tailwind + Biophilic Design)                    │
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
│                      GO BACKEND SERVICE                            │
│  (Docker Containerized)                                            │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │ 1. Auth Handlers (JWT Issue/Verify)                         │ │
│  │ 2. Crisis Middleware (Regex Interceptor)                    │ │
│  │    ↓ DETECT CRISIS KEYWORDS → ROUTE TO HELPLINE             │ │
│  │ 3. Chat Handler (Message Processing)                        │ │
│  │    ↓ Route to AI Provider                                   │ │
│  │                                                              │ │
│  │ ┌─────────────────────────────────────────────────────────┐ │
│  │ │       MULTI-PROVIDER LOAD BALANCER                     │ │
│  │ │  ┌────────────────────────────────────────────────────┐ │ │
│  │ │  │ Tier 1 (Speed): Groq Llama 3.1 8B                │ │ │
│  │ │  └────────────────────────────────────────────────────┘ │ │
│  │ │  ↓ (Fallback)                                            │ │
│  │ │  ┌────────────────────────────────────────────────────┐ │ │
│  │ │  │ Tier 2 (Logic): Gemini 1.5 Flash                 │ │ │
│  │ │  └────────────────────────────────────────────────────┘ │ │
│  │ └─────────────────────────────────────────────────────────┘ │
│  │                                                              │ │
│  │ 4. Database: PostgreSQL + pgvector                        │ │
│  │    ├─ RAG Knowledge Base (Clinical Vectors)               │ │
│  │    ├─ User Sessions (Encrypted)                           │ │
│  │    └─ Conversation History                                │ │
│  └──────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────┘
```

---

## 3. TECHNICAL SPECIFICATIONS

### Frontend Stack
- **Framework**: React 18
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **State Management**: React Context API + Zustand
- **Navigation**: React Router DOM

### Backend Stack
- **Runtime**: Go 1.22
- **AI Providers**: Groq & Google Gemini
- **Database**: PostgreSQL with `pgvector`
- **Authentication**: JWT (HS256)

---

## 4. SECURITY & PRIVACY
- **Crisis Interceptor**: Real-time keyword monitoring for immediate helpline routing.
- **Data Encryption**: JWT-secured endpoints and encrypted-at-rest database.
- **Local Dev Isolation**: Full-stack Docker environment for consistent development.

---

## MISSION STATUS: EXECUTING
*Let's bring VANA to life.*
