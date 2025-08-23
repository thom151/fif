package main

import (
	"database/sql"
	"encoding/json"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerSetUserAvatarID(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusBadRequest, "missing token", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "invalid/expired token", err)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userUUID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get user", err)
		return
	}

	type params struct {
		AvatarID string `json:"avatar_id"`
	}

	decoder := json.NewDecoder(r.Body)
	var avatarIDparams params
	err = decoder.Decode(&avatarIDparams)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't decode params", err)
		return
	}

	log.Printf("setting avatar url for %s: %s", user.Username, avatarIDparams.AvatarID)

	updatedUser, err := cfg.db.SetUserAvatarID(r.Context(), database.SetUserAvatarIDParams{
		AvatarUrl: sql.NullString{String: avatarIDparams.AvatarID, Valid: avatarIDparams.AvatarID != ""},
		ID:        user.ID,
	})
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't set user avatar_id", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, updatedUser)

}
