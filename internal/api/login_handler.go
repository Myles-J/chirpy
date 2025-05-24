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

func LoginHandler(db *database.Queries, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var requestPayload struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestPayload)
		if err != nil {
			// Client sent a request body that couldn't be parsed.
			utils.RespondWithError(
				w,
				http.StatusBadRequest,
				"Invalid request format. Please ensure the request body is valid JSON with 'email' and 'password' fields.",
				err,
			)
			return
		}

		dbUser, err := db.GetUserByEmail(r.Context(), requestPayload.Email)
		if err != nil {
			// Other database errors during user lookup are internal.
			// Log the internal error but tell the client there was an issue.
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not retrieve user information.", err)
			return
		}

		err = auth.CheckPassword(dbUser.HashedPassword, requestPayload.Password)
		if err != nil {
			// Password incorrect. Combine with user not found message for security.
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid email or password.", nil)
			return
		}

		accessToken, err := auth.MakeJWT(dbUser.ID, jwtSecret, 1*time.Hour)
		if err != nil {
			// Error creating access token. This is an internal system issue.
			utils.RespondWithError(
				w,
				http.StatusInternalServerError,
				"Could not generate access token. Please try again later.",
				err,
			)
			return
		}

		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			// Error creating refresh token. This is an internal system issue.
			utils.RespondWithError(
				w,
				http.StatusInternalServerError,
				"Could not generate refresh token. Please try again later.",
				err,
			)
			return
		}

		_, err = db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
			Token:     refreshToken,
			UserID:    dbUser.ID,
			ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
		})
		if err != nil {
			// Error saving refresh token. This is likely a database or system issue.
			utils.RespondWithError(
				w,
				http.StatusInternalServerError,
				"Could not save refresh token. Please try again later.",
				err,
			)
			return
		}

		// Successful login - Respond with user data and tokens
		type user struct {
			ID           uuid.UUID `json:"id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			RefreshToken string    `json:"refresh_token"`
		}

		utils.RespondWithJSON(w, http.StatusOK, user{
			ID:           dbUser.ID,
			CreatedAt:    dbUser.CreatedAt,
			UpdatedAt:    dbUser.UpdatedAt,
			Email:        dbUser.Email,
			Token:        accessToken,
			RefreshToken: refreshToken,
		})
	}
}
