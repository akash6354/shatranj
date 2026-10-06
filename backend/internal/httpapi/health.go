package httpapi

import (
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		WriteError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	if err := WriteJSON(w, http.StatusOK, struct {
		Status string `json:"status"`
	}{Status: "ok"}); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
