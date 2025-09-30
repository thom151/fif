package dropbox

import (
	"bytes"
	"log"
	"os"
	"path/filepath"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

func UploadToDropbox(filePath, folder, accToken string) (url string, err error) {
	config := dropbox.Config{
		Token:    accToken,
		LogLevel: dropbox.LogInfo,
	}

	client := files.New(config)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	fileName := filepath.Base(filePath)

	dropboxPath := filepath.Join("/", folder, fileName)
	arg := files.NewUploadArg(dropboxPath)

	_, err = client.Upload(arg, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	log.Println("Uploaded file to dropbox: ", dropboxPath)
	return dropboxPath, nil

}
