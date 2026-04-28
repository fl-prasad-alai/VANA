# Project VANA - Deployment Guide

## 🚀 Local Development (Docker)

The fastest way to run VANA is using Docker Compose.

### Quick Start
```bash
# 1. Start all services
docker compose up -d

# 2. Verify containers
docker compose ps

# 3. Seed test data
docker exec -i emerald-moss-postgres psql -U postgres -d emerald_moss < docs/SEED_DATA.sql

# 4. Access
- Frontend: http://localhost:5173
- Backend: http://localhost:3000
- pgAdmin: http://localhost:5050 (admin@emerald-moss.local / admin)
```

## 🛠️ Production Deployment (Vercel + Supabase)

### Backend (Vercel Go Functions)
1. Set up a Vercel project.
2. Add environment variables:
   - `JWT_SECRET`
   - `DATABASE_URL`
   - `GROQ_API_KEY`
   - `GEMINI_API_KEY`
3. Vercel will automatically detect the `/api/*.go` handlers.

### Database (Supabase / Managed Postgres)
1. Provision a PostgreSQL instance with the `pgvector` extension.
2. Run the schema from `docs/SUPABASE_SCHEMA.sql`.
3. Set up the connection string in your backend environment variables.

### Frontend (Vercel React)
1. Link your repository.
2. Set build command: `npm run build`.
3. Set output directory: `dist`.
4. Set environment variable: `VITE_API_URL`.

---
**Last Updated**: April 28, 2026
**Status**: Production Ready
