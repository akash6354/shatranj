CREATE TABLE IF NOT EXISTS matchmaking_queue (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mode VARCHAR(12) NOT NULL CHECK (mode IN ('bullet', 'blitz', 'rapid')),
    initial_seconds INTEGER NOT NULL CHECK (initial_seconds > 0),
    increment_seconds INTEGER NOT NULL CHECK (increment_seconds >= 0),
    rating INTEGER NOT NULL CHECK (rating >= 0),
    status VARCHAR(12) NOT NULL DEFAULT 'queued'
        CHECK (status IN ('queued', 'matched', 'paired', 'cancelled')),
    match_id UUID,
    game_id UUID REFERENCES games(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT matchmaking_match_consistency CHECK (
        (status = 'paired' AND match_id IS NOT NULL AND game_id IS NOT NULL)
        OR (status = 'matched' AND match_id IS NOT NULL AND game_id IS NULL)
        OR (status IN ('queued', 'cancelled') AND match_id IS NULL AND game_id IS NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS matchmaking_one_active_entry_per_user
    ON matchmaking_queue (user_id) WHERE status IN ('queued', 'matched', 'paired');
CREATE INDEX IF NOT EXISTS matchmaking_candidates_idx
    ON matchmaking_queue (mode, initial_seconds, increment_seconds, status, created_at)
    WHERE status = 'queued';
