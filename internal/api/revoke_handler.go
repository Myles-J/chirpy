package api

import (
	"context"
	"net/http"

	"github.com/Myles-J/chirpy/internal/auth"
	"github.com/Myles-J/chirpy/internal/database"
	"github.com/Myles-J/chirpy/internal/utils"
)

func RevokeHandler(db *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
			return
		}

		err = db.RevokeRefreshToken(context.Background(), refreshToken)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
