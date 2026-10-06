package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthRoute(t *testing.T) {
	response := httptest.NewRecorder()
	NewRouter().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !json.Valid(response.Body.Bytes()) {
		t.Fatalf("health response is not valid JSON: %q", response.Body.String())
	}
}

func TestUnknownRouteUsesErrorEnvelope(t *testing.T) {
	response := httptest.NewRecorder()
	NewRouter().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/unknown", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}

	if !json.Valid(response.Body.Bytes()) {
		t.Fatalf("error response is not valid JSON: %q", response.Body.String())
	}
}

func TestAuthUnavailableWithoutDependencies(t *testing.T) {
	response := httptest.NewRecorder()
	router := NewRouter(func(mux *http.ServeMux) {
		mux.HandleFunc("/api/v1/auth/", func(w http.ResponseWriter, _ *http.Request) {
			WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "database unavailable")
		})
	})
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}
