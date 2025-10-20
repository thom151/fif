package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/thom151/fif/internal/dropbox"
	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerTestUpload(w http.ResponseWriter, r *http.Request) {
	dropboxFolder := "/test"
	filePath := "Whittlesea.mp4"

	var dropboxPath string
	var err error
	for {
		dropboxPath, err = dropbox.UploadToDropbox(filePath, dropboxFolder, cfg.dropboxAccToken)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			if strings.Contains(err.Error(), "expired_access_token") {
				newAccTok, err := dropbox.GetNewAccessToken(cfg.dropboxRefreshToken, cfg.dropboxClientID, cfg.dropboxClientSecret)
				if err != nil {
					httpapi.RespondWithError(w, http.StatusInternalServerError, "error getting acces token", err)
					return
				}
				cfg.dropboxAccToken = newAccTok.AccessToken
				continue
			} else {
				httpapi.RespondWithError(w, http.StatusInternalServerError, "error uploading to dropbox", err)
				return

			}
		}
		break
	}

	for {
		link, err := dropbox.GetDropboxLink(dropboxPath, cfg.dropboxAccToken)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
			if strings.Contains(err.Error(), "expired_access_token") {
				newAccTok, err := dropbox.GetNewAccessToken(cfg.dropboxRefreshToken, cfg.dropboxClientID, cfg.dropboxClientSecret)
				if err != nil {
					httpapi.RespondWithError(w, http.StatusInternalServerError, "error getting acces token", err)
					return
				}
				cfg.dropboxAccToken = newAccTok.AccessToken
				continue
			} else {

				httpapi.RespondWithError(w, http.StatusInternalServerError, "error getting dropbox link", err)
				return
			}
		}
		fmt.Printf("link: %s", link)
		break
	}

	return
}

func (cfg *apiConfig) handlerTestUploadV2(w http.ResponseWriter, r *http.Request) {
	dropboxFolder := "/test"
	filePath := "Whittlesea.mp4"

	var dropboxPath string
	var err error
	if time.Now().After(cfg.dropboxAccTokenExpiresAt) {
		newAccTok, err := dropbox.GetNewAccessToken(cfg.dropboxRefreshToken, cfg.dropboxClientID, cfg.dropboxClientSecret)
		if err != nil {
			httpapi.RespondWithError(w, http.StatusInternalServerError, "error getting acces token", err)
			return
		}
		cfg.dropboxAccToken = newAccTok.AccessToken

	}
	dropboxPath, err = dropbox.UploadToDropbox(filePath, dropboxFolder, cfg.dropboxAccToken)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error uploading to dropbox", err)
		return
	}

	link, err := dropbox.GetDropboxLink(dropboxPath, cfg.dropboxAccToken)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err.Error())
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error getting dropbox link", err)
		return
	}
	fmt.Printf("link: %s", link)

	return
}
