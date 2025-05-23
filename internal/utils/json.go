package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// RespondWithError responds to the client with a JSON payload representing an error.
// It logs the error and the response code if it's a server error (5XX).
func RespondWithError(w http.ResponseWriter, code int, message string, err error) {
	if err != nil {
		slog.Error("An error occurred", "error", err)
	}

	if code > 499 {
		slog.Error("Responding with 5XX error: ", "error", code)
	}

	type errorResponse struct {
		Error string `json:"error"`
	}

	RespondWithJSON(w, code, errorResponse{
		Error: message,
	})
}

// RespondWithJSON responds to the client with a JSON payload.
// It sets the Content-Type header to application/json, marshals the payload to JSON,
// and writes the JSON data to the response writer.
// If there is an error marshalling the JSON, it logs the error and responds with a 500 Internal Server Error.
func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Error marshalling JSON", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}
