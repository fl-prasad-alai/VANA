-- 007: a short private memory per user, rebuilt from their conversations
-- (on sign-out, when the tab closes, and before a new conversation starts).
-- memory_updated_at is the created_at of the newest message already folded
-- into the memory, so each refresh only reads newer messages.
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS memory_summary text;
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS memory_updated_at timestamptz;
