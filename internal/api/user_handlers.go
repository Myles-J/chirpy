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
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func CreateUserHandler(db *database.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Bad Request", err)
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
			return
		}

		dbUser, err := db.CreateUser(r.Context(), database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Could not create user", err)
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, User{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		})
	}
}
