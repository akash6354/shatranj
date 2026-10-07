package matchmaking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/akash6354/shatranj/backend/internal/games"
	"github.com/akash6354/shatranj/backend/internal/ratings"
)

type fakeRatings struct {
	values map[string]int
}

func (r fakeRatings) Get(_ context.Context, userID string, mode ratings.Mode) (ratings.Record, error) {
	if !mode.Valid() {
		return ratings.Record{}, ratings.ErrUnsupportedMode
	}
	return ratings.Record{UserID: userID, Mode: mode, Rating: r.values[userID]}, nil
}

type fakeQueue struct {
	entries map[string]Entry
}

func (q *fakeQueue) JoinQueue(_ context.Context, entry Entry) (Entry, *Entry, error) {
	if q.entries == nil {
		q.entries = make(map[string]Entry)
	}
	for _, existing := range q.entries {
		if existing.UserID == entry.UserID && existing.Status != StatusCanceled {
			return Entry{}, nil, ErrAlreadyQueued
		}
	}
	entry.ID, entry.CreatedAt, entry.Status = entry.UserID, time.Now(), StatusQueued
	for key, existing := range q.entries {
		if existing.Status == StatusQueued &&
			existing.Mode == entry.Mode &&
			existing.TimeControl == entry.TimeControl &&
			ratingsCompatible(entry.Rating, 0, existing.Rating, time.Since(existing.CreatedAt)) {
			matchID := "match-1"
			entry.Status, entry.MatchID = StatusMatched, matchID
			existing.Status, existing.MatchID = StatusMatched, matchID
			q.entries[key] = existing
			q.entries[entry.ID] = entry
			copy := existing
			return entry, &copy, nil
		}
	}
	q.entries[entry.ID] = entry
	return entry, nil, nil
}

func (q *fakeQueue) LeaveQueue(_ context.Context, userID string) error {
	for id, entry := range q.entries {
		if entry.UserID == userID && (entry.Status == StatusQueued || entry.Status == StatusMatched || entry.Status == StatusPaired) {
			entry.Status, entry.MatchID, entry.GameID = StatusCanceled, "", ""
			q.entries[id] = entry
			return nil
		}
	}
	return ErrQueueNotFound
}

func (q *fakeQueue) GetQueueStatus(_ context.Context, userID string) (Entry, error) {
	for _, entry := range q.entries {
		if entry.UserID == userID && (entry.Status == StatusQueued || entry.Status == StatusMatched || entry.Status == StatusPaired) {
			return entry, nil
		}
	}
	return Entry{}, ErrQueueNotFound
}

func (q *fakeQueue) CompleteMatch(_ context.Context, matchID, gameID string) error {
	for id, entry := range q.entries {
		if entry.MatchID == matchID && entry.Status == StatusMatched {
			entry.Status, entry.GameID = StatusPaired, gameID
			q.entries[id] = entry
		}
	}
	return nil
}

func (q *fakeQueue) AbortMatch(_ context.Context, matchID string) error {
	for id, entry := range q.entries {
		if entry.MatchID == matchID {
			entry.Status, entry.MatchID, entry.GameID = StatusCanceled, "", ""
			q.entries[id] = entry
		}
	}
	return nil
}

type fakeGames struct {
	players []string
}

func (g *fakeGames) Create(_ context.Context, userID string, input games.CreateInput) (games.Game, error) {
	g.players = append(g.players, userID)
	if !input.Rated || input.Mode != "blitz" {
		return games.Game{}, errors.New("expected rated blitz game")
	}
	return games.Game{ID: "created-game"}, nil
}

func (g *fakeGames) Join(_ context.Context, gameID, userID string) (games.Game, error) {
	g.players = append(g.players, userID)
	return games.Game{ID: gameID}, nil
}

func TestJoinQueueCreatesRatedGameForCompatiblePlayers(t *testing.T) {
	queue := &fakeQueue{}
	gameCreator := new(fakeGames)
	service := NewService(queue, fakeRatings{values: map[string]int{"first": 1200, "second": 1290}}, gameCreator)
	request := Request{Mode: ratings.ModeBlitz, TimeControl: games.TimeControl{InitialSeconds: 180, IncrementSeconds: 2}}

	entry, err := service.Join(context.Background(), "first", request)
	if err != nil || entry.Status != StatusQueued {
		t.Fatalf("first Join() = %+v, %v", entry, err)
	}
	entry, err = service.Join(context.Background(), "second", request)
	if err != nil {
		t.Fatalf("second Join() error = %v", err)
	}
	if entry.Status != StatusPaired || entry.GameID != "created-game" {
		t.Fatalf("matched entry = %+v", entry)
	}
	if len(gameCreator.players) != 2 || gameCreator.players[0] == gameCreator.players[1] ||
		(gameCreator.players[0] != "first" && gameCreator.players[0] != "second") ||
		(gameCreator.players[1] != "first" && gameCreator.players[1] != "second") {
		t.Fatalf("game players = %v", gameCreator.players)
	}
}

func TestJoinQueueRejectsUnsupportedSettingsAndDuplicate(t *testing.T) {
	queue := &fakeQueue{}
	service := NewService(queue, fakeRatings{values: map[string]int{"first": 1200}}, new(fakeGames))
	request := Request{Mode: ratings.ModeBlitz, TimeControl: games.TimeControl{InitialSeconds: 300}}
	if _, err := service.Join(context.Background(), "first", request); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Join(context.Background(), "first", request); !errors.Is(err, ErrAlreadyQueued) {
		t.Fatalf("duplicate join error = %v, want ErrAlreadyQueued", err)
	}
	request.Mode = ratings.ModePuzzle
	if _, err := service.Join(context.Background(), "second", request); !errors.Is(err, ErrInvalidQueue) {
		t.Fatalf("puzzle queue error = %v, want ErrInvalidQueue", err)
	}
}
