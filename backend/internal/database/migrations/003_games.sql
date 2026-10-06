CREATE TABLE IF NOT EXISTS games (
    id UUID PRIMARY KEY,
    white_player_id UUID NOT NULL REFERENCES users(id),
    black_player_id UUID REFERENCES users(id),
    time_control_initial_seconds INTEGER NOT NULL CHECK (time_control_initial_seconds >= 0),
    time_control_increment_seconds INTEGER NOT NULL DEFAULT 0 CHECK (time_control_increment_seconds >= 0),
    mode VARCHAR(12) NOT NULL DEFAULT 'blitz' CHECK (mode IN ('bullet', 'blitz', 'rapid')),
    rated BOOLEAN NOT NULL DEFAULT FALSE,
    current_fen TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'waiting'
        CHECK (status IN ('waiting', 'active', 'finished')),
    result VARCHAR(8) NOT NULL DEFAULT '*'
        CHECK (result IN ('*', '1-0', '0-1', '1/2-1/2')),
    end_reason VARCHAR(32),
    draw_offer_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT games_distinct_players CHECK (black_player_id IS NULL OR black_player_id <> white_player_id),
    CONSTRAINT games_finished_consistency CHECK (
        (status = 'finished' AND result <> '*' AND finished_at IS NOT NULL)
        OR (status <> 'finished' AND result = '*' AND finished_at IS NULL)
    ),
    CONSTRAINT games_draw_offer_player CHECK (
        draw_offer_by IS NULL OR draw_offer_by = white_player_id OR draw_offer_by = black_player_id
    )
);

CREATE INDEX IF NOT EXISTS games_white_player_created_idx ON games (white_player_id, created_at DESC);
CREATE INDEX IF NOT EXISTS games_black_player_created_idx ON games (black_player_id, created_at DESC);
CREATE INDEX IF NOT EXISTS games_waiting_created_idx ON games (created_at) WHERE status = 'waiting';
