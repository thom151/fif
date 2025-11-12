package dropbox

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/sharing"
)

func UploadToDropbox(filePath, folder, accToken string) (url string, err error) {
	config := dropbox.Config{
		Token:    accToken,
		LogLevel: dropbox.LogInfo,
	}

	client := files.New(config)
	sharingClient := sharing.New(config)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(filePath)

	//newName := getNewFileName(folder, "fif", ".mp4", accToken)
	log.Printf("folder(raw): %q", folder)
	fmt.Printf("fileName: %s\n", fileName)

	folder = strings.TrimSpace(folder)
	folder = strings.ReplaceAll(folder, "/", "-")
	dropboxPath := filepath.Join("/", folder, fileName)
	arg := files.NewUploadArg(dropboxPath)
	arg.Autorename = true

	res, err := client.Upload(arg, bytes.NewReader(data))
	if err != nil {
		fmt.Printf("error upload: %v", err)
		return "", err
	}

	finalPath := res.PathDisplay

	link, err := sharingClient.CreateSharedLinkWithSettings(
		sharing.NewCreateSharedLinkWithSettingsArg(finalPath),
	)
	if err != nil {
		return "", err
	}

	if fileLink, ok := link.(*sharing.FileLinkMetadata); ok {
		log.Println("Uploaded file to dropbox ", dropboxPath)
		return fileLink.Url, nil
	}

	return "", fmt.Errorf("Unexpected link type: %T", link)

}
