package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/thom151/fif/internal/assets"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/fifS3"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerUploadMusic(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	musicIDString := r.PathValue("musicID")
	musicID, err := uuid.Parse(musicIDString)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't parse broll uuid", err)
		return
	}

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

	music, err := cfg.db.GetMusicById(r.Context(), musicID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get broll", err)
		return
	}

	if user.ID != music.UserID {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "unauthorized access", err)
		return
	}

	file, header, err := r.FormFile("music")
	if err != nil {
		httpapi.RespondWithError(w, http.StatusBadRequest, "couldn't find broll", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		httpapi.RespondWithError(w, http.StatusBadRequest, "invalid content-type", err)
		return
	}

	if mediaType != "audio/mpeg" {
		httpapi.RespondWithError(w, http.StatusBadRequest, "invalid media type", err)
		return
	}

	tempDir := os.TempDir()
	musicFile, err := os.CreateTemp(tempDir, "fif-music-upload-*.mp3")
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error saving file", err)
		return
	}

	defer os.Remove(musicFile.Name())
	defer musicFile.Close()
	log.Println("created temp successful")

	_, err = io.Copy(musicFile, file)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error copying broll to file", err)
		return
	}

	_, err = musicFile.Seek(0, io.SeekStart)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "Could not reset file pointer", err)
		return
	}

	key := assets.GetAssestPath(mediaType)
	key = filepath.Join(user.ID, "music", key)

	openedMusicFile, err := os.Open(musicFile.Name())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't open processed broll", err)
		return
	}
	defer openedMusicFile.Close()

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         aws.String(key),
		Body:        openedMusicFile,
		ContentType: aws.String(mediaType),
	})

	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error uploading file to s3", err)
		return
	}

	bucketKey := fmt.Sprintf("%s,%s", cfg.s3Bucket, key)
	music.S3Url = sql.NullString{String: bucketKey, Valid: true}

	err = cfg.db.UpdateMusic(r.Context(), database.UpdateMusicParams{
		Title:       music.Title,
		Description: music.Description,
		S3Url:       music.S3Url,
		UserID:      music.UserID,
		ID:          music.ID,
	})

	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error updating broll url", err)
		return
	}

	music, err = fifS3.DbMusicToSignedMusic(music, cfg.s3Client)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get signed broll", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, music)

}

func (cfg *apiConfig) handlerGetUploadMusicPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpapi.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	httpapi.RenderTemplate(w, "music_upload", nil)
	return
}
