# VANA Deployment Guide (Vercel + Supabase)

Follow these steps to deploy VANA to production.

## 1. Supabase Setup (Database)

1.  **Create a Project**: Sign in to [Supabase](https://supabase.com) and create a new project.
2.  **Database Schema**:
    *   Open the **SQL Editor** in your Supabase dashboard.
    *   Paste and run the contents of `docs/SUPABASE_SCHEMA.sql` (found in this repository).
    *   This will set up the `users`, `conversations`, `messages`, and `clinical_knowledge` tables, along with the `pgvector` extension for RAG.
3.  **Get Connection String**:
    *   Go to **Project Settings > Database**.
    *   Under **Connection String**, select **URI**.
    *   Copy the string (it looks like `postgres://postgres.[project-id]:[password]@aws-0-us-east-1.pooler.supabase.com:5432/postgres`).
    *   **Keep this safe!** You will need it for Vercel.

## 2. Vercel Setup (Frontend + API)

1.  **Push to GitHub**: Ensure your latest changes are pushed to a GitHub repository.
2.  **Import to Vercel**:
    *   Sign in to [Vercel](https://vercel.com).
    *   Click **Add New > Project** and import your VANA repository.
3.  **Configure Project Settings**:
    *   **Framework Preset**: Select `Vite` (Vercel should auto-detect this).
    *   **Root Directory**: Leave as `./`.
    *   **Build Command**: `npm run build`.
    *   **Output Directory**: `dist`.
4.  **Environment Variables**:
    Add the following keys in **Project Settings > Environment Variables**:
    *   `DATABASE_URL`: The Supabase URI you copied in Step 1.
    *   `JWT_SECRET`: A long, random string (e.g., `openssl rand -base64 32`).
    *   `GROQ_API_KEY`: Your Groq API key (from [console.groq.com](https://console.groq.com)).
    *   `GEMINI_API_KEY`: Your Google Gemini API key (from [aistudio.google.com](https://aistudio.google.com)).
    *   `VITE_API_URL`: `/api` (Vercel will route this correctly via `vercel.json`).

## 3. Post-Deployment: Clinical Knowledge (RAG)

To enable the AI to use clinical context, you need to seed the `clinical_knowledge` table in Supabase.

1.  Use the Supabase SQL Editor to run `docs/SEED_DATA.sql`.
2.  Alternatively, use a tool like `psql` to run the seed file against your Supabase URI:
    ```bash
    psql "[YOUR_SUPABASE_URI]" -f docs/SEED_DATA.sql
    ```

---

## 🛠️ Maintenance & Monitoring

*   **Logs**: View real-time API logs in the Vercel **Logs** tab. Look for `api/index.go` execution.
*   **Database**: Use the Supabase **Table Editor** to view users and conversation history.
*   **Scaling**: Vercel handles the API scaling automatically. Supabase can be scaled in the project settings if traffic increases.

**Status**: 🚀 Production Ready | **Last Updated**: May 8, 2026
