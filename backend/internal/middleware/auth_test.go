package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticateAddsPrincipal(t *testing.T) {
	handler := Authenticate(func(token string) (Principal, error) {
		if token != "valid" {
			return Principal{}, errInvalidToken
		}
		return Principal{UserID: "user-1", Email: "user@example.com"}, nil
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := PrincipalFromContext(r.Context())
		if !ok || principal.UserID != "user-1" {
			t.Error("authenticated principal not attached to request")
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}

	request = httptest.NewRequest(http.MethodGet, "/", nil)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("missing-token status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

type testTokenError string

func (e testTokenError) Error() string { return string(e) }

var errInvalidToken = testTokenError("invalid token")
