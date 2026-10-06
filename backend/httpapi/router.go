package httpapi

import (
	"log/slog"
	"net/http"
)

// NewRouter creates the HTTP API router.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthHandler("ok"))
	mux.Handle("/readyz", healthHandler("ready"))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := WriteError(w, http.StatusNotFound, "not_found", "route not found"); err != nil {
			slog.Error("write API error response", "error", err)
		}
	})
	return mux
}
