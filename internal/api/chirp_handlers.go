package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Myles-J/chirpy/internal/auth"
	"github.com/Myles-J/chirpy/internal/database"
	"github.com/Myles-J/chirpy/internal/utils"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func CreateChirpHandler(db *database.Queries, tokenSecret string) http.HandlerFunc {
	const maxChirpLength = 140
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	type RequestPayload struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
			return
		}

		userID, err := auth.ValidateJWT(token, tokenSecret)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
			return
		}

		var requestPayload RequestPayload
		if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
			return
		}

		if len(requestPayload.Body) > maxChirpLength {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", errors.New("chirp is too long"))
			return
		}

		cleanedBody := getCleanedBody(requestPayload.Body, badWords)

		dbChirp, err := db.CreateChirp(context.Background(), database.CreateChirpParams{
			Body:   cleanedBody,
			UserID: userID,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not create chirp", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    userID,
		})
	}
}

func ListChirpsHandler(db *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbChirps, err := db.ListChirps(context.Background())
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
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

		utils.RespondWithJSON(w, http.StatusOK, chirps)
	}
}

func GetChirpHandler(db *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
			return
		}

		dbChirp, err := db.GetChirp(context.Background(), id)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Chirp not found", err)
			return
		}

		chirp := Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		}

		utils.RespondWithJSON(w, http.StatusOK, chirp)
	}
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}
