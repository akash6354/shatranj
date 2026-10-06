package realtime

import "time"

const (
	EventConnected       = "connected"
	EventPresence        = "presence"
	EventPresenceUpdated = "presence.updated"
	EventPing            = "ping"
	EventPong            = "pong"
	EventGameCreated     = "game_created"
	EventGameJoined      = "game_joined"
	EventMove            = "move"
	EventGameState       = "game_state"
	EventResigned        = "resigned"
	EventDrawOffered     = "draw_offered"
	EventGameEnded       = "game_ended"
)

type Event struct {
	Type      string    `json:"type"`
	GameID    string    `json:"game_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Online    bool      `json:"online,omitempty"`
	Count     int       `json:"count,omitempty"`
	Data      any       `json:"data,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
