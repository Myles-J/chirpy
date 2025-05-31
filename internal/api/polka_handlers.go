package api

import (
	"encoding/json"
	"net/http"

	"github.com/Myles-J/chirpy/internal/auth"
	"github.com/Myles-J/chirpy/internal/database"
	"github.com/Myles-J/chirpy/internal/utils"
	"github.com/google/uuid"
)

func PolkaWebhookHandler(dbQueries *database.Queries, polkaSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r.Header)
		if err != nil || apiKey != polkaSecret {
			utils.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
			return
		}

		var payloadRequest struct {
			Event string `json:"event"`
			Data  struct {
				UserID uuid.UUID `json:"user_id"`
			} `json:"data"`
		}

		err = json.NewDecoder(r.Body).Decode(&payloadRequest)
		if err != nil {
			utils.RespondWithError(
				w,
				http.StatusBadRequest,
				"Invalid request format. Please ensure the request body is valid JSON with 'event' and 'data' fields.",
				err,
			)
			return
		}

		if payloadRequest.Event != "user.upgraded" {
			utils.RespondWithError(
				w,
				http.StatusNoContent,
				"Invalid event.",
				nil,
			)
			return
		}

		_, err = dbQueries.UpdateUserIsChirpyRed(r.Context(), database.UpdateUserIsChirpyRedParams{
			IsChirpyRed: true,
			ID:          payloadRequest.Data.UserID,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "User not found.", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusNoContent, nil)
	}
}
