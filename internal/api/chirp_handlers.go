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

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response{Valid: false, Error: "Bad Request"})
			return
		}

		if len(params.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response{Valid: false, Error: "Chirp is too long"})
			return
		}

		words := strings.Split(params.Body, " ")
		for i, word := range words {
			if _, ok := badWords[strings.ToLower(word)]; ok {
				words[i] = "****"
			}
		}
		cleanedBody := strings.Join(words, " ")

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

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(chirp)
	}
}

func ListChirpsHandler(db *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbChirps, err := db.ListChirps(context.Background())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		chirps := make([]Chirp, len(dbChirps))
		for i, dbChirp := range dbChirps {
			chirps[i] = Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.UpdatedAt,
				Body:      dbChirp.Body,
				UserID:    dbChirp.UserID,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chirps)
	}
}