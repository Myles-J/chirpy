package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Myles-J/chirpy/internal/auth"
	"github.com/Myles-J/chirpy/internal/database"
	"github.com/Myles-J/chirpy/internal/utils"
)

func RefreshHandler(db *database.Queries, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
			return
		}

		user, err := db.GetUserFromRefreshToken(context.Background(), refreshToken)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Couldn't get user for refresh token", err)
			return
		}

		accessToken, err := auth.MakeJWT(user.ID, jwtSecret, time.Hour)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't validate token", err)
			return
		}

		if user.Token != refreshToken {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", errors.New("invalid refresh token"))
			return
		}

		type response struct {
			Token string `json:"token"`
		}

		utils.RespondWithJSON(w, http.StatusOK, response{
			Token: accessToken,
		})
	}
}
