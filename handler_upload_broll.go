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
	"github.com/thom151/fif/internal/httpapi"
	"github.com/thom151/fif/internal/media"
)

func (cfg *apiConfig) handlerUploadBroll(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<30)
	brollIDString := r.PathValue("brollID")
	brollID, err := uuid.Parse(brollIDString)
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

	broll, err := cfg.db.GetBrollById(r.Context(), brollID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get broll", err)
		return
	}

	if user.ID != broll.UserID {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "unauthorized access", err)
		return
	}

	file, header, err := r.FormFile("broll")
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

	if mediaType != "video/mp4" {
		httpapi.RespondWithError(w, http.StatusBadRequest, "invalid media type", err)
		return
	}

	tempDir := os.TempDir()
	brollFile, err := os.CreateTemp(tempDir, "fif-broll-upload-*.mp4")
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error saving file", err)
		return
	}

	defer os.Remove(brollFile.Name())
	defer brollFile.Close()
	log.Println("created temp successful")

	_, err = io.Copy(brollFile, file)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error copying broll to file", err)
		return
	}

	transcodedBroll, err := media.TranscodeH264(brollFile.Name())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't transcode file", err)
		return
	}
	defer os.Remove(transcodedBroll)

	_, err = brollFile.Seek(0, io.SeekStart)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "Could not reset file pointer", err)
		return
	}

	key := assets.GetAssestPath(mediaType)
	key = filepath.Join(user.ID, "broll", key)

	processedBroll, err := media.ProcessVideoForFastStart(transcodedBroll)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't process video for fast start", err)
		return
	}
	defer os.Remove(processedBroll)

	processedBrollFile, err := os.Open(processedBroll)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't open processed broll", err)
		return
	}
	defer processedBrollFile.Close()

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.s3Bucket),
		Key:         aws.String(key),
		Body:        processedBrollFile,
		ContentType: aws.String(mediaType),
	})

	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error uploading file to s3", err)
		return
	}

	bucketKey := fmt.Sprintf("%s,%s", cfg.s3Bucket, key)
	broll.S3Url = sql.NullString{String: bucketKey, Valid: true}

	err = cfg.db.UpdateBroll(r.Context(), database.UpdateBrollParams{
		Title:       broll.Title,
		Description: broll.Description,
		S3Url:       broll.S3Url,
		UserID:      broll.UserID,
		ID:          broll.ID,
	})

	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error updating broll url", err)
		return
	}

	broll, err = httpapi.DbBrollToSignedBroll(broll, cfg.s3Client)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get signed broll", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, broll)

}

func (cfg *apiConfig) handlerGetUploadBrollPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpapi.RespondWithError(w, http.StatusMethodNotAllowed, "method not allowed", nil)
		return
	}
	httpapi.RenderTemplate(w, "broll_upload", nil)
	return
}
