CREATE TABLE IF NOT EXISTS game_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL UNIQUE REFERENCES games(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(16) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'completed', 'unavailable', 'failed')),
    result JSONB,
    error_message TEXT,
    attempts SMALLINT NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT game_review_result_status CHECK (
        (status = 'completed' AND result IS NOT NULL)
        OR (status <> 'completed' AND result IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS game_reviews_pending_jobs_idx
    ON game_reviews (available_at, created_at)
    WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS game_reviews_abandoned_jobs_idx
    ON game_reviews (updated_at)
    WHERE status = 'running';
