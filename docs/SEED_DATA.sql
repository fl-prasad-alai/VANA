-- Seed data for VANA Mental Health Platform

-- 1. Create a dummy user in auth.users (to satisfy foreign key if needed)
INSERT INTO auth.users (id, email)
VALUES ('d290f1ee-6c54-4b01-90e6-d701748f0851', 'admin@vana.com')
ON CONFLICT (id) DO NOTHING;

-- 2. Create the user profile in public.users
INSERT INTO public.users (id, email, full_name, privacy_mode, is_active)
VALUES ('d290f1ee-6c54-4b01-90e6-d701748f0851', 'admin@vana.com', 'VANA Admin', 'encrypted', true)
ON CONFLICT (id) DO NOTHING;

-- 3. Create a test conversation
INSERT INTO public.conversations (id, user_id, title, status, ai_provider)
VALUES (gen_random_uuid(), 'd290f1ee-6c54-4b01-90e6-d701748f0851', 'Initial Stress Assessment', 'active', 'groq');

-- 4. Add some clinical knowledge
INSERT INTO public.clinical_knowledge (title, content, source, category)
VALUES 
('Mindful Breathing', 'Focus on your breath. Inhale for 4, hold for 4, exhale for 4.', 'mindfulness', 'coping'),
('Anxiety Management', 'Recognize the physical signs of anxiety and use grounding techniques (5-4-3-2-1).', 'clinical', 'anxiety');

-- 5. Add a second user for testing
INSERT INTO auth.users (id, email)
VALUES ('a1b2c3d4-e5f6-4a5b-b6c7-d8e9f0a1b2c3', 'user@example.com')
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.users (id, email, full_name, privacy_mode, is_active)
VALUES ('a1b2c3d4-e5f6-4a5b-b6c7-d8e9f0a1b2c3', 'user@example.com', 'Arjun Mehta', 'encrypted', true)
ON CONFLICT (id) DO NOTHING;
