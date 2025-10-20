package formulas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/thom151/fif/internal/editor"
)

var defaultVF = editor.VideoSeriesFormat{
	VideoCodec:    "libx264",
	AudioCodec:    "aac",
	FrameRate:     "30",
	PixelFormat:   "yuv420p",
	SampleRate:    "44100",
	ChannelLayout: "stereo",
}

var defaultOverlayConfig = editor.OverlayFadeConfig{
	Color:    "#70bf44",
	Opacity:  0.50,
	Duration: 3,
	CRF:      18,
	Preset:   "veryfast",
}

func FormulaV1(ctx context.Context, dgKey, base, avatarPath, brollPath, fifPath, musicPath string, cutIndex int) (fifFinalPath string, err error) {

	avatarAudio, err := extractAudio(avatarPath)
	if err != nil {
		return "", fmt.Errorf("error extracting audio: %v", err)
	}
	defer os.Remove(avatarAudio)

	timestamp, err := getCutTimestamp(dgKey, avatarAudio, cutIndex)
	if err != nil {
		return "", fmt.Errorf("failed to get timestamp: %v", err)
	}

	cutAvatar := filepath.Join(base, "cut.mp4")
	err = editor.CutAndSaveVideo(avatarPath, cutAvatar, 0, timestamp-0.15, defaultVF)
	if err != nil {
		return "", fmt.Errorf("error cutting avatar: %v", err)
	}
	defer os.Remove(cutAvatar)
	log.Printf("avatar successfully cut")

	avatarNormalized := filepath.Join(base, "avatar_norm.mp4")
	brollNormalized := filepath.Join(base, "broll_norm.mp4")

	if err := editor.NormalizeVideoV2(ctx, cutAvatar, avatarNormalized, defaultVF); err != nil {
		return "", fmt.Errorf("normalize avatar: %w", err)
	}
	if err := editor.NormalizeVideoV2(ctx, brollPath, brollNormalized, defaultVF); err != nil {
		return "", fmt.Errorf("normalize broll: %w", err)
	}

	log.Printf("avatar + broll successfully normalized")

	/*
		emptyColor := filepath.Join(base, "color.mp4")
		colorPath, err := editor.AddColorFadeOverlay(ctx, avatarNormalized, emptyColor, defaultOverlayConfig)
		if err != nil {
			return "", fmt.Errorf("failed to put overlay: %v", err)
		}
		defer os.Remove(colorPath)

		log.Printf("color overlay successful")
	*/

	concatList := filepath.Join(base, "concat.txt")
	list := fmt.Sprintf("file '%s'\nfile '%s'\n",
		filepath.Base(avatarNormalized),
		filepath.Base(brollNormalized),
	)
	if err := os.WriteFile(concatList, []byte(list), 0o600); err != nil {
		return "", fmt.Errorf("write concat list: %w", err)
	}
	defer os.Remove(concatList)
	defer os.Remove(avatarNormalized)
	defer os.Remove(brollNormalized)

	emptyConcatenated := filepath.Join(base, "concatenated.mp4")
	concatenated, err := editor.ConcatVideosFromTextFile(concatList, base, emptyConcatenated, defaultVF)
	if err != nil {
		return "", fmt.Errorf("failed to concat videos: %v", err)
	}
	defer os.Remove(concatenated)

	fifDuration, err := editor.GetTotalDuration(concatenated)
	if err != nil {
		return "", fmt.Errorf("error getting avatar duration : %v", err)
	}

	cutAudio := filepath.Join(base, "cut_audio.mp3")
	err = editor.CutAndSaveAudio(musicPath, cutAudio, fifDuration, defaultVF)
	if err != nil {
		return "", fmt.Errorf("failed to cut audio: %v", err)
	}
	defer os.Remove(cutAudio)

	outPath := fifPath
	if outPath == "" {
		outPath = filepath.Join(base, "final.mp4")
	}

	out, err := editor.OverlayAudio(concatenated, cutAudio, outPath, defaultVF)
	if err != nil {
		return "", fmt.Errorf("failed to overlay music: %v", err)
	}

	log.Printf("overlay audio successful")

	return out, nil

}

func getCutTimestamp(key, audioPath string, index int) (float64, error) {
	url := "https://api.deepgram.com/v1/listen?smart_format=true"
	//nolint:gosec // G304: safePath-validated
	file, err := os.Open(audioPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Token "+key)
	req.Header.Set("Content-Type", "audio/mpeg")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var dgSmartResp deepgramSmartResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&dgSmartResp)
	if err != nil {
		return 0, err
	}

	if len(dgSmartResp.Results.Channels) == 0 || len(dgSmartResp.Results.Channels[0].Alternatives) == 0 {
		return 0, fmt.Errorf("no transcript results from deepgram")
	}

	with := dgSmartResp.Results.Channels[0].Alternatives[0].Words[index].Word

	if strings.ToLower(with) != "thank" {

		if strings.ToLower(dgSmartResp.Results.Channels[0].Alternatives[0].Words[index+1].Word) == "thank" {
			log.Printf("returning timestamp early. Word: %s\n", dgSmartResp.Results.Channels[0].Alternatives[0].Words[index].Word)
			return dgSmartResp.Results.Channels[0].Alternatives[0].Words[index+1].Start, nil
		}
		log.Printf("iterating through all words")
		for _, word := range dgSmartResp.Results.Channels[0].Alternatives[0].Words {
			log.Printf("Word: %s\n", word.Word)
			if strings.ToLower(word.Word) == "thank" {
				log.Printf("iterated. Word: %s\n", word.Word)
				return word.Start, nil
			}
		}
	}

	log.Printf("thank got straight away. Word: %s\n", dgSmartResp.Results.Channels[0].Alternatives[0].Words[index].Word)
	indexTime := dgSmartResp.Results.Channels[0].Alternatives[0].Words[index].Start

	return indexTime, nil
}

func extractAudio(videoPath string) (string, error) {
	dir := filepath.Dir(videoPath)
	base := filepath.Base(videoPath)
	audioFileName := strings.Replace(base, "video-", "audio-", 1)
	audioFileName = strings.Replace(audioFileName, ".mp4", ".mp3", 1)
	audioPath := filepath.Join(dir, audioFileName)

	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	//nolint:gosec // G204: videoPath and audioPath are safePath-validated
	args := []string{
		"-i", videoPath,
		"-vn", // no video
		"-af", "volume=1.5",
		"-acodec", "mp3",
		audioPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	log.Printf("Running ffmpeg extract audio command: %v", cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to extract audio: %v", err)
	}

	if _, err := os.Stat(audioPath); err != nil {
		return "", fmt.Errorf("audio file was not created: %v", err)
	}

	return audioPath, nil
}

type deepgramSmartResponse struct {
	Metadata struct {
		TransactionKey string    `json:"transaction_key"`
		RequestID      string    `json:"request_id"`
		Sha256         string    `json:"sha256"`
		Created        time.Time `json:"created"`
		Duration       float64   `json:"duration"`
		Channels       int       `json:"channels"`
		Models         []string  `json:"models"`
		ModelInfo      struct {
			NAMING_FAILED struct {
				Name    string `json:"name"`
				Version string `json:"version"`
				Arch    string `json:"arch"`
			} `json:""`
		} `json:"model_info"`
	} `json:"metadata"`
	Results struct {
		Channels []struct {
			Alternatives []struct {
				Transcript string  `json:"transcript"`
				Confidence float64 `json:"confidence"`
				Words      []struct {
					Word           string  `json:"word"`
					Start          float64 `json:"start"`
					End            float64 `json:"end"`
					Confidence     float64 `json:"confidence"`
					PunctuatedWord string  `json:"punctuated_word"`
				} `json:"words"`
				Paragraphs struct {
					Transcript string `json:"transcript"`
					Paragraphs []struct {
						Sentences []struct {
							Text  string  `json:"text"`
							Start float64 `json:"start"`
							End   float64 `json:"end"`
						} `json:"sentences"`
						NumWords int     `json:"num_words"`
						Start    float64 `json:"start"`
						End      float64 `json:"end"`
					} `json:"paragraphs"`
				} `json:"paragraphs"`
			} `json:"alternatives"`
		} `json:"channels"`
	} `json:"results"`
}
