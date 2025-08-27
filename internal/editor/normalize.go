package editor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func NormalizeVideo(ctx context.Context, inPath, outPath string, vf VideoSeriesFormat) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Width/height are already identical across clips.
	// We just unify fps, pixel format, and audio params.
	args := []string{
		"-y",
		"-i", inPath,
		"-r", vf.FrameRate, // e.g., 30
		"-pix_fmt", vf.PixelFormat, // e.g., yuv420p
		"-c:v", vf.VideoCodec, // e.g., libx264
		"-c:a", vf.AudioCodec, // e.g., aac
		"-ar", vf.SampleRate, // e.g., 44100
		"-ac", "2", // stereo
		"-movflags", "+faststart",
		outPath,
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg normalize failed: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

func NormalizeVideoV2(ctx context.Context, inPath, outPath string, vf VideoSeriesFormat) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// Audio chain:
	// - loudnorm: targets podcast/YouTube-friendly loudness
	// - highpass: cut rumble; lowpass: tame top-end harshness (HeyGen often sounds a bit bright)
	audioFilter := "loudnorm=I=-16:TP=-1.5:LRA=11,highpass=f=80,lowpass=f=12000"

	args := []string{
		"-y",
		"-i", inPath,
		// video normalization
		"-r", vf.FrameRate, // e.g. 30
		"-pix_fmt", vf.PixelFormat, // e.g. yuv420p
		"-c:v", vf.VideoCodec, // e.g. libx264
		// audio normalization
		"-c:a", vf.AudioCodec, // e.g. aac
		"-ar", vf.SampleRate, // e.g. 48000 or 44100
		"-ac", "2", // stereo
		"-af", audioFilter, // <- key addition
		// keep MP4 streamable
		"-movflags", "+faststart",
		outPath,
	}

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg normalize failed: %v, stderr: %s", err, stderr.String())
	}
	return nil
}
