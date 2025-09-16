package editor

import (
	"context"
	"fmt"
	"os/exec"
)

type OverlayFadeConfig struct {
	Color    string  // Hex without '#', e.g., "70bf44"
	Opacity  float64 // If >1 it's treated as 0–255 alpha; if <=1 it's treated as 0–1
	Duration float64 // Seconds the fade-out lasts (starts at t=0)
	CRF      int     // e.g., 20 (default)
	Preset   string  // e.g., "veryfast" (default)
}

func AddColorFadeOverlay(ctx context.Context, inputPath, outputPath string, cfg OverlayFadeConfig) (string, error) {
	filter := fmt.Sprintf(
		`[0:v]format=rgba[base];`+
			`color=c=%s:s=16x16:d=9999,format=rgba[solid];`+
			`[solid][base]scale2ref=w=iw:h=ih[ovl][base_sized];`+
			`[ovl]colorchannelmixer=aa=%0.2f,fade=t=out:st=0:d=%0.2f:alpha=1[ovl_faded];`+
			`[base_sized][ovl_faded]overlay=shortest=1,format=yuv420p[v]`,
		cfg.Color, cfg.Opacity, cfg.Duration,
	)

	args := []string{
		"-y",
		"-i", inputPath,
		"-filter_complex", filter,
		"-map", "[v]",
		"-map", "0:a?",
		"-c:v", "libx264",
		"-crf", fmt.Sprintf("%d", cfg.CRF),
		"-preset", cfg.Preset,
		"-c:a", "copy",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	return outputPath, cmd.Run()
}
