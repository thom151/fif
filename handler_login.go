package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	loginParams := parameters{}
	err := decoder.Decode(&loginParams)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), loginParams.Email)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't find email", err)
		return
	}

	err = auth.CheckPasswordHash(loginParams.Password, user.PasswordHash)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "invalid password", err)
		return
	}
	userUUID, err := uuid.Parse(user.ID)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't parse user uuid", err)
		return
	}

	accessToken, err := auth.MakeJWT(userUUID, cfg.jwtSecret, time.Hour)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't create access token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour).Format(time.RFC3339),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't save refresh token", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        userUUID,
			CreatedAt: user.CreatedAt.Time.String(),
			UpdatedAt: user.UpdatedAt.Time.String(),
			Email:     user.Email,
		},
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}
