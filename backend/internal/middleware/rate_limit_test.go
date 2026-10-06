package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitRejectsAfterLimit(t *testing.T) {
	handler := RateLimit(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	first := httptest.NewRequest(http.MethodPost, "/", nil)
	first.RemoteAddr = "192.0.2.10:1234"
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, first)
	if response.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", response.Code, http.StatusNoContent)
	}

	second := httptest.NewRequest(http.MethodPost, "/", nil)
	second.RemoteAddr = first.RemoteAddr
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, second)
	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
}
