# Project Emerald Moss - Deployment Guide

## Pre-Deployment Checklist

### 1. Environment Setup
- [ ] Copy `.env.example` to `.env.local` (development) or `.env.production` (production)
- [ ] Fill in all required API keys (Groq, Gemini, JWT secret)
- [ ] Configure Supabase credentials
- [ ] Set appropriate JWT expiration and secrets

### 2. Database Setup
- [ ] Create Supabase project at https://supabase.com
- [ ] Run SQL migrations from `docs/SUPABASE_SCHEMA.sql`
- [ ] Enable Row-Level Security (RLS) on all tables
- [ ] Configure pgvector extension
- [ ] Populate clinical_anchors table with initial prompts
- [ ] Populate crisis_keywords table with safety terms

### 3. API Key Configuration
#### Groq
1. Sign up at https://console.groq.com
2. Generate API key
3. Set rate limit alerts (30 RPM)
4. Add to `.env`: `GROQ_API_KEY=gsk_...`

#### Google Gemini
1. Visit https://makersuite.google.com
2. Create API key
3. Enable Generative Language API
4. Set rate limit alerts (15 RPM)
5. Add to `.env`: `GEMINI_API_KEY=...`

### 4. Security Hardening
- [ ] Change JWT secret from default
- [ ] Enable HTTPS/TLS
- [ ] Configure CORS whitelist
- [ ] Set secure HTTP headers
- [ ] Enable rate limiting
- [ ] Configure HSTS
- [ ] Set CSP headers

---

## Local Development with Docker

### Quick Start
```bash
# 1. Install dependencies
npm install
npm ci

# 2. Copy environment template
cp .env.example .env.local

# 3. Fill in API keys in .env.local

# 4. Start Docker containers
docker-compose up -d

# 5. Wait for database to initialize (check health)
docker-compose ps

# 6. Seed database (if not auto-migrated)
docker exec emerald-moss-postgres psql -U postgres -d emerald_moss -f /docker-entrypoint-initdb.d/01-schema.sql

# 7. Open browser to http://localhost:4200
```

### Container Management
```bash
# View logs
docker-compose logs -f frontend
docker-compose logs -f backend
docker-compose logs -f postgres

# Rebuild containers
docker-compose rebuild

# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Access PostgreSQL
docker exec -it emerald-moss-postgres psql -U postgres -d emerald_moss

# Access pgAdmin
# Visit http://localhost:5050
# Email: admin@emerald-moss.local
# Password: admin
```

---

## Vercel Deployment (Production)

### Prerequisites
- [ ] Vercel account at https://vercel.com
- [ ] GitHub repository connected to Vercel
- [ ] Supabase project created and configured

### Step 1: Prepare Repository
```bash
# Ensure clean git status
git status

# Push to main branch
git push origin main
```

### Step 2: Create Vercel Project
1. Log in to Vercel dashboard
2. Click "New Project"
3. Select your GitHub repository
4. Configure:
   - **Framework**: Angular (or auto-detected)
   - **Root Directory**: `./`
   - **Build Command**: `npm run build`
   - **Output Directory**: `dist/vana`
   - **Install Command**: `npm install`

### Step 3: Configure Environment Variables
In Vercel Project Settings → Environment Variables, add:

```
JWT_SECRET=your-production-secret-key
JWT_EXPIRATION=604800

SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key

GROQ_API_KEY=gsk_your_key
GROQ_MODEL=groq-cloud/llama-3.1-8b-instant
GROQ_RPM_LIMIT=30

GEMINI_API_KEY=your_key
GEMINI_MODEL=gemini-1.5-flash
GEMINI_RPM_LIMIT=15

VITE_API_URL=https://your-vercel-url.vercel.app/api

LOG_LEVEL=info
NODE_ENV=production
```

### Step 4: Deploy
1. Push changes to main branch
2. Vercel auto-deploys
3. Monitor build in Vercel dashboard
4. Check deployment at `https://your-project.vercel.app`

### Step 5: Configure Vercel Functions
```json
// vercel.json
{
  "functions": {
    "api/**/*.go": {
      "runtime": "go@1.22"
    }
  },
  "rewrites": [
    {
      "source": "/api/(.*)",
      "destination": "/api/$1"
    }
  ]
}
```

---

## Post-Deployment Verification

### Healthchecks
```bash
# Frontend health
curl https://your-domain.vercel.app

# Backend health
curl https://your-domain.vercel.app/api/health

# Database connection
curl https://your-domain.vercel.app/api/db-status
```

### Test Authentication Flow
1. Visit https://your-domain/login
2. Register new account with email + password
3. Verify JWT token in localStorage
4. Test login/logout flow

### Test AI Providers
1. Log in to dashboard
2. Send test message
3. Verify response from Groq (Tier 1)
4. Monitor rate limiting in database

### Monitor Crisis Detection
1. Test with crisis keyword: "I want to die"
2. Verify emergency response appears
3. Check logs for escalation trigger

---

## Scaling Considerations

### Rate Limiting
Adjust `GROQ_RPM_LIMIT` and `GEMINI_RPM_LIMIT` based on usage:
- Small launch: 30 and 15
- Growth phase: 100 and 50
- Scale: Request higher limits from providers

### Database
- Monitor Supabase bandwidth
- Add read replicas if needed
- Implement connection pooling

### Caching Strategy
- Cache clinical knowledge embeddings
- Cache frequently asked responses
- Implement Redis layer for session management

### CDN & Edge Functions
- Enable Vercel Edge Middleware for auth checks
- Cache static assets globally
- Use ISR (Incremental Static Regeneration) for clinical content

---

## Monitoring & Alerting

### Sentry Integration (Error Tracking)
```bash
npm install @sentry/angular

# Configure in main.ts
import * as Sentry from "@sentry/angular";

Sentry.init({
  dsn: process.env['VITE_SENTRY_DSN'],
  environment: process.env['NODE_ENV'],
});
```

### Key Metrics to Track
- **API Response Time**: Target < 1000ms
- **Crisis Detection Rate**: Monitor false positives/negatives
- **Token Usage**: Groq vs Gemini distribution
- **Error Rate**: < 0.5% of requests
- **User Retention**: Daily/Monthly active users

### Logging Best Practices
```go
// Always log with context
log.Printf("[%s] User %s: %s", requestID, userID, message)
```

---

## Security Hardening Checklist

- [ ] Enable HTTPS everywhere
- [ ] Set HSTS header: `max-age=31536000; includeSubDomains`
- [ ] Configure CSP header
- [ ] Enable CORS only for trusted domains
- [ ] Implement rate limiting per IP/user
- [ ] Encrypt sensitive data in Supabase (PII)
- [ ] Enable audit logging for database changes
- [ ] Rotate JWT secret monthly
- [ ] Backup database daily
- [ ] Implement DDoS protection (Cloudflare)
- [ ] Regular security audits
- [ ] PII data retention policy (delete after 1 year)

---

## Troubleshooting

### Issue: 502 Bad Gateway
**Cause**: Backend function timeout or crash
**Solution**:
```bash
# Check logs
vercel logs --tail

# Increase function timeout (if available)
# Check Groq/Gemini API key validity
# Verify Supabase connection
```

### Issue: Crisis Keywords Not Detected
**Cause**: Keywords table empty or regex mismatch
**Solution**:
```sql
-- Verify crisis keywords exist
SELECT * FROM public.crisis_keywords;

-- Test regex pattern
SELECT 'I want to die'::text ~ '(suicide|kill|die)' as match;
```

### Issue: High Rate Limiting
**Cause**: Groq/Gemini RPM exhausted
**Solution**:
- Upgrade API plans with providers
- Implement queuing for non-urgent requests
- Cache common responses
- Increase Tier 2 fallback tolerance

---

## Support & Resources

- **Groq Support**: https://support.groq.com
- **Gemini Support**: https://support.google.com/ai
- **Supabase Support**: https://supabase.com/support
- **Vercel Support**: https://vercel.com/support
- **Angular Docs**: https://angular.io/docs

---

**Last Updated**: April 28, 2026
**Status**: Production Ready
**Version**: 1.0.0
