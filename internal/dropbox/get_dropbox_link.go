package dropbox

import (
	"fmt"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

func GetDropboxLink(dropboxPath, accToken string) (url string, err error) {
	config := dropbox.Config{
		Token:    accToken,
		LogLevel: dropbox.LogInfo,
	}

	client := sharing.New(config)

	arg := sharing.NewCreateSharedLinkWithSettingsArg(dropboxPath)
	link, err := client.CreateSharedLinkWithSettings(arg)
	if err != nil {
		return "", err
	}

	if fileLink, ok := link.(*sharing.FileLinkMetadata); ok {
		return fileLink.Url, nil
	}

	return "", fmt.Errorf("unexpected link type: %T", link)
}
