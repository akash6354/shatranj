package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shatranj/backend/internal/auth"
)

func TestConfiguredRoutesMount(t *testing.T) {
	tokens, err := auth.NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	handler := newHandler(new(sql.DB), tokens)
	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodPost, path: "/api/v1/auth/register", body: "{", want: http.StatusBadRequest},
		{method: http.MethodPost, path: "/api/v1/auth/login", body: "{", want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/users/me", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/profiles/me", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/auth/validate", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/puzzles", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/puzzles/daily", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/lessons", want: http.StatusUnauthorized},
		{method: http.MethodPut, path: "/api/v1/lessons/chapters/00000000-0000-0000-0000-000000000000/progress", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/tournaments", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/tournaments", body: "{", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/clubs", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/clubs", body: "{", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/friends", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/friends/requests", body: "{", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/chat/rooms", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/chat/rooms/44444444-4444-4444-8444-444444444444/ws", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/achievements/me", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/notifications", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/notifications/unread-count", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/subscriptions/me", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/subscriptions/plans", want: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/subscriptions/entitlements/advanced_puzzles", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/payments/orders", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/payments", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/payments/razorpay/webhook", body: "{}", want: http.StatusServiceUnavailable},
		{method: http.MethodPut, path: "/api/v1/coaches/me", body: "{", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/coaches/55555555-5555-4555-8555-555555555555/bookings", body: "{", want: http.StatusUnauthorized},
		{method: http.MethodPut, path: "/api/v1/admin/news/66666666-6666-4666-8666-666666666666/status", body: "{", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/admin/news", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/admin/dashboard", want: http.StatusUnauthorized},
		{method: http.MethodPut, path: "/api/v1/admin/users/77777777-7777-4777-8777-777777777777/status", body: "{", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/admin/payments", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/admin/payment-flags", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/games/33333333-3333-4333-8333-333333333333/review", want: http.StatusUnauthorized},
		{method: http.MethodGet, path: "/api/v1/games/33333333-3333-4333-8333-333333333333/review", want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/games/33333333-3333-4333-8333-333333333333/draw-claims", body: `{"reason":"fifty_move_rule"}`, want: http.StatusUnauthorized},
		{method: http.MethodPost, path: "/api/v1/auth/otp/request", want: http.StatusNotImplemented},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d (body %q)", response.Code, test.want, response.Body.String())
			}
		})
	}
}

func TestUnconfiguredDatabaseKeepsHealthAndReportsFeatureUnavailable(t *testing.T) {
	handler := newHandler(nil, nil)
	for _, test := range []struct {
		path string
		want int
	}{
		{path: "/healthz", want: http.StatusOK},
		{path: "/api/v1/games", want: http.StatusServiceUnavailable},
		{path: "/api/v1/admin/dashboard", want: http.StatusServiceUnavailable},
		{path: "/api/v1/news", want: http.StatusServiceUnavailable},
	} {
		t.Run(test.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != test.want {
				t.Fatalf("status = %d, want %d (body %q)", response.Code, test.want, response.Body.String())
			}
			if test.want == http.StatusServiceUnavailable && !strings.Contains(response.Body.String(), `"error"`) {
				t.Fatalf("unavailable response is not JSON error shape: %q", response.Body.String())
			}
		})
	}
}
