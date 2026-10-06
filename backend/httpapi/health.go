package httpapi

import (
	"log/slog"
	"net/http"
)

type healthResponse struct {
	Status string `json:"status"`
}

func healthHandler(status string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			if err := WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed"); err != nil {
				slog.Error("write API response", "error", err)
			}
			return
		}

		if err := WriteJSON(w, http.StatusOK, healthResponse{Status: status}); err != nil {
			slog.Error("write health response", "error", err)
		}
	}
}
