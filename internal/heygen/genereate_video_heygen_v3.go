package heygen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func GenerateVideoHeygenV3(ctx context.Context, script, key, avatarId, voiceID, fifTitle string) (videoID string, err error) {

	url := "https://api.heygen.com/v3/videos"
	isPhoto, err := isTalkingPhoto(ctx, key, avatarId)
	if err != nil {
		log.Printf("isPhoto error: %v", err)
		return "", err
	}

	avatarSetting := CharacterSettings{
		Scale:       1.0,
		AvatarStyle: "normal",
	}

	if isPhoto {
		fmt.Println("a talking photo")
		avatarSetting.Type = "talking_photo"
		avatarSetting.TalkingPhotoID = avatarId
		avatarSetting.TalkingStyle = "stable"
		avatarSetting.UseAvatarIVModel = true
	} else {
		avatarSetting.Type = "avatar"
		avatarSetting.AvatarID = avatarId
	}

	payload := HeygenV3Request{
		Type:         "avatar",
		AvatarID:     avatarId,
		Title:        fifTitle,
		AspectRatio:  "16:9",
		Resolution:   "1080p",
		Script:       script,
		VoiceID:      voiceID,
		OutputFormat: "mp4",
	}

	payload.VoiceSettings.Speed = 1.0

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	log.Println("Getting request from heygen")

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", key)

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	log.Println("response got")
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return string(body), fmt.Errorf("API returned non-200 status: %s", res.Status)
	}

	var videoRes VideoResponseHeyGen
	err = json.Unmarshal(body, &videoRes)
	if err != nil {
		return "", err
	}

	if videoRes.Error != nil {
		return "", fmt.Errorf("API error: %s - %s", videoRes.Error.Code, videoRes.Error.Message)
	}

	return videoRes.Data.VideoID, nil

}
