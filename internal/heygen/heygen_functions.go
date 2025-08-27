// #nosec G204 G304

package heygen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func GenerateVideoHeygen(ctx context.Context, script, key, avatarId, voiceID, fifTitle string) (videoID string, err error) {
	url := "https://api.heygen.com/v2/video/generate"
	payload := VideoRequest{
		Caption: false,
		Title:   fifTitle,
		VideoInputs: []VideoInput{
			{
				Character: &CharacterSettings{
					Type:        "avatar",
					AvatarID:    avatarId,
					Scale:       1.0,
					AvatarStyle: "normal",
				},
				Voice: VoiceSettings{
					Type:      "text",
					VoiceID:   voiceID,
					InputText: script,
					Emotion:   "Excited",
					Speed:     1.0,
					ElevenLabs: &ElevenLabsSettings{
						Stability:  0.35,
						Similarity: 0.90,
						Style:      0.70,
					},
				},
			},
		},
		Dimension: Dimension{
			Width:  1280,
			Height: 720,
		},
	}

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

func DownloadHeygenVideo(ctx context.Context, shareURL, safeOutputPath string) error {

	if err := os.MkdirAll(filepath.Dir(safeOutputPath), 0o750); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, shareURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("download non-2xx: %s body=%s", resp.Status, strings.TrimSpace(string(b)))
	}
	//nolint:gosec // G304: safePath-validated
	out, err := os.OpenFile(safeOutputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GetVideoStatus(ctx context.Context, key, videoID string) (avatarUrl string, err error) {
	url := fmt.Sprintf("https://api.heygen.com/v1/video_status.get?video_id=%s", videoID)
	client := &http.Client{Timeout: 30 * time.Second}

	for {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "", fmt.Errorf("creating status request: %w", err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Api-Key", key)

		res, err := client.Do(req)
		if err != nil {
			//SELECT CHANGE
			return "", err
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return "", err
		}

		if res.StatusCode != http.StatusOK {
			return "", err
		}

		var statusResp struct {
			Data struct {
				Status   string `json:"status"`
				VideoURL string `json:"video_url"`
			} `json:"data"`
		}
		err = json.Unmarshal(body, &statusResp)
		if err != nil {
			return "", fmt.Errorf("unmarshaling status response: %w", err)
		}

		switch statusResp.Data.Status {
		case "completed":
			return statusResp.Data.VideoURL, nil
		case "pending", "processing", "waiting":
			fmt.Printf("Video %s status: %s, checking again in 5 seconds...\n", videoID, statusResp.Data.Status)
			time.Sleep(5 * time.Second)
		case "failed":
			return "", fmt.Errorf("video generation failed")
		default:
			return "", fmt.Errorf("unknown status: %s", statusResp.Data.Status)
		}

		log.Println("Status: ", statusResp.Data.Status)
	}
}

type VideoRequest struct {
	Caption     bool         `json:"caption,omitempty"`
	Title       string       `json:"title,omitempty"`
	VideoInputs []VideoInput `json:"video_inputs"`
	Dimension   Dimension    `json:"dimension"`
}

type VideoInput struct {
	Character  *CharacterSettings  `json:"character,omitempty"`
	Voice      VoiceSettings       `json:"voice"`
	Background *BackgroundSettings `json:"background,omitempty"`
}

type CharacterSettings struct {
	Type        string  `json:"type"`
	AvatarID    string  `json:"avatar_id"`
	Scale       float64 `json:"scale,omitempty"`
	AvatarStyle string  `json:"avatar_style,omitempty"`
}

type VoiceSettings struct {
	Type       string              `json:"type"`
	VoiceID    string              `json:"voice_id,omitempty"`
	InputText  string              `json:"input_text,omitempty"`
	Emotion    string              `json:"emotion"`
	Speed      float64             `json"speed"`
	ElevenLabs *ElevenLabsSettings `json:"elevenlabs_settings"`
}

type ElevenLabsSettings struct {
	Similarity float64 `json:"similarity_boost"`
	Stability  float64 `json:"stability"`
	Style      float64 `json:"style"`
}

type BackgroundSettings struct {
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

type Dimension struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type VideoResponseHeyGen struct {
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Data struct {
		VideoID string `json:"video_id"`
	} `json:"data"`
}

type ShareRequest struct {
	VideoID string `json:"video_id"`
}

type ShareResponse struct {
	Code    int    `json:"code"`
	Data    string `json:"data"`
	Msg     any    `json:"msg"`
	Message any    `json:"message"`
}
