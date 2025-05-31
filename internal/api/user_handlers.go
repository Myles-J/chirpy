package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Myles-J/chirpy/internal/auth"
	"github.com/Myles-J/chirpy/internal/database"
	"github.com/Myles-J/chirpy/internal/utils"
	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type requestParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func CreateUserHandler(db *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestParams requestParams

		err := json.NewDecoder(r.Body).Decode(&requestParams)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
			return
		}

		hashedPassword, err := auth.HashPassword(requestParams.Password)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
			return
		}

		dbUser, err := db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          requestParams.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, User{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
			Email:       dbUser.Email,
			IsChirpyRed: dbUser.IsChirpyRed,
		})
	}
}

func UpdateUserHandler(db *database.Queries, tokenSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Authenticate user
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

		// Decode request
		var params requestParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
			return
		}

		// Hash password only if provided
		hashedPassword := ""
		if params.Password != "" {
			var hashErr error
			hashedPassword, hashErr = auth.HashPassword(params.Password)
			if hashErr != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", hashErr)
				return
			}
		}

		// Update user
		dbUser, err := db.UpdateUser(r.Context(), database.UpdateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
			ID:             userID,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not update user", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, User{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			UpdatedAt:   dbUser.UpdatedAt,
			Email:       dbUser.Email,
			IsChirpyRed: dbUser.IsChirpyRed,
		})
	}
}
