-- VANA version-6: password authentication + persistent conversation history
-- Safe to run more than once. Purely additive: existing code keeps working.

-- 1. Passwords: bcrypt hash per user. NULL = legacy account created before
--    version-6; such accounts cannot log in until a password is set.
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS password_hash text;

-- Emails are matched case-insensitively at login/registration.
CREATE UNIQUE INDEX IF NOT EXISTS users_email_lower_key ON public.users (lower(email));

-- 2. History: fast "my conversations, newest first" and "messages in order".
CREATE INDEX IF NOT EXISTS idx_conversations_user_updated
    ON public.conversations (user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_conversation_created
    ON public.messages (conversation_id, created_at);
-- Rate limiting counts a user's recent messages.
CREATE INDEX IF NOT EXISTS idx_messages_user_sender_created
    ON public.messages (user_id, sender, created_at DESC);

-- 3. Crisis anchor: Indian crisis resources instead of US-only numbers.
UPDATE public.clinical_anchors
SET prompt_text = 'If the user mentions self-harm, suicide, or severe distress, immediately acknowledge their pain and provide crisis resources: Tele-MANAS (Govt. of India, free, 24x7): 14416 or 1-800-891-4416. In immediate danger: call 112 or go to the nearest hospital emergency.',
    version = version + 1
WHERE name = 'crisis_safety' AND prompt_text NOT LIKE '%14416%';
