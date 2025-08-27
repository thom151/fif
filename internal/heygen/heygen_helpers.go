package heygen

import (
	"context"
)

func GenerateAndDownloadAvatar(ctx context.Context, key, script, avatarID, voiceID, fifTitle, avatarOutPath string) (file string, err error) {
	videoID, err := GenerateVideoHeygen(ctx, script, key, avatarID, voiceID, fifTitle)
	if err != nil {
		return "", err
	}

	avatarURL, err := GetVideoStatus(ctx, key, videoID)
	if err != nil {
		return "", err
	}

	err = DownloadHeygenVideo(ctx, avatarURL, avatarOutPath)
	if err != nil {
		return "", err
	}

	return avatarOutPath, nil

}
