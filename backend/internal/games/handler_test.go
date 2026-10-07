package games

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type unusedRepository struct{}

func (unusedRepository) Create(context.Context, string, CreateInput) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) Join(context.Context, string, string) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) Get(context.Context, string) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) SaveMove(context.Context, string, string, string, MoveUpdate) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) Resign(context.Context, string, string) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) OfferDraw(context.Context, string, string) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) AcceptDraw(context.Context, string, string) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) ClaimDraw(context.Context, string, string, string) (Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) ExpireDueGames(context.Context, int) ([]Game, error) {
	panic("unexpected repository call")
}
func (unusedRepository) ClaimCompletionJobs(context.Context, int) ([]string, error) {
	panic("unexpected repository call")
}
func (unusedRepository) CompleteCompletionJob(context.Context, string) error {
	panic("unexpected repository call")
}
func (unusedRepository) RetryCompletionJob(context.Context, string, error) error {
	panic("unexpected repository call")
}

func TestInvalidMoveJSONReturnsClientError(t *testing.T) {
	handler := NewHandler(NewService(unusedRepository{}, nil))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/games/"+testGameID+"/moves", strings.NewReader("{"))
	request.SetPathValue("gameID", testGameID)
	response := httptest.NewRecorder()
	handler.Move(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}
