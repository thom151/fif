package fifYouTube

import (
	"context"
//	"fmt"
	"os"

	"google.golang.org/api/option"
	yt "google.golang.org/api/youtube/v3"
)

func UploadVideo(filename, title, description string) (string, error) {
	client := GetHTTPClient()

	service, err := yt.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	video := &yt.Video{
		Snippet: &yt.VideoSnippet{
			Title:       title,
			Description: description,
			CategoryId:  "22",
		},
		Status: &yt.VideoStatus{
			PrivacyStatus: "unlisted",
		},
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, video)
	response, err := call.Media(file).Do()
	if err != nil {
		return "", err
	}

	link := "https://www.youtube.com/watch?v=" + response.Id
	return link, nil
}
