package formulas

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/thom151/fif/internal/editor"
)

var defaultVF = editor.VideoSeriesFormat{
	VideoCodec:    "libx264",
	AudioCodec:    "aac",
	FrameRate:     "30",
	PixelFormat:   "yuv420p",
	SampleRate:    "44100",
	ChannelLayout: "stereo",
}

var defaultOverlayConfig = editor.OverlayFadeConfig{
	Color:    "#00ff00@1.0",
	Start:    0.0,
	Duration: 1.5,
	CRF:      20,
	Preset:   "veryfast",
}

func FormulaV1(ctx context.Context, base, avatarPath, brollPath, fifPath string) (fifFinalPath string, err error) {

	avatarDuration, err := editor.GetTotalDuration(avatarPath)
	if err != nil {
		return "", fmt.Errorf("error getting avatar duration : %v", err)
	}
	cutAvatar := filepath.Join(base, "cut.mp4")
	err = editor.CutAndSaveVideo(avatarPath, cutAvatar, 0, avatarDuration-0.7, defaultVF)
	if err != nil {
		return "", fmt.Errorf("error cutting avatar: %v", err)
	}
	defer os.Remove(cutAvatar)
	log.Printf("avatar successfully cut")

	avatarNormalized := filepath.Join(base, "avatar_norm.mp4")
	brollNormalized := filepath.Join(base, "broll_norm.mp4")

	if err := editor.NormalizeVideoV2(ctx, avatarPath, avatarNormalized, defaultVF); err != nil {
		return "", fmt.Errorf("normalize avatar: %w", err)
	}
	if err := editor.NormalizeVideoV2(ctx, brollPath, brollNormalized, defaultVF); err != nil {
		return "", fmt.Errorf("normalize broll: %w", err)
	}

	log.Printf("avatar + broll successfully normalized")

	concatList := filepath.Join(base, "concat.txt")
	list := fmt.Sprintf("file '%s'\nfile '%s'\n",
		filepath.Base(avatarNormalized),
		filepath.Base(brollNormalized),
	)
	if err := os.WriteFile(concatList, []byte(list), 0o600); err != nil {
		return "", fmt.Errorf("write concat list: %w", err)
	}
	defer os.Remove(concatList)
	defer os.Remove(avatarNormalized)
	defer os.Remove(brollNormalized)

	emptyConcatenated := filepath.Join(base, "concatenated.mp4")
	concatenated, err := editor.ConcatVideosFromTextFile(concatList, base, emptyConcatenated, defaultVF)
	if err != nil {
		return "", fmt.Errorf("failed to concat videos: %v", err)
	}
	defer os.Remove(concatenated)

	outPath := fifPath
	if outPath == "" {
		outPath = filepath.Join(base, "final.mp4")
	}

	out, err := editor.AddColorFadeOverlay(ctx, concatenated, outPath, defaultOverlayConfig)
	if err != nil {
		return "", fmt.Errorf("failed to put overlay: %v", err)
	}

	return out, nil

}
