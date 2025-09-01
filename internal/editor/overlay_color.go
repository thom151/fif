package editor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
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
		cfg.Color = "#70BF44"
	}
	if cfg.Duration <= 0 {
		cfg.Duration = 2.0
	}
	if cfg.CRF <= 0 {
		cfg.CRF = 20
	}
	if cfg.Preset == "" {
		cfg.Preset = "veryfast"
	}

	// ~16% opacity (Premiere “Opacity: 16”)
	const baseOpacity = 0.16

	// Make the color source way longer than the input (24h).
	const veryLong = 24.0 * 60.0 * 60.0

	// Strip accidental @alpha in cfg.Color to avoid "#rrggbb@x@y"
	col := cfg.Color
	if i := strings.Index(col, "@"); i >= 0 {
		col = col[:i]
	}

	filter := fmt.Sprintf(
		`[0:v]format=rgba[base];`+
			`color=c=%s:s=16x16:d=%0.3f,format=rgba[solid];`+
			`[solid][base]scale2ref=w=iw:h=ih[ovl][base_sized];`+
			`[ovl]colorchannelmixer=aa=%0.3f,fade=t=out:st=%0.3f:d=%0.3f:alpha=1[ovl_faded];`+
			`[base_sized][ovl_faded]blend=all_mode=multiply:all_opacity=1[vout]`,
		col, veryLong, baseOpacity, cfg.Start, cfg.Duration,
	)

	args := []string{
		"-y",
		"-i", inputPath,
		"-filter_complex", filter,
		"-map", "[vout]",
		"-map", "0:a?",
		"-c:v", "libx264", "-crf", fmt.Sprint(cfg.CRF), "-preset", cfg.Preset,
		"-pix_fmt", "yuv420p",
		"-c:a", "copy",
		"-movflags", "+faststart",
		outputPath,
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg failed: %v\n%s", err, stderr.String())
	}
	return outputPath, nil
}
