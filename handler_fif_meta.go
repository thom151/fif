package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerFifMeta(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't find token", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "invalid/missing token", err)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userUUID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get user", err)
		return
	}

	type params struct {
		Title         string `json:"title"`
		Personalized  string `json:"personalize"`
		ClientName    string `json:"client_name"`
		ClientAddress string `json:"client_address"`
	}

	decoder := json.NewDecoder(r.Body)
	var fifParams params
	err = decoder.Decode(&fifParams)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error decoding fif parameters", err)
		return
	}

	fifMeta, err := cfg.db.CreateFifMeta(r.Context(), database.CreateFifMetaParams{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		Title:       fifParams.Title,
		Description: sql.NullString{String: fifParams.Personalized, Valid: fifParams.Personalized != ""},
	})
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't make fif meta", err)
		return
	}
	log.Printf("fif meta created: %s\n", fifMeta.ID)

	httpapi.RespondWithJSON(w, http.StatusOK, fifMeta)

}
