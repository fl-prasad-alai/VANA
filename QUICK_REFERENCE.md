# PROJECT VANA - Developer Quick Reference

## 🚀 Get Started in 5 Minutes

### 1. Clone & Install
```bash
git clone <repo-url>
cd VANA
npm install
```

### 2. Run Locally (Docker)
```bash
docker compose up -d
open http://localhost:5173
```

### 3. Seed Test Data
```bash
docker exec -i emerald-moss-postgres psql -U postgres -d emerald_moss < docs/SEED_DATA.sql
```

---

## 📁 Key Files to Know

### Frontend (React)
| File | Purpose |
|------|---------|
| `src/main.tsx` | Entry point |
| `src/App.tsx` | Routing & Layout |
| `src/pages/Dashboard.tsx` | Main Chat Interface |
| `src/hooks/useAuth.tsx` | Auth Hook |
| `src/contexts/ThemeContext.tsx` | Biophilic Theme State |

### Backend (Go)
| File | Purpose |
|------|---------|
| `api/main.go` | Main Entry Point |
| `api/database/supabase.go` | DB Logic |
| `api/middleware/crisis.go` | Crisis Interception |
| `api/orchestration/provider.go` | AI Load Balancer |

---

## 🔑 Common Tasks

### Run Dev Server
```bash
npm run dev
```

### Build for Production
```bash
npm run build
```

### Check Database
```bash
docker exec -it emerald-moss-postgres psql -U postgres -d emerald_moss
```

### View Logs
```bash
docker compose logs -f backend
docker compose logs -f frontend
```

---

## 📊 API Endpoints (Local)
- **POST** `/api/auth/register`: Create account
- **POST** `/api/auth/login`: Sign in
- **POST** `/api/chat`: Send message
- **GET** `/api/health`: Health check

---
**Happy Coding! 🌿**
