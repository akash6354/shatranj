CREATE TABLE IF NOT EXISTS tournaments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    format VARCHAR(12) NOT NULL CHECK (format IN ('swiss', 'arena')),
    status VARCHAR(16) NOT NULL DEFAULT 'upcoming'
        CHECK (status IN ('upcoming', 'live', 'completed', 'cancelled')),
    time_control_seconds INTEGER NOT NULL CHECK (time_control_seconds BETWEEN 30 AND 86400),
    increment_seconds INTEGER NOT NULL DEFAULT 0 CHECK (increment_seconds BETWEEN 0 AND 3600),
    max_players INTEGER NOT NULL CHECK (max_players BETWEEN 2 AND 512),
    starts_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS tournaments_status_start_idx ON tournaments (status, starts_at);

CREATE TABLE IF NOT EXISTS tournament_participants (
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL DEFAULT 1200 CHECK (rating >= 0),
    score INTEGER NOT NULL DEFAULT 0 CHECK (score >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tournament_id, user_id)
);

CREATE TABLE IF NOT EXISTS tournament_rounds (
    tournament_id UUID NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL CHECK (round_number > 0),
    status VARCHAR(16) NOT NULL DEFAULT 'in_progress'
        CHECK (status IN ('in_progress', 'completed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,
    PRIMARY KEY (tournament_id, round_number)
);

CREATE TABLE IF NOT EXISTS tournament_pairings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id UUID NOT NULL,
    round_number INTEGER NOT NULL,
    white_player_id UUID NOT NULL REFERENCES users(id),
    black_player_id UUID REFERENCES users(id),
    game_id UUID REFERENCES games(id) ON DELETE SET NULL,
    result VARCHAR(8) NOT NULL DEFAULT '*'
        CHECK (result IN ('*', '1-0', '0-1', '1/2-1/2')),
    white_score INTEGER NOT NULL DEFAULT 0 CHECK (white_score >= 0),
    black_score INTEGER NOT NULL DEFAULT 0 CHECK (black_score >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (tournament_id, round_number)
        REFERENCES tournament_rounds(tournament_id, round_number) ON DELETE CASCADE,
    CONSTRAINT tournament_pairing_distinct CHECK (black_player_id IS NULL OR black_player_id <> white_player_id),
    UNIQUE (tournament_id, round_number, white_player_id),
    UNIQUE (tournament_id, round_number, black_player_id)
);

CREATE TABLE IF NOT EXISTS clubs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(80) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    visibility VARCHAR(12) NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS club_members (
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(12) NOT NULL DEFAULT 'member' CHECK (role IN ('owner', 'admin', 'moderator', 'member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (club_id, user_id)
);

CREATE TABLE IF NOT EXISTS club_join_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    club_id UUID NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(12) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (club_id, user_id)
);

CREATE INDEX IF NOT EXISTS club_join_requests_pending_idx
    ON club_join_requests (club_id, created_at) WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS friendships (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(12) NOT NULL CHECK (status IN ('pending', 'accepted', 'blocked')),
    requested_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id),
    CONSTRAINT friendship_distinct_users CHECK (user_id <> friend_id)
);

CREATE INDEX IF NOT EXISTS friendships_user_status_idx ON friendships (user_id, status);
CREATE INDEX IF NOT EXISTS friendships_friend_status_idx ON friendships (friend_id, status);

CREATE TABLE IF NOT EXISTS chat_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kind VARCHAR(12) NOT NULL CHECK (kind IN ('game', 'club', 'direct')),
    reference_id UUID,
    direct_key TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chat_room_reference_kind CHECK (
        (kind IN ('game', 'club') AND reference_id IS NOT NULL)
        OR (kind = 'direct' AND reference_id IS NULL AND direct_key IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS chat_room_reference_unique
    ON chat_rooms (kind, reference_id) WHERE reference_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS chat_room_members (
    room_id UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
);

CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL CHECK (length(content) BETWEEN 1 AND 2000),
    status VARCHAR(12) NOT NULL DEFAULT 'visible' CHECK (status IN ('visible', 'hidden')),
    moderation_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS chat_messages_room_created_idx
    ON chat_messages (room_id, created_at DESC, id DESC);
