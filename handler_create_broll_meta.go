package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerCreateBrollMeta(w http.ResponseWriter, r *http.Request) {

	token, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "couldn't find jwt", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "couldn't validate jwt", err)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userUUID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get user", err)
		return
	}

	type params struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	decoder := json.NewDecoder(r.Body)
	var brollMetaParams params
	err = decoder.Decode(&brollMetaParams)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't decode broll params", err)
		return
	}

	brollMeta, err := cfg.db.CreateBrollMeta(r.Context(), database.CreateBrollMetaParams{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		Title:       brollMetaParams.Title,
		Description: sql.NullString{String: brollMetaParams.Description, Valid: brollMetaParams.Description != ""}})

	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't create broll meta", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, brollMeta)

}
