package httpapi

import (
	"net/http"
)

// NewRouter creates the base API routes and invokes feature route registrars.
func NewRouter(registerRoutes ...func(*http.ServeMux)) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)
	for _, register := range registerRoutes {
		register(mux)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		WriteError(w, http.StatusNotFound, "not_found", "route not found")
	})

	return mux
}
