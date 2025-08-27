package editor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func CutAndSaveVideo(inputPath, outputPath string, startTime, duration float64, videoFormat VideoSeriesFormat, muteAudio ...bool) error {
	args := []string{
		"-ss", fmt.Sprintf("%.2f", startTime),
		"-i", inputPath,
		"-t", fmt.Sprintf("%.2f", duration),
		"-c:v", videoFormat.VideoCodec,
		"-preset", "fast",
		"-r", videoFormat.FrameRate,
		"-pix_fmt", videoFormat.PixelFormat,
	}

	if len(muteAudio) == 0 || !muteAudio[0] {
		args = append(args,
			"-c:a", videoFormat.AudioCodec,
			"-ar", videoFormat.SampleRate,
			"-channel_layout", videoFormat.ChannelLayout,
		)
	} else {
		args = append(args, "-an") // Mute audio
	}

	var stderr bytes.Buffer
	args = append(args, "-f", "mp4", "-movflags", "faststart", outputPath)

	//nolint:gosec // G204: safePath-validated
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = &stderr

	// Log the FFmpeg command for debugging
	log.Printf("Running FFmpeg command: %v", cmd.Args)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run FFmpeg: %v, stderr: %s", err, stderr.String())
	}
	if _, err := os.Stat(outputPath); err != nil {
		return fmt.Errorf("output file was not created: %v", err)
	}
	return nil

}
