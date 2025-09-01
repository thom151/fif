package editor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type OverlayFadeConfig struct {
	Color    string
	Start    float64
	Duration float64
	CRF      int    // e.g., 20
	Preset   string // e.g., "veryfast"
}

// AddColorFadeOverlay places a colored layer over the video and fades its alpha to 0.
func AddColorFadeOverlay(ctx context.Context, inputPath, outputPath string, cfg OverlayFadeConfig) (string, error) {
	if cfg.Color == "" {
		cfg.Color = "#70BF44@0.16" // bright green
	}
	if cfg.Duration <= 0 {
		cfg.Duration = 2.0
	}
	// Overlay generator must exist long enough to cover the fade interval.
	colorDur := cfg.Start + cfg.Duration + 0.1

	// Build filtergraph for full-frame vs boxed overlay.
	// Full-frame overlay that auto-matches input size.
	filter := fmt.Sprintf(
		`[0:v]format=rgba[base];`+
			`color=c='%s':s=16x16:d=%g[solid];`+
			`[solid][base]scale2ref=w=iw:h=ih[fg][base_sized];`+
			`[fg]format=rgba,fade=t=out:st=%g:d=%g:alpha=1[fgfaded];`+
			`[base_sized][fgfaded]overlay=0:0:eof_action=pass`,
		"0x70BF44@0.16", // use 0x… and wrap in single quotes
		colorDur,
		cfg.Start, cfg.Duration,
	)
	crf := 20
	if cfg.CRF > 0 {
		crf = cfg.CRF
	}
	preset := "veryfast"
	if cfg.Preset != "" {
		preset = cfg.Preset
	}

	args := []string{
		"-y", "-i", inputPath,
		"-filter_complex", filter,
		// Re-encode video due to filtering; keep audio stream as-is.
		"-c:v", "libx264", "-crf", fmt.Sprint(crf), "-preset", preset,
		"-pix_fmt", "yuv420p",
		"-c:a", "copy",
		"-movflags", "+faststart",
		outputPath,
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v\n%s", err, stderr.String())
	}
	return outputPath, nil
}
