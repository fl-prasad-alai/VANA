# PROJECT EMERALD MOSS - Developer Quick Reference

## 🚀 Get Started in 5 Minutes

### 1. Clone & Install
```bash
git clone <repo-url>
cd emerald-moss
npm install
```

### 2. Environment Setup
```bash
cp .env.example .env.local

# Edit .env.local and add:
# - GROQ_API_KEY=gsk_...
# - GEMINI_API_KEY=...
# - JWT_SECRET=your-secret
```

### 3. Run Locally
```bash
# Option A: Docker (Recommended)
docker-compose up -d
open http://localhost:4200

# Option B: Manual
npm start                    # Terminal 1
cd api && go run index.go   # Terminal 2
```

### 4. Test Flow
```
Login Page: http://localhost:4200/login
Register: http://localhost:4200/register
Dashboard: http://localhost:4200/dashboard
```

---

## 📁 Key Files to Know

### Frontend
| File | Purpose |
|------|---------|
| `src/app/app.routes.ts` | Route definitions (Login, Register, Dashboard) |
| `src/app/services/auth.service.ts` | Authentication logic |
| `src/app/services/chat.service.ts` | Chat message management |
| `src/app/services/voice.service.ts` | Web Speech API (STT/TTS) |
| `src/theme.scss` | Emerald Moss color system + animations |
| `src/app/pages/login/login.component.ts` | Login UI |
| `src/app/pages/register/register.component.ts` | Register UI |

### Backend
| File | Purpose |
|------|---------|
| `api/index.go` | Main entry point |
| `api/database/models.go` | Data structures |
| `api/database/supabase.go` | Database client |
| `api/middleware/crisis.go` | Crisis detection |
| `api/orchestration/provider.go` | AI load balancer |
| `api/orchestration/groq.go` | Groq API client |
| `api/orchestration/gemini.go` | Gemini API client |

### Configuration
| File | Purpose |
|------|---------|
| `.env.example` | Environment template |
| `docker-compose.yml` | Local dev services |
| `vercel.json` | Vercel deployment config |
| `tailwind.config.js` | Tailwind + custom colors |

### Documentation
| File | Purpose |
|------|---------|
| `README.md` | Project overview |
| `workspace_plan.md` | Architecture details |
| `API_SPEC.md` | API endpoints |
| `DEPLOYMENT.md` | Production guide |
| `SUPABASE_SCHEMA.sql` | Database setup |

---

## 🔑 Important Code Snippets

### Add a Component
```bash
ng generate component features/my-component --standalone
```

### Create a Service
```bash
ng generate service services/my-service
```

### Add a Route
```typescript
// src/app/app.routes.ts
{
  path: 'new-page',
  component: NewPageComponent,
  canActivate: [authGuard]
}
```

### Send Chat Message
```typescript
// From component
constructor(private chatService: ChatService) {}

sendMessage(text: string) {
  const request: ChatRequest = {
    message: text,
    conversationId: this.conversationId,
    isVoice: false
  };
  
  this.chatService.sendMessage(request).subscribe({
    next: (response) => console.log(response),
    error: (err) => console.error(err)
  });
}
```

### Add Crisis Keyword (Go)
```go
// api/middleware/crisis.go
patterns := []string{
  `\b(your-keyword|another-keyword)\b`,
  // ...
}
```

---

## 🎨 Design Tokens

### Colors
```scss
$moss-primary: #2D5A27;     // Primary actions
$moss-accent: #7AA973;      // Highlights
$moss-light: #F4F7F4;       // Light background
$moss-dark: #0B1A0E;        // Dark background
```

### Use in Components
```html
<button class="btn btn-primary dark:btn-dark">
  Click me
</button>
```

### Glassmorphism
```html
<div class="glass-panel dark:glass-panel">
  Content here
</div>
```

---

## 🧪 Common Tasks

### Run Tests
```bash
npm test                    # Angular tests
cd api && go test ./...    # Go tests
```

### Build for Production
```bash
npm run build              # Creates dist/
docker build -f Dockerfile.backend -t emerald-moss:latest .
```

### Check Database
```bash
# Docker exec
docker exec -it emerald-moss-postgres \
  psql -U postgres -d emerald_moss

# List tables
\dt

# View users
SELECT * FROM public.users;
```

### View Logs
```bash
docker-compose logs -f frontend
docker-compose logs -f backend
docker-compose logs -f postgres
```

---

## 🚨 Debugging Tips

### Frontend Issue: White Screen
```bash
# Check browser console for errors
# F12 → Console tab
# Clear localStorage: DevTools → Application → localStorage → clear
```

### Backend Issue: 500 Error
```bash
# Check backend logs
docker-compose logs -f backend

# Check database connection
curl http://localhost:3000/api/health
```

### Database Issue: No Data
```bash
# Verify tables exist
docker exec emerald-moss-postgres \
  psql -U postgres -d emerald_moss -c "\dt"

# Re-run migrations
docker exec emerald-moss-postgres \
  psql -U postgres -d emerald_moss \
  -f /docker-entrypoint-initdb.d/01-schema.sql
```

---

## 📊 API Endpoints (Local)

| Method | Endpoint | Purpose |
|--------|----------|---------|
| POST | `/api/auth/register` | Create account |
| POST | `/api/auth/login` | Sign in |
| POST | `/api/conversations` | New chat |
| POST | `/api/chat` | Send message |
| GET | `/api/conversations/{id}/messages` | View history |
| POST | `/api/triage` | Submit GAD-7 |
| GET | `/api/health` | Health check |

---

## 🔐 Authentication Flow

```
1. Register → JWT issued
2. JWT stored in localStorage
3. AuthInterceptor attaches JWT to requests
4. Backend verifies JWT on every request
5. Unauthorized → redirect to login
```

### Test Auth
```bash
# Register
curl -X POST http://localhost:3000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Pass123!","full_name":"Test"}'

# Login
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Pass123!"}'
```

---

## 🎯 Crisis Detection Test

```bash
# Send crisis message
curl -X POST http://localhost:3000/api/chat \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "conversation_id": "uuid",
    "message": "I want to kill myself",
    "is_voice": false
  }'

# Expected response: Crisis flag + helpline resources
```

---

## 📈 Performance Monitoring

### Check API Response Time
```bash
time curl http://localhost:3000/api/health
```

### Monitor Database Queries
```sql
-- Check slow queries
SELECT * FROM pg_stat_statements 
ORDER BY total_exec_time DESC 
LIMIT 10;
```

### Check Memory Usage
```bash
docker stats emerald-moss-backend
```

---

## 🚀 Deployment Checklist

- [ ] All tests passing
- [ ] No lint errors (`ng lint`)
- [ ] Environment variables set
- [ ] Database migrations run
- [ ] Crisis keywords verified
- [ ] API keys valid
- [ ] CORS configured
- [ ] Security headers set
- [ ] Rate limiting enabled
- [ ] Monitoring configured

---

## 📞 Common Questions

**Q: How do I change the AI provider?**
A: Edit `api/orchestration/provider.go` to adjust timeouts or switch providers.

**Q: Where are user passwords stored?**
A: In Supabase Auth (never stored in `users` table). Only JWT tokens in localStorage.

**Q: How does crisis detection work?**
A: Go middleware runs regex pattern matching on all chat messages before AI processing.

**Q: Can I add custom crisis keywords?**
A: Yes, insert into `crisis_keywords` table and API loads them on startup.

**Q: How do I increase rate limits?**
A: Edit environment variables `GROQ_RPM_LIMIT` and `GEMINI_RPM_LIMIT`.

**Q: Where is user data stored?**
A: Supabase PostgreSQL (encrypted at rest). Client can optionally use local encryption.

---

## 🎓 Learning Resources

- Angular 21: https://angular.io/docs
- Go: https://golang.org/doc
- Supabase: https://supabase.com/docs
- Web Speech API: https://developer.mozilla.org/en-US/docs/Web/API/Web_Speech_API
- Vercel Functions: https://vercel.com/docs/functions

---

## 🆘 Getting Help

1. **Check the docs**: `docs/` folder
2. **Search issues**: GitHub Issues
3. **Review code**: Look at existing implementations
4. **Run tests**: `npm test && go test ./...`
5. **Check logs**: `docker-compose logs -f`

---

## 📋 Standard Commit Messages

```bash
git commit -m "feat: add login component"
git commit -m "fix: crisis detection regex"
git commit -m "docs: update API documentation"
git commit -m "refactor: optimize database queries"
git commit -m "test: add auth service tests"
```

---

**Happy Coding! 🌿**

*Remember: Great code is not complex code. It's clear, documented, tested code that solves real problems.*

---

**Last Updated**: April 28, 2026
**Version**: 1.0.0
