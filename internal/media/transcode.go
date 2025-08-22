// / internal/media/transcode.go
package media

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// TranscodeH264 takes a video file path and writes a H.264/AAC MP4 next to it.
// It returns the output path (caller can defer os.Remove on it) or an error.
func TranscodeH264(inPath string) (string, error) {
	base := strings.TrimSuffix(inPath, filepath.Ext(inPath))
	outPath := fmt.Sprintf("%s-transcoded.mp4", base)

	var stderr bytes.Buffer
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inPath,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-f", "mp4",
		outPath,
	)
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %w, stderr: %s", err, stderr.String())
	}
	return outPath, nil
}
