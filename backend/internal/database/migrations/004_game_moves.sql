CREATE TABLE IF NOT EXISTS game_moves (
    game_id UUID NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    ply INTEGER NOT NULL CHECK (ply > 0),
    move_number INTEGER NOT NULL CHECK (move_number > 0),
    player_id UUID NOT NULL REFERENCES users(id),
    uci VARCHAR(5) NOT NULL,
    san VARCHAR(16) NOT NULL,
    fen_after TEXT NOT NULL,
    white_clock_ms BIGINT CHECK (white_clock_ms IS NULL OR white_clock_ms >= 0),
    black_clock_ms BIGINT CHECK (black_clock_ms IS NULL OR black_clock_ms >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (game_id, ply)
);

CREATE INDEX IF NOT EXISTS game_moves_player_idx ON game_moves (player_id);
