package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	tests := []struct {
		path         string
		wantStatus   int
		wantBodyPart string
	}{
		{path: "/healthz", wantStatus: http.StatusOK, wantBodyPart: `"status":"ok"`},
		{path: "/readyz", wantStatus: http.StatusOK, wantBodyPart: `"status":"ready"`},
		{path: "/missing", wantStatus: http.StatusNotFound, wantBodyPart: `"code":"not_found"`},
	}
	handler := NewRouter()

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if response.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Fatalf("Content-Type = %q, want JSON", response.Header().Get("Content-Type"))
			}
			if !json.Valid(response.Body.Bytes()) {
				t.Fatalf("response is not valid JSON: %q", response.Body.String())
			}
			if got := response.Body.String(); !strings.Contains(got, tt.wantBodyPart) {
				t.Fatalf("response = %q, does not contain %q", got, tt.wantBodyPart)
			}
		})
	}
}

func TestHealthRouteRejectsNonGet(t *testing.T) {
	response := httptest.NewRecorder()
	NewRouter().ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/healthz", nil))

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if !json.Valid(response.Body.Bytes()) {
		t.Fatalf("response is not valid JSON: %q", response.Body.String())
	}
}
