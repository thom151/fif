package heygen

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

func GenerateAndDownloadAvatar(ctx context.Context, key, script, avatarID, voiceID, fifTitle, avatarOutPath string) (file string, err error) {
	videoID, err := GenerateVideoHeygen(ctx, script, key, avatarID, voiceID, fifTitle)
	if err != nil {
		return "", err
	}

	avatarURL, err := GetVideoStatus(ctx, key, videoID)
	if err != nil {
		return "", err
	}

	err = DownloadHeygenVideo(ctx, avatarURL, avatarOutPath)
	if err != nil {
		return "", err
	}

	return avatarOutPath, nil

}

func isTalkingPhoto(ctx context.Context, key, avatarID string) (bool, error) {
	url := "https://api.heygen.com/v2/photo_avatar/" + avatarID

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-key", key)

	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		return true, nil
	}

	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		return false, nil
	}

	body, _ := io.ReadAll(res.Body)
	return false, fmt.Errorf("photo avatar details check failed: %s (%s)", res.Status, string(body))

}
