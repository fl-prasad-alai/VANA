# VANA Mental Health Platform - Testing Guide

This document provides a comprehensive guide to running and testing the VANA platform across all services.

---

## 🚀 1. Environment Setup

Ensure the platform is running in Docker:

```bash
# Start all services
docker compose up -d

# Check service status
docker compose ps
```

**Services & Ports:**
- **Frontend (React/Vite)**: [http://localhost:5173](http://localhost:5173)
- **Backend API (Go)**: [http://localhost:3000](http://localhost:3000)
- **Database Admin (pgAdmin)**: [http://localhost:5050](http://localhost:5050)

---

## 🔐 2. Test Credentials

The database is pre-seeded with the following test accounts:

| User Role | Email | Password | Purpose |
| :--- | :--- | :--- | :--- |
| **Administrator** | `admin@vana.com` | `password123` | Full access, seeded conversation |
| **Standard User** | `user@example.com` | `password123` | Clean account for testing registration/onboarding |

---

## 🧪 3. Backend API Testing

### A. Health & Connectivity
Verify the backend is alive and connected to the database.

```bash
curl http://localhost:3000/api/health
# Expected: {"status": "healthy"}
```

### B. Authentication Flow
Test the login mechanism. Note that the system currently uses email validation for mock local testing.

```bash
curl -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@vana.com", "password": "password123"}'
```

### C. Chat & Crisis Interception
VANA features a real-time safety layer that intercepts crisis keywords.

**Scenario 1: Empathetic Chat**
```bash
curl -X POST http://localhost:3000/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "I feel a bit overwhelmed today"}'
```

**Scenario 2: Crisis Trigger (Safety Test)**
Type a message containing crisis keywords (e.g., "suicide", "end my life").
```bash
curl -X POST http://localhost:3000/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "I want to end my life"}'
```
*Expected: The system should respond with immediate crisis resources (988 Lifeline).*

---

## 🖥️ 4. Frontend UI/UX Testing Flow

Follow this flow to test the user journey:

1.  **Onboarding**:
    *   Navigate to `http://localhost:5173`.
    *   Verify the biophilic (green/calming) theme loads correctly.
2.  **Authentication**:
    *   Click "Login" and enter `admin@vana.com` / `password123`.
    *   Verify redirect to the Chat Dashboard.
3.  **Conversation**:
    *   Send a message: "Help me with my anxiety."
    *   Verify the AI response includes evidence-based mindfulness tips (RAG Knowledge).
4.  **Responsive Design**:
    *   Open Browser DevTools (F12) and toggle "Device Toolbar".
    *   Verify the layout adapts gracefully to Mobile (iPhone/Pixel) and Tablet views.

---

## 📊 5. Database & Analytics Testing

### pgAdmin Verification
1.  Open [http://localhost:5050](http://localhost:5050).
2.  Login with `admin@emerald-moss.local` / `admin`.
3.  Add a new Server:
    *   **Name**: `VANA DB`
    *   **Connection Host**: `emerald-moss-postgres`
    *   **Username/Password**: `postgres` / `postgres`
4.  **Check Tables**:
    *   `public.users`: Verify user counts.
    *   `public.messages`: Verify chat history persistence.
    *   `public.clinical_knowledge`: Verify RAG data is available.

---

## 🛠️ 6. Troubleshooting

| Issue | Solution |
| :--- | :--- |
| **Port 3000 conflict** | Run `docker ps` to find and stop other processes on port 3000. |
| **Database connection refused** | Wait 10 seconds for Postgres to initialize, then run `docker compose restart backend`. |
| **Changes not reflecting** | Run `docker compose up -d --build` to force a rebuild of the containers. |
| **Missing test data** | Run: `docker exec -i emerald-moss-postgres psql -U postgres -d emerald_moss < docs/SEED_DATA.sql` |
