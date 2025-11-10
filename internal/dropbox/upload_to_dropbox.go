package dropbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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

	//fileName := filepath.Base(filePath)

	newName := getNewFileName(folder, "fif", ".mp4", accToken)
	dropboxPath := filepath.Join("/", folder, newName)
	arg := files.NewUploadArg(dropboxPath)

	_, err = client.Upload(arg, bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	log.Println("Uploaded file to dropbox: ", dropboxPath)
	return dropboxPath, nil

}

func listFolder(path, token string) ([]string, error) {
	url := "https://api.dropboxapi.com/2/files/list_folder"
	data := fmt.Sprintf(`{"path":"%s"}`, path)

	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	entries, _ := result["entries"].([]any)
	names := []string{}
	for _, e := range entries {
		entry := e.(map[string]any)
		names = append(names, entry["name"].(string))
	}

	fmt.Sprintf("folders successfully listed")
	return names, nil
}

func getNewFileName(folder, base, ext, token string) string {
	files, err := listFolder(folder, token)
	if err != nil {
		return base + ext
	}
	count := 0
	for _ = range files {
		count++
	}
	if count == 0 {
		return base + ext
	}
	return fmt.Sprintf("%s_%d%s", base, count+1, ext)
}

func main() {

}
