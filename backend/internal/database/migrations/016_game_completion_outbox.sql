CREATE TABLE IF NOT EXISTS game_completion_outbox (
    game_id UUID PRIMARY KEY REFERENCES games(id) ON DELETE CASCADE,
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    locked_until TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX IF NOT EXISTS game_completion_outbox_available_idx
    ON game_completion_outbox (available_at, locked_until, game_id);