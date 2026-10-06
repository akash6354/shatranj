package middleware

import (
	"log/slog"
	"net/http"

	"github.com/shatranj/backend/internal/httpapi"
)

// Recovery logs panics and sends a structured internal error response.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.ErrorContext(r.Context(), "recovered HTTP panic",
						"request_id", RequestIDFromContext(r.Context()),
						"panic", recovered,
					)
					httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
