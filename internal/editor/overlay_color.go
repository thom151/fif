package editor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type OverlayFadeConfig struct {
	Color    string  // Hex without '#', e.g., "70bf44"
	Opacity  float64 // If >1 it's treated as 0–255 alpha; if <=1 it's treated as 0–1
	Duration float64 // Seconds the fade-out lasts (starts at t=0)
	CRF      int     // e.g., 20 (default)
	Preset   string  // e.g., "veryfast" (default)
}

func AddColorFadeOverlay(ctx context.Context, inputPath, outputPath string, cfg OverlayFadeConfig) (string, error) {
	// Normalize color
	col := strings.TrimSpace(cfg.Color)
	if col == "" {
		col = "70bf44"
	}
	if col[0] != '#' {
		col = "#" + col
	}

	// Normalize opacity: accept 0–1 or 0–255 (e.g., "16" -> 16/255)
	baseOpacity := cfg.Opacity
	if baseOpacity > 1.0 {
		baseOpacity = baseOpacity / 255.0
	}
	if baseOpacity < 0 {
		baseOpacity = 0
	}
	if baseOpacity > 1 {
		baseOpacity = 1
	}
	if baseOpacity == 0 {
		// Ensure we actually see a tint before fading
		baseOpacity = 16.0 / 255.0
	}

	// Duration defaults to something reasonable if unset
	dur := cfg.Duration
	if dur <= 0 {
		dur = 1.25
	}

	crf := cfg.CRF
	if crf <= 0 {
		crf = 20
	}
	preset := cfg.Preset
	if preset == "" {
		preset = "veryfast"
	}

	// Filter graph:
	// 1) Ensure base video is rgba
	// 2) Create tiny solid color with alpha, scale to match base via scale2ref
	// 3) Fade the overlay's alpha out from t=0 over 'dur' seconds
	// 4) Overlay on top; map video+optional audio
	filter := fmt.Sprintf(
		`[0:v]format=rgba[base];`+
			`color=c=%s@%s:s=16x16:d=1[solid];`+
			`[solid][base]scale2ref=w=iw:h=ih[ovl][base_sized];`+
			`[ovl]format=rgba,fade=t=out:st=0:d=%0.3f:alpha=1[ovl_faded];`+
			`[base_sized][ovl_faded]overlay=shortest=1[vout]`,
		col, trimFloat(baseOpacity), dur,
	)

	args := []string{
		"-y",
		"-i", inputPath,
		"-filter_complex", filter,
		"-map", "[vout]",
		"-map", "0:a?",
		"-c:v", "libx264",
		"-crf", strconv.Itoa(crf),
		"-preset", preset,
		"-c:a", "copy",
		"-movflags", "+faststart",
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to add color fade overlay: %v\nffmpeg: %s", err, stderr.String())
	}

	return outputPath, nil
}

func trimFloat(f float64) string {
	// Keep a short string for ffmpeg args
	s := fmt.Sprintf("%.6f", f)
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}
