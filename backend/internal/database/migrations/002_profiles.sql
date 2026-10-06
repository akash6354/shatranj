CREATE TABLE IF NOT EXISTS profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    username VARCHAR(24) NOT NULL UNIQUE,
    display_name VARCHAR(80) NOT NULL,
    avatar_url TEXT,
    country VARCHAR(2),
    bio VARCHAR(500),
    ratings_bullet INTEGER NOT NULL DEFAULT 1200 CHECK (ratings_bullet >= 0),
    ratings_blitz INTEGER NOT NULL DEFAULT 1200 CHECK (ratings_blitz >= 0),
    ratings_rapid INTEGER NOT NULL DEFAULT 1200 CHECK (ratings_rapid >= 0),
    ratings_classical INTEGER NOT NULL DEFAULT 1200 CHECK (ratings_classical >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT profiles_username_format CHECK (username ~ '^[a-z0-9_]{3,24}$'),
    CONSTRAINT profiles_display_name_not_empty CHECK (length(trim(display_name)) > 0)
);
