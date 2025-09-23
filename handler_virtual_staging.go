package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/thom151/fif/internal/httpapi"
)

func (cfg *apiConfig) handlerVritualStaging(w http.ResponseWriter, r *http.Request) {

	type params struct {
		TextContent string `json:"text_content"`
		HtmlContent string `json:"html_content"`
	}

	decoder := json.NewDecoder(r.Body)
	var email_params params
	err := decoder.Decode(&email_params)
	if err != nil {
		httpapi.RespondWithError(w, http.StatusInternalServerError, "error decoding parameters", err)
		return
	}

	log.Printf("text: %s\n\n", email_params.TextContent)
	log.Printf("html: %s\n\n", email_params.HtmlContent)

}
