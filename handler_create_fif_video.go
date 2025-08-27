package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/fifS3"
	"github.com/thom151/fif/internal/formulas"
	"github.com/thom151/fif/internal/heygen"
	"github.com/thom151/fif/internal/httpapi"
	"github.com/thom151/fif/internal/openai"
)

type fifVideoParameters struct {
	BrollID       string `json:"broll_id"`
	AgentName     string `json:"agent_name"`
	ClientName    string `json:"client_name"`
	ClientAddress string `json:"client_address"`
}

func (cfg *apiConfig) handlerCreateFifVideo(w http.ResponseWriter, r *http.Request) {
	fifID := r.PathValue("fifID")

	token, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusBadRequest, "missing token", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "invalid/expired token", nil)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userUUID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get user", err)
		return
	}

	if user.AvatarUrl.String == "" || user.VoiceUrl.String == "" {
		httpapi.RespondWithError(w, http.StatusBadRequest, "user has no avatar/voice configured", nil)
		return
	}

	log.Println(fifID)
	log.Println(userUUID.String())

	decoder := json.NewDecoder(r.Body)
	var fifVideoParams fifVideoParameters
	err = decoder.Decode(&fifVideoParams)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error decoding fif params", err)
		return
	}

	fif, err := cfg.db.GetFifById(r.Context(), fifID)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get fif", err)
		return
	}

	broll, err := cfg.db.GetBrollById(r.Context(), fifVideoParams.BrollID)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get broll", err)
		return
	}

	fifDetails := fmt.Sprintf("Agent Name: %s, Client Name: %s, Client Address: %s", fifVideoParams.AgentName, fifVideoParams.ClientName, fifVideoParams.ClientAddress)

	fifScript, err := openai.GenerateFifScript(r.Context(), cfg.openaiClient, fifDetails, cfg.openaiAssistantID)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "coulnd't generate fif script", err)
		return
	}

	//I use taskID to separate the tasks for each person and no duplicate files in temp
	taskID := uuid.New().String()
	base := filepath.Join(cfg.tempDir, user.ID, taskID)
	emptyAvatarOutPath := filepath.Join(base, "avatar.mp4")
	emptyBrollOutPath := filepath.Join(base, "broll.mp4")
	empttyFifOutPath := filepath.Join(base, "fif.mp4")

	//GENERATE HEYGEN THEN DOWNLOAD IN GET THE FILENAME
	avatarOutPath, err := heygen.GenerateAndDownloadAvatar(r.Context(), cfg.heygenApiKey, fifScript, user.AvatarUrl.String, user.VoiceUrl.String, fif.Title, emptyAvatarOutPath)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't generate avatar", err)
		return
	}
	//DOWNLOAD THE BROLL
	brollOutPath, err := fifS3.DownloadAssetFromS3(r.Context(), broll.S3Url.String, cfg.s3Bucket, emptyBrollOutPath)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't download broll", err)
		return
	}

	//CONCATENATE HEYGEN + BROLL
	finalPath, err := formulas.FormulaV1(r.Context(), base, avatarOutPath, brollOutPath, empttyFifOutPath)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't formulate", err)
		return
	}

	log.Printf("FiF path: %s\n", finalPath)

}
