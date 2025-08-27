package main

import (
	"net/http"

	"github.com/thom151/fif/internal/auth"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/fifS3"
	"github.com/thom151/fif/internal/httpapi"
)

type fif_brolls struct {
	UserID string
	Brolls []database.Broll
}

func (cfg *apiConfig) handlerGetBrolls(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header, r.Cookies())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "invalid or missing token", err)
		return
	}

	userUUID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusUnauthorized, "invalid or expired token", nil)
		return
	}

	user, err := cfg.db.GetUserByID(r.Context(), userUUID.String())
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "coudln't get user", nil)
		return
	}

	brolls, err := cfg.db.GetBrollByUser(r.Context(), user.ID)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't get  brolls", err)
		return
	}

	for i := range brolls {

		if !brolls[i].S3Url.Valid || brolls[i].S3Url.String == "" {
			continue
		}

		signedBroll, err := fifS3.DbBrollToSignedBroll(brolls[i], cfg.s3Client)
		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "couldn't sign brolls", err)
			return
		}

		brolls[i] = signedBroll
	}

	brollsResponse := fif_brolls{
		UserID: user.ID,
		Brolls: brolls,
	}

	httpapi.RenderTemplate(w, "brolls", brollsResponse)

}
