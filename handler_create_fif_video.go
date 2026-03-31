package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	//	"github.com/aws/aws-sdk-go-v2/aws"
	//	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/thom151/fif/internal/assets"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/dropbox"
	"github.com/thom151/fif/internal/fifS3"
	"github.com/thom151/fif/internal/fifYouTube"
	"github.com/thom151/fif/internal/formulas"
	"github.com/thom151/fif/internal/heygen"
	"github.com/thom151/fif/internal/httpapi"
	"github.com/thom151/fif/internal/media"
	"github.com/thom151/fif/internal/openai"
)

type fifVideoParameters struct {
	BrollID       string `json:"broll_id"`
	MusicID       string `json:"music_id"`
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
	log.Println(userUUID.String(), user.Username)

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

	music, err := cfg.db.GetMusicById(r.Context(), fifVideoParams.MusicID)
	if err != nil {
		log.Printf("err: %v", err)
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get fif", err)
		return
	}

	log.Printf("got music url : %s\n", music.S3Url.String)

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
	emptyMusicOutPath := filepath.Join(base, "music.mp3")

	empttyFifOutPath := filepath.Join(base, "fif.mp4")

	//GENERATE HEYGEN THEN DOWNLOAD IN GET THE FILENAME
	avatarOutPath, err := heygen.GenerateAndDownloadAvatar(r.Context(), cfg.heygenApiKey, fifScript.FullScript, user.AvatarUrl.String, user.VoiceUrl.String, fif.Title, emptyAvatarOutPath)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't generate avatar", err)
		return
	}
	defer os.Remove(avatarOutPath)
	//DOWNLOAD THE BROLL
	brollOutPath, err := fifS3.DownloadAssetFromS3(r.Context(), broll.S3Url.String, cfg.s3Bucket, emptyBrollOutPath)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't download broll", err)
		return
	}
	defer os.Remove(brollOutPath)

	musicOutPath, err := fifS3.DownloadAssetFromS3(r.Context(), music.S3Url.String, cfg.s3Bucket, emptyMusicOutPath)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't download music", err)
		return
	}
	defer os.Remove(musicOutPath)

	//CONCATENATE HEYGEN + BROLL
	finalPath, err := formulas.FormulaV1_1(r.Context(), cfg.deepgramApiKey, base, avatarOutPath, brollOutPath, empttyFifOutPath, musicOutPath, fifScript.CutIndex)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't formulate", err)
		return
	}
	defer os.Remove(finalPath)

	log.Printf("FiF path for (%s) : %s\n", user.Username, finalPath)

	mediaType := "video/mp4"
	key := assets.GetAssestPath(mediaType)
	key = filepath.Join(user.ID, "fif", key)

	processedFiF, err := media.ProcessVideoForFastStart(finalPath)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't process fif", err)
		return
	}

	processedFiFFile, err := os.Open(processedFiF)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't open processed broll", err)
		return
	}
	defer processedFiFFile.Close()
	defer os.Remove(processedFiF)

	opCtx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	/*
		log.Printf("trying to upload fif to s3")
		_, err = cfg.s3Client.PutObject(opCtx, &s3.PutObjectInput{
			Bucket:      aws.String(cfg.s3Bucket),
			Key:         aws.String(key),
			Body:        processedFiFFile,
			ContentType: aws.String(mediaType),
		})

		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "error uploading file to s3", err)
			return
		}
		log.Printf("fif successfully uploaded")

		urlCdn := fmt.Sprintf("%s/%s", cfg.s3CfDistribution, key)
	*/

	finalFilePath := strings.TrimSuffix(processedFiF, ".processing")

	if err := os.Rename(processedFiF, finalFilePath); err != nil {
		log.Fatal("failed to rename processed file:", err)
	}

	//dropboxFolder := filepath.Join(user.Email, time.Now().Format("02-01-2006"), fifVideoParams.ClientAddress)

	if time.Now().After(cfg.dropboxAccTokenExpiresAt) {
		newAccTok, err := dropbox.GetNewAccessToken(cfg.dropboxRefreshToken, cfg.dropboxClientID, cfg.dropboxClientSecret)
		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "error getting acces token", err)
			return
		}
		cfg.dropboxAccToken = newAccTok.AccessToken
	}
	/*
		link, err := dropbox.UploadToDropbox(finalFilePath, dropboxFolder, cfg.dropboxAccToken)
		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "error uploading to dropbox", err)
			return
		}

	*/

	youtubeLink, err := fifYouTube.UploadVideo(finalFilePath, fif.Title, fif.Description.String)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error uploading to youtube", err)
		return
	}

	fif.S3Url = sql.NullString{String: youtubeLink, Valid: true}

	_, err = cfg.db.UpdateFif(opCtx, database.UpdateFifParams{
		Title:       fif.Title,
		Description: fif.Description,
		S3Url:       fif.S3Url,
		UserID:      fif.UserID,
		ID:          fif.ID,
	})

	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error updating broll url", err)
		return
	}

	log.Printf("fif url successfuly updated")

	/*
		fif, err = fifS3.DbFiFToSignedFiF(fif, cfg.s3Client)
		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get signed broll", err)
			return
		}
	*/

	log.Printf("fif link: %s\n", fif.S3Url.String)
	httpapi.RespondWithJSON(w, http.StatusOK, fif)

}
