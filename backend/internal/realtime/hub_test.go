package realtime

import (
	"io"
	"log/slog"
	"testing"
)

func TestHubBroadcastAndPresence(t *testing.T) {
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	client := &Client{send: make(chan []byte, 4), userID: "user-1", gameID: "game-1", logger: hub.logger}
	if count := hub.register(client); count != 1 {
		t.Fatalf("connection count = %d, want 1", count)
	}
	hub.connected(client, 1)
	if online, count := hub.Presence("user-1"); !online || count != 1 {
		t.Fatalf("presence = (%v, %d), want (true, 1)", online, count)
	}
	hub.Publish("game-1", Event{Type: EventMove})
	if len(client.send) != 3 {
		t.Fatalf("queued events = %d, want connected, presence, and move", len(client.send))
	}
	hub.disconnected(client)
	if online, count := hub.Presence("user-1"); online || count != 0 {
		t.Fatalf("presence after disconnect = (%v, %d), want (false, 0)", online, count)
	}
}

func TestHubRemovesSlowClient(t *testing.T) {
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	client := &Client{send: make(chan []byte, 1), userID: "slow-user", gameID: "game-2", logger: hub.logger}
	hub.register(client)
	hub.Publish("game-2", Event{Type: EventMove})
	hub.Publish("game-2", Event{Type: EventMove})
	if online, _ := hub.Presence("slow-user"); online {
		t.Fatal("slow client remained online after its queue filled")
	}
}
