package realtime

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shatranj/backend/internal/auth"
	"github.com/shatranj/backend/internal/games"
)

func TestGameWebSocketAuthenticatesAndBroadcasts(t *testing.T) {
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := tokens.Issue(auth.User{ID: "user-1", Email: "user@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	hub := NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	mux := http.NewServeMux()
	RegisterGameRoute(mux, hub, tokens, nil, func(_ context.Context, gameID, userID string) error {
		if gameID != "game-1" || userID != "user-1" {
			return errNotAPlayer
		}
		return nil
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	dialer := websocket.Dialer{}
	headers := http.Header{"Authorization": []string{"Bearer " + token}}
	connection, response, err := dialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/api/v1/games/game-1/ws", headers)
	if err != nil {
		t.Fatalf("Dial() error = %v, response = %#v", err, response)
	}
	defer connection.Close()
	_ = connection.SetReadDeadline(time.Now().Add(time.Second))
	for _, expected := range []string{EventConnected, EventPresence} {
		var event Event
		if err := connection.ReadJSON(&event); err != nil {
			t.Fatalf("ReadJSON(%s): %v", expected, err)
		}
		if event.Type != expected {
			t.Fatalf("event type = %q, want %q", event.Type, expected)
		}
	}
	hub.Publish("game-1", Event{Type: EventMove, GameID: "game-1"})
	var event Event
	if err := connection.ReadJSON(&event); err != nil {
		t.Fatalf("ReadJSON(move): %v", err)
	}
	if event.Type != EventMove {
		t.Fatalf("event type = %q, want %q", event.Type, EventMove)
	}
}

func TestGameWebSocketRejectsNonPlayer(t *testing.T) {
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := tokens.Issue(auth.User{ID: "outsider"})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	RegisterGameRoute(mux, NewHub(nil), tokens, nil, func(context.Context, string, string) error {
		return errNotAPlayer
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	request, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/games/game-1/ws", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusForbidden)
	}
}

var errNotAPlayer = games.ErrForbidden
