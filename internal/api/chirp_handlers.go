package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
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

		userID, validateJWTError := auth.ValidateJWT(token, tokenSecret)
		if validateJWTError != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", validateJWTError)
			return
		}

		var requestPayload RequestPayload
		if decodeErr := json.NewDecoder(r.Body).Decode(&requestPayload); decodeErr != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", decodeErr)
			return
		}

		if len(requestPayload.Body) > maxChirpLength {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", errors.New("chirp is too long"))
			return
		}

		cleanedBody := getCleanedBody(requestPayload.Body, badWords)

		dbChirp, createChirpErr := db.CreateChirp(context.Background(), database.CreateChirpParams{
			Body:   cleanedBody,
			UserID: userID,
		})
		if createChirpErr != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not create chirp", createChirpErr)
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
		ctx := context.Background()
		query := r.URL.Query()
		authorIDStr := query.Get("author_id")
		sortParam := query.Get("sort")

		var (
			dbChirps []database.Chirp
			err      error
		)

		if authorIDStr != "" {
			parsedAuthorID, parseErr := uuid.Parse(authorIDStr)
			if parseErr != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", parseErr)
				return
			}
			dbChirps, err = db.ListChirpsByAuthor(ctx, parsedAuthorID)
		} else {
			dbChirps, err = db.ListChirps(ctx)
		}
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error", err)
			return
		}

		switch sortParam {
		case "desc":
			sort.Slice(dbChirps, func(i, j int) bool {
				return dbChirps[i].CreatedAt.After(dbChirps[j].CreatedAt)
			})
		case "asc":
			sort.Slice(dbChirps, func(i, j int) bool {
				return dbChirps[i].CreatedAt.Before(dbChirps[j].CreatedAt)
			})
		}

		chirps := make([]Chirp, len(dbChirps))
		for i := range dbChirps {
			chirps[i] = Chirp{
				ID:        dbChirps[i].ID,
				CreatedAt: dbChirps[i].CreatedAt,
				UpdatedAt: dbChirps[i].UpdatedAt,
				Body:      dbChirps[i].Body,
				UserID:    dbChirps[i].UserID,
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

func DeleteChirpHandler(db *database.Queries, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		chirpID, err := uuid.Parse(idStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
			return
		}
		token, tokenErr := auth.GetBearerToken(r.Header)
		if tokenErr != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", tokenErr)
			return
		}

		userID, validateJWTError := auth.ValidateJWT(token, tokenSecret)
		if validateJWTError != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", validateJWTError)
			return
		}
		dbChirp, err := db.GetChirp(context.Background(), chirpID)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Chirp not found", err)
			return
		}

		if dbChirp.UserID != userID {
			utils.RespondWithError(
				w,
				http.StatusForbidden,
				"Forbidden",
				errors.New("you are not the owner of this chirp"),
			)
			return
		}

		err = db.DeleteChirp(context.Background(), database.DeleteChirpParams{
			ID:     chirpID,
			UserID: userID,
		})
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusNoContent)
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
	return strings.Join(words, " ")
}
