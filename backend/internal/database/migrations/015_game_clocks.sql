ALTER TABLE games
    ADD COLUMN IF NOT EXISTS white_clock_ms BIGINT NOT NULL DEFAULT 0 CHECK (white_clock_ms >= 0),
    ADD COLUMN IF NOT EXISTS black_clock_ms BIGINT NOT NULL DEFAULT 0 CHECK (black_clock_ms >= 0),
    ADD COLUMN IF NOT EXISTS clock_updated_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS games_active_clock_idx
    ON games (clock_updated_at, id) WHERE status = 'active';

UPDATE games
SET white_clock_ms = time_control_initial_seconds::BIGINT * 1000,
    black_clock_ms = time_control_initial_seconds::BIGINT * 1000
WHERE white_clock_ms = 0 AND black_clock_ms = 0;

UPDATE games
SET clock_updated_at = clock_timestamp()
WHERE status = 'active' AND clock_updated_at IS NULL;