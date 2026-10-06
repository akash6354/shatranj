CREATE TABLE IF NOT EXISTS ratings (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mode VARCHAR(12) NOT NULL CHECK (mode IN ('bullet', 'blitz', 'rapid', 'puzzle')),
    rating INTEGER NOT NULL CHECK (rating >= 0),
    highest INTEGER NOT NULL CHECK (highest >= rating),
    games_played INTEGER NOT NULL DEFAULT 0 CHECK (games_played >= 0),
    wins INTEGER NOT NULL DEFAULT 0 CHECK (wins >= 0),
    draws INTEGER NOT NULL DEFAULT 0 CHECK (draws >= 0),
    losses INTEGER NOT NULL DEFAULT 0 CHECK (losses >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, mode),
    CONSTRAINT ratings_result_count CHECK (games_played = wins + draws + losses)
);

CREATE TABLE IF NOT EXISTS rated_game_results (
    game_id UUID PRIMARY KEY REFERENCES games(id) ON DELETE CASCADE,
    mode VARCHAR(12) NOT NULL CHECK (mode IN ('bullet', 'blitz', 'rapid', 'puzzle')),
    result VARCHAR(8) NOT NULL CHECK (result IN ('1-0', '0-1', '1/2-1/2')),
    white_user_id UUID NOT NULL REFERENCES users(id),
    black_user_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT rated_game_distinct_players CHECK (white_user_id <> black_user_id)
);

CREATE INDEX IF NOT EXISTS ratings_mode_rating_idx ON ratings (mode, rating DESC);
