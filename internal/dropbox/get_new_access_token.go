package dropbox

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type dropboxTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func GetNewAccessToken(refreshToken, appKey, appSecret string) (dropboxTokenResp, error) {
	endpoint := "https://api.dropbox.com/oauth2/token"
	data := url.Values{}

	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, _ := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	req.SetBasicAuth(appKey, appSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return dropboxTokenResp{}, err
	}

	defer res.Body.Close()

	var dropboxToken dropboxTokenResp
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&dropboxToken)
	if err != nil {
		return dropboxTokenResp{}, err
	}

	return dropboxToken, nil
}
