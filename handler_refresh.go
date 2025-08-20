package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusBadRequest, "couldn't find token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), database.GetUserFromRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "couldn't get user for refresh token", err)
		return
	}

	userUUID, err := uuid.Parse(user.ID)
	accessToken, err := auth.MakeJWT(userUUID, cfg.jwtSecret, time.Hour)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "couldn't validate token", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, response{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusBadRequest, "coudln't find token", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), database.RevokeRefreshTokenParams{
		RevokedAt: sql.NullString{String: time.Now().UTC().Format(time.RFC3339)},
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Token:     refreshToken,
	})

	httpapi.RespondWithJSON(w, 204, nil)
}
