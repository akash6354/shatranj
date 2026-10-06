CREATE TABLE IF NOT EXISTS puzzles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fen TEXT NOT NULL,
    solution_moves JSONB NOT NULL CHECK (
        jsonb_typeof(solution_moves) = 'array' AND jsonb_array_length(solution_moves) > 0
    ),
    themes JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(themes) = 'array'),
    difficulty VARCHAR(16) NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard', 'expert')),
    rating INTEGER NOT NULL DEFAULT 1200 CHECK (rating >= 0),
    explanation TEXT NOT NULL DEFAULT '',
    status VARCHAR(16) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS puzzles_published_rating_idx
    ON puzzles (rating, id) WHERE status = 'published';
CREATE INDEX IF NOT EXISTS puzzles_themes_idx ON puzzles USING GIN (themes);

CREATE TABLE IF NOT EXISTS daily_puzzles (
    puzzle_date DATE PRIMARY KEY,
    puzzle_id UUID NOT NULL REFERENCES puzzles(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS puzzle_ratings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL DEFAULT 1200 CHECK (rating >= 0),
    highest INTEGER NOT NULL DEFAULT 1200 CHECK (highest >= rating),
    games_played INTEGER NOT NULL DEFAULT 0 CHECK (games_played >= 0),
    correct INTEGER NOT NULL DEFAULT 0 CHECK (correct >= 0 AND correct <= games_played),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS puzzle_rush_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    current_puzzle_id UUID REFERENCES puzzles(id) ON DELETE SET NULL,
    status VARCHAR(12) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'finished', 'expired')),
    score INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
    streak INTEGER NOT NULL DEFAULT 0 CHECK (streak >= 0),
    time_limit_seconds INTEGER NOT NULL CHECK (time_limit_seconds BETWEEN 30 AND 1800),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT puzzle_rush_lifecycle CHECK (
        (status = 'active' AND finished_at IS NULL)
        OR (status <> 'active' AND finished_at IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS puzzle_rush_user_started_idx
    ON puzzle_rush_sessions (user_id, started_at DESC);
CREATE INDEX IF NOT EXISTS puzzle_rush_active_user_idx
    ON puzzle_rush_sessions (user_id, expires_at) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS puzzle_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    puzzle_id UUID NOT NULL REFERENCES puzzles(id) ON DELETE RESTRICT,
    rush_session_id UUID REFERENCES puzzle_rush_sessions(id) ON DELETE SET NULL,
    submitted_moves JSONB NOT NULL CHECK (jsonb_typeof(submitted_moves) = 'array'),
    correct BOOLEAN NOT NULL,
    user_rating_before INTEGER NOT NULL CHECK (user_rating_before >= 0),
    user_rating_after INTEGER NOT NULL CHECK (user_rating_after >= 0),
    puzzle_rating_before INTEGER NOT NULL CHECK (puzzle_rating_before >= 0),
    puzzle_rating_after INTEGER NOT NULL CHECK (puzzle_rating_after >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS puzzle_attempts_user_created_idx
    ON puzzle_attempts (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS puzzle_attempts_puzzle_created_idx
    ON puzzle_attempts (puzzle_id, created_at DESC);
CREATE INDEX IF NOT EXISTS puzzle_attempts_rush_session_idx
    ON puzzle_attempts (rush_session_id, created_at DESC)
    WHERE rush_session_id IS NOT NULL;
