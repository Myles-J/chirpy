package api

import (
	"net/http"
)

func HandleHealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
