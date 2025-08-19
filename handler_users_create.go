package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	userParams := parameters{}

	err := decoder.Decode(&userParams)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't decode parameter", err)
		return
	}

	hashed, err := auth.HashPassword(userParams.Password)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error hashing password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New().String(),
		Username:     userParams.Username,
		PasswordHash: hashed,
		Email:        userParams.Email,
	})

	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error creating db user", err)
		return
	}

	log.Printf(user.Email + " has been created successfully")

	httpapi.RespondWithJSON(w, http.StatusOK, user)
}
