package editor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func MakeAvatarMatchBroll(ctx context.Context, avatarIn, brollRef, outPath string) error {
	// 1) Probe b-roll for params to match
	br, err := ProbeMedia(ctx, brollRef)
	if err != nil {
		return fmt.Errorf("probe b-roll: %w", err)
	}
	if br.Width <= 0 || br.Height <= 0 {
		return fmt.Errorf("invalid probed size: %dx%d", br.Width, br.Height)
	}
	fps := sanitizeFPS(br.FPS) // accepts "25" or "30000/1001"
	ar := br.AudioRate         // e.g., "48000" or "44100"
	if ar == "" {
		ar = "44100"
	}

	// 2) Ensure output dir exists
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// 3) Build filters
	// - scale to fit inside WxH preserving AR
	// - pad to exact WxH (required for concat-copy when dimensions must match)
	// - set fps to match b-roll (keeps caption/lipsync timelines aligned)
	vf := fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,"+
			"pad=%d:%d:(ow-iw)/2:(oh-ih)/2,fps=%s",
		br.Width, br.Height, br.Width, br.Height, fps,
	)

	// Audio chain:
	// - loudnorm: consistent loudness
	// - aresample=async: align audio PTS to video PTS (prevents drift)
	af := "loudnorm=I=-16:TP=-1.5:LRA=11,aresample=async=1:first_pts=0"

	// 4) ffmpeg args (CPU-light: ultrafast + 1–2 threads)
	args := []string{
		"-y",
		"-i", avatarIn,

		// VIDEO: match size/fps & keep it easy on CPU
		"-vf", vf,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		"-threads", "2", // drop to "1" on tiny boxes

		// AUDIO: match b-roll sample rate & normalize
		"-c:a", "aac",
		"-ar", ar,
		"-ac", "2",
		"-af", af,

		// good MP4 hygiene
		"-movflags", "+faststart",
		outPath,
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg avatar→match b-roll failed: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

// sanitizeFPS allows values like "25" or "30000/1001"; falls back to "30" if empty.
func sanitizeFPS(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "0/0" {
		return "30"
	}
	return s
}
