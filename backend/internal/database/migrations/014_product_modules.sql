ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role VARCHAR(16) NOT NULL DEFAULT 'user'
        CHECK (role IN ('user', 'coach', 'moderator', 'admin')),
    ADD COLUMN IF NOT EXISTS status VARCHAR(16) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended'));

CREATE TABLE IF NOT EXISTS achievement_definitions (
    key VARCHAR(80) PRIMARY KEY,
    name VARCHAR(120) NOT NULL,
    description VARCHAR(500) NOT NULL,
    event_type VARCHAR(80) NOT NULL,
    threshold INTEGER NOT NULL CHECK (threshold > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_achievement_progress (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(80) NOT NULL,
    progress BIGINT NOT NULL DEFAULT 0 CHECK (progress >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, event_type)
);

CREATE TABLE IF NOT EXISTS achievement_events (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_type VARCHAR(80) NOT NULL,
    source_id VARCHAR(160) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, event_type, source_id)
);

CREATE TABLE IF NOT EXISTS user_achievements (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_key VARCHAR(80) NOT NULL REFERENCES achievement_definitions(key) ON DELETE CASCADE,
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, achievement_key)
);

INSERT INTO achievement_definitions (key, name, description, event_type, threshold) VALUES
    ('first_game', 'First Game', 'Complete your first chess game.', 'game_completed', 1),
    ('ten_games', 'Getting Started', 'Complete ten chess games.', 'game_completed', 10),
    ('first_puzzle', 'Tactical Start', 'Solve your first puzzle.', 'puzzle_solved', 1),
    ('hundred_puzzles', 'Puzzle Regular', 'Solve one hundred puzzles.', 'puzzle_solved', 100)
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(80) NOT NULL,
    title VARCHAR(160) NOT NULL,
    body VARCHAR(2000) NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notifications_user_unread_idx
    ON notifications (user_id, created_at DESC) WHERE read_at IS NULL;
CREATE INDEX IF NOT EXISTS notifications_user_created_idx
    ON notifications (user_id, created_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(24) NOT NULL CHECK (provider IN ('razorpay')),
    provider_order_id VARCHAR(80) UNIQUE,
    provider_payment_id VARCHAR(80) UNIQUE,
    amount_paise BIGINT NOT NULL CHECK (amount_paise > 0),
    currency CHAR(3) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'paid', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS payments_user_created_idx ON payments (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS payments_status_created_idx ON payments (status, created_at DESC);

CREATE TABLE IF NOT EXISTS admin_payment_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL REFERENCES users(id),
    reason VARCHAR(1000) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved')),
    resolved_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS admin_payment_flags_open_idx
    ON admin_payment_flags (created_at DESC) WHERE status = 'open';

CREATE TABLE IF NOT EXISTS user_subscriptions (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    plan_id VARCHAR(24) NOT NULL CHECK (plan_id IN ('free', 'premium')),
    status VARCHAR(16) NOT NULL CHECK (status IN ('active', 'cancelled', 'expired')),
    started_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS subscription_payment_activations (
    payment_id UUID PRIMARY KEY REFERENCES payments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS coach_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    bio VARCHAR(3000) NOT NULL DEFAULT '',
    rate_paise BIGINT NOT NULL DEFAULT 0 CHECK (rate_paise >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    specialties JSONB NOT NULL DEFAULT '[]'::jsonb,
    availability JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS coach_profiles_public_idx ON coach_profiles (rate_paise, user_id) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS coach_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES coach_profiles(user_id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes BETWEEN 15 AND 240),
    notes VARCHAR(1000) NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled', 'completed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT coach_booking_distinct_users CHECK (coach_id <> student_id)
);

CREATE INDEX IF NOT EXISTS coach_bookings_coach_time_idx
    ON coach_bookings (coach_id, starts_at) WHERE status IN ('pending', 'accepted');
CREATE INDEX IF NOT EXISTS coach_bookings_student_created_idx
    ON coach_bookings (student_id, created_at DESC);

CREATE TABLE IF NOT EXISTS news_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(180) NOT NULL,
    slug VARCHAR(180) NOT NULL UNIQUE,
    excerpt VARCHAR(500) NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    cover_url TEXT,
    tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(16) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'archived')),
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS news_posts_published_idx
    ON news_posts (published_at DESC, id DESC) WHERE status = 'published';
