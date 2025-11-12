package editor

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type MediaParams struct {
	Width     int
	Height    int
	FPS       string
	AudioRate string
}

func ProbeMedia(ctx context.Context, path string) (MediaParams, error) {
	var m MediaParams

	// Get video info: width, height, fps
	out, err := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height,r_frame_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		return m, fmt.Errorf("probe video: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) >= 3 {
		m.Width, _ = strconv.Atoi(lines[0])
		m.Height, _ = strconv.Atoi(lines[1])
		m.FPS = lines[2]
	} else {
		return m, fmt.Errorf("unexpected ffprobe video output")
	}

	// Get audio sample rate
	out, err = exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=sample_rate",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	).Output()
	if err != nil {
		m.AudioRate = "44100" // fallback if no audio
	} else {
		m.AudioRate = strings.TrimSpace(string(out))
		if m.AudioRate == "" {
			m.AudioRate = "44100"
		}
	}

	return m, nil
}
