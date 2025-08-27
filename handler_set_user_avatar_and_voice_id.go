package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerSetUserAvatarAndVoiceID(w http.ResponseWriter, r *http.Request) {
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
		VoiceID  string `json:"voice_id"`
	}

	decoder := json.NewDecoder(r.Body)
	var avatarAndVoiceParams params
	err = decoder.Decode(&avatarAndVoiceParams)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't decode params", err)
		return
	}

	updatedUser, err := cfg.db.SetUserAvatarAndVoiceID(r.Context(), database.SetUserAvatarAndVoiceIDParams{
		AvatarUrl: sql.NullString{String: avatarAndVoiceParams.AvatarID, Valid: avatarAndVoiceParams.AvatarID != ""},
		VoiceUrl:  sql.NullString{String: avatarAndVoiceParams.VoiceID, Valid: avatarAndVoiceParams.VoiceID != ""},
		UpdatedAt: sql.NullTime{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		ID: user.ID,
	})
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't set user avatar_id", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, updatedUser)

}
