package main

import (
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerDeleteBroll(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusBadRequest, "couldn't find acc_token", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "invalid or expired token", nil)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userUUID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "coulnd't find user", err)
		return
	}

	brollIDStr := r.PathValue("brollID")
	if brollIDStr == "" {
		httpapi.RespondWithError(w, http.StatusBadRequest, "missing broll id", nil)
		return
	}

	broll, err := cfg.db.GetBrollById(r.Context(), brollIDStr)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't find broll", err)
		return
	}

	if broll.UserID != user.ID {
		httpapi.RespondWithError(w, http.StatusForbidden, "request forbidden", nil)
		return
	}

	if broll.S3Url.String != "" {
		key, err := httpapi.GetKey(broll.S3Url.String)
		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get key", err)
			return
		}

		_, err = cfg.s3Client.DeleteObject(r.Context(), &s3.DeleteObjectInput{
			Bucket: aws.String(cfg.s3Bucket),
			Key:    aws.String(key),
		})

		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't delete s3 object", err)
		}
	}

	log.Printf("broll: %s\n", broll.ID)
	log.Printf("user: %s\n", user.ID)

	deletedBroll, err := cfg.db.DeleteBroll(r.Context(), database.DeleteBrollParams{
		ID:     broll.ID,
		UserID: user.ID,
	})

	if err != nil {
		log.Printf("error: %s", err.Error())
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't delete broll", err)
		return
	}

	httpapi.RespondWithJSON(w, http.StatusOK, deletedBroll)

}
