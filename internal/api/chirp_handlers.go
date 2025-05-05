package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Myles-J/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func CreateChirpHandler(db *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		badWords := map[string]bool{
			"kerfuffle": true,
			"sharbert":  true,
			"fornax":    true,
		}
		type parameters struct {
			Body   string    `json:"body"`
			UserID uuid.UUID `json:"user_id"`
		}

		type response struct {
			Valid       bool   `json:"valid"`
			CleanedBody string `json:"cleaned_body,omitempty"`
			Error       string `json:"error,omitempty"`
		}

		w.Header().Set("Content-Type", "application/json")

		decoder := json.NewDecoder(r.Body)

		params := parameters{}

		err := decoder.Decode(&params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp := response{
				Valid: false,
				Error: "Bad Request",
			}
			jsonResp, _ := json.Marshal(resp)
			w.Write(jsonResp)
			return
		}

		if len(params.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			resp := response{
				Valid: false,
				Error: "Chirp is too long",
			}
			jsonResp, _ := json.Marshal(resp)
			w.Write(jsonResp)
			return
		}

		words := strings.Split(params.Body, " ")
		cleanedWords := make([]string, len(words))

		for i, word := range words {
			if _, ok := badWords[strings.ToLower(word)]; ok {
				cleanedWords[i] = "****"
			} else {
				cleanedWords[i] = word
			}
		}

		cleanedBody := strings.Join(cleanedWords, " ")

		dbChirp, err := db.CreateChirp(context.Background(), database.CreateChirpParams{
			Body:   cleanedBody,
			UserID: params.UserID,
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		chirp := Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}

		w.Header().Set("Content-Type", "application/json")
		// Return the chirp along with a 201 Created status
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(chirp)
	}
}
