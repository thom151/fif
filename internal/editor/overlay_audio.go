package editor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func OverlayAudio(inputPath, musicPath, outputPath string, vf VideoSeriesFormat) (string, error) {
	//nolint:gosec // G204: safePath-validated

	filter := fmt.Sprintf(
		"[0:a]volume=%.2f[a0];"+
			"[1:a]volume=%.2f[a1];"+
			"[a0][a1]amix=inputs=2:duration=longest:dropout_transition=0[aout]",
		1.8, // voice gain
		0.7, // music volume
	)

	cmd := exec.Command(
		"ffmpeg",
		"-y",
		"-i", inputPath, // 0: video (voice)
		"-i", musicPath, // 1: music
		"-filter_complex", filter,
		"-map", "0:v:0",
		"-map", "[aout]",
		"-c:v", "copy",
		"-c:a", vf.AudioCodec, // e.g., aac
		"-ar", vf.SampleRate, // e.g., 44100
		"-ac", "2",
		"-movflags", "+faststart",
		"-shortest",
		outputPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to overlay audio: %v, stderr: %s", err, stderr.String())
	}
	return outputPath, nil
}

func CutAndSaveAudio(audioPath, outputPath string, duration float64, videoFormat VideoSeriesFormat) error {
	totalDuration, err := GetTotalDuration(audioPath)
	if err != nil {
		return err
	}

	if totalDuration < duration {
		return fmt.Errorf("requested duration exceeds music duration")
	}

	fadeDuration := 2.0
	fadeStart := duration - fadeDuration
	args := []string{
		"-y",
		"-ss", "4",
		"-i", audioPath,
		"-t", fmt.Sprintf("%.2f", duration),
		"-map", "0:a:0",
		"-af", fmt.Sprintf("volume=0.3,afade=t=out:st=%.2f:d=%.2f", fadeStart, fadeDuration),
		"-vn",
		"-c:a", "libmp3lame", // ✅ use MP3 encoder
		"-b:a", "192k",
		"-ar", videoFormat.SampleRate,
		"-ac", "2",
		outputPath,
	}
	//nolint:gosec // G204: safePath-validated
	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Printf("Running cut audio  command: %v", cmd.Args)

	if err := cmd.Run(); err != nil {
		log.Printf("ffmpeg error: %v, stderr: %s", err, stderr.String())
		return fmt.Errorf("failed to cut audio: %v", err)
	}

	if _, err := os.Stat(outputPath); err != nil {
		fmt.Errorf("output audio file was not created: %v", err)
		return err
	}

	return nil
}
