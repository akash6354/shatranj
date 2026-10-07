package realtime

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/shatranj/backend/internal/cache"
)

type testFanoutBroker struct {
	mu          sync.Mutex
	subscribers map[chan []byte]struct{}
}

func newTestFanoutBroker() *testFanoutBroker {
	return &testFanoutBroker{subscribers: make(map[chan []byte]struct{})}
}

func (b *testFanoutBroker) Get(context.Context, string) ([]byte, error) {
	return nil, cache.ErrCacheMiss
}
func (b *testFanoutBroker) Set(context.Context, string, []byte, time.Duration) error { return nil }
func (b *testFanoutBroker) Delete(context.Context, string) error                     { return nil }
func (b *testFanoutBroker) Ping(context.Context) error                               { return nil }

func (b *testFanoutBroker) Publish(_ context.Context, _ string, message []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for subscriber := range b.subscribers {
		select {
		case subscriber <- append([]byte(nil), message...):
		default:
		}
	}
	return nil
}

func (b *testFanoutBroker) Subscribe(_ context.Context, _ string) (cache.PubSubSubscription, error) {
	channel := make(chan []byte, 8)
	b.mu.Lock()
	b.subscribers[channel] = struct{}{}
	b.mu.Unlock()
	return &testFanoutSubscription{broker: b, messages: channel}, nil
}

type testFanoutSubscription struct {
	broker   *testFanoutBroker
	messages chan []byte
	once     sync.Once
}

func (s *testFanoutSubscription) Messages() <-chan []byte { return s.messages }
func (s *testFanoutSubscription) Close() error {
	s.once.Do(func() {
		s.broker.mu.Lock()
		delete(s.broker.subscribers, s.messages)
		close(s.messages)
		s.broker.mu.Unlock()
	})
	return nil
}

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

func TestHubFansGameEventsAcrossInstancesWithoutEcho(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	broker := newTestFanoutBroker()
	first := NewHubWithPresence(logger, broker)
	second := NewHubWithPresence(logger, broker)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := first.EnableGameEventFanout(ctx); err != nil {
		t.Fatal(err)
	}
	if err := second.EnableGameEventFanout(ctx); err != nil {
		t.Fatal(err)
	}
	firstClient := &Client{send: make(chan []byte, 4), userID: "white", gameID: "game", logger: logger}
	secondClient := &Client{send: make(chan []byte, 4), userID: "black", gameID: "game", logger: logger}
	first.register(firstClient)
	second.register(secondClient)
	first.Publish("game", Event{Type: EventMove, GameID: "game"})
	select {
	case <-secondClient.send:
	case <-time.After(time.Second):
		t.Fatal("remote hub did not receive game event")
	}
	select {
	case <-firstClient.send:
	case <-time.After(time.Second):
		t.Fatal("local hub did not receive its own game event")
	}
	select {
	case duplicate := <-firstClient.send:
		t.Fatalf("publisher received an echoed duplicate event: %s", duplicate)
	case <-time.After(25 * time.Millisecond):
	}
	first.unregister(firstClient)
	second.unregister(secondClient)
	first.CloseAll()
	second.CloseAll()
}
