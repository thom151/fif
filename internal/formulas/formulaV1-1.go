package formulas

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/thom151/fif/internal/editor"
)

func FormulaV1_1(ctx context.Context, dgKey, base, avatarPath, brollPath, fifPath, musicPath string, cutIndex int) (fifFinalPath string, err error) {

	avatarAudio, err := extractAudio(avatarPath)
	if err != nil {
		return "", fmt.Errorf("error extracting audio: %v", err)
	}
	defer os.Remove(avatarAudio)

	timestamp, err := getCutTimestamp(dgKey, avatarAudio, cutIndex)
	if err != nil {
		return "", fmt.Errorf("failed to get timestamp: %v", err)
	}

	cutAvatar := filepath.Join(base, "cutAvatar.mp4")
	err = editor.CutAndSaveVideo(avatarPath, cutAvatar, 0, timestamp-0.15, defaultVF)
	if err != nil {
		return "", fmt.Errorf("error cutting avatar: %v", err)
	}
	defer os.Remove(cutAvatar)

	log.Printf("avatar successfully cut")

	avatarMatchBroll := filepath.Join(base, "avatarMatchBroll.mp4")
	err = editor.MakeAvatarMatchBroll(ctx, cutAvatar, brollPath, avatarMatchBroll)

	log.Printf("avatar + broll successfully normalized")

	concatList := filepath.Join(base, "concat.txt")
	list := fmt.Sprintf("file '%s'\nfile '%s'\n",
		filepath.Base(avatarMatchBroll),
		filepath.Base(brollPath),
	)
	if err := os.WriteFile(concatList, []byte(list), 0o600); err != nil {
		return "", fmt.Errorf("write concat list: %w", err)
	}
	defer os.Remove(concatList)
	defer os.Remove(avatarMatchBroll)
	defer os.Remove(brollPath)

	emptyConcatenated := filepath.Join(base, "concatenated.mp4")
	concatenated, err := editor.ConcatVideosFromTextFile(concatList, base, emptyConcatenated, defaultVF)
	if err != nil {
		return "", fmt.Errorf("failed to concat videos: %v", err)
	}
	defer os.Remove(concatenated)

	fifDuration, err := editor.GetTotalDuration(concatenated)
	if err != nil {
		return "", fmt.Errorf("error getting avatar duration : %v", err)
	}

	cutAudio := filepath.Join(base, "cut_audio.mp3")
	err = editor.CutAndSaveAudio(musicPath, cutAudio, fifDuration, defaultVF)
	if err != nil {
		return "", fmt.Errorf("failed to cut audio: %v", err)
	}
	defer os.Remove(cutAudio)

	outPath := fifPath
	if outPath == "" {
		outPath = filepath.Join(base, "final.mp4")
	}

	out, err := editor.OverlayAudio(concatenated, cutAudio, outPath, defaultVF)
	if err != nil {
		return "", fmt.Errorf("failed to overlay music: %v", err)
	}

	log.Printf("overlay audio successful")

	return out, nil

}
