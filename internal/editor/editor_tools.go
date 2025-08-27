package editor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type VideoSeriesFormat struct {
	VideoCodec    string
	AudioCodec    string
	FrameRate     string
	PixelFormat   string
	SampleRate    string
	ChannelLayout string
}

func ConcatVideosFromTextFile(textFile, base, outPath string, videoFormat VideoSeriesFormat) (string, error) {

	//nolint:gosec // G204: safePath-validated
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-f", "concat",
		"-safe", "0",
		"-i", textFile,
		"-c", "copy", // preserve both video+audio (no re-encode)
		"-movflags", "+faststart",
		outPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", err

	}
	return outPath, nil
}

func GetTotalDuration(video string) (float64, error) {
	var out, stderr bytes.Buffer
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", video)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("failed to get video duration: %v, stderr: %s", err, stderr.String())
	}
	duration, err := strconv.ParseFloat(strings.TrimSpace(out.String()), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %v", err)
	}
	return duration, nil
}
