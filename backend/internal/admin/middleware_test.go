package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shatranj/backend/internal/auth"
)

type roleCheckerFunc func(context.Context, string) (bool, error)

func (f roleCheckerFunc) IsAdmin(ctx context.Context, userID string) (bool, error) {
	return f(ctx, userID)
}

func TestAdminOnlyRequiresValidAdminPrincipal(t *testing.T) {
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, _, err := tokens.Issue(auth.User{ID: "00000000-0000-4000-8000-000000000001"})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		admin  bool
		header string
		want   int
	}{
		{name: "missing token", want: http.StatusUnauthorized},
		{name: "ordinary user", admin: false, header: "Bearer " + token, want: http.StatusForbidden},
		{name: "admin", admin: true, header: "Bearer " + token, want: http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			checker := roleCheckerFunc(func(context.Context, string) (bool, error) { return test.admin, nil })
			handler := AdminOnly(tokens, checker)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}))
			request := httptest.NewRequest(http.MethodGet, "/admin", nil)
			request.Header.Set("Authorization", test.header)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d", response.Code, test.want)
			}
		})
	}
}
