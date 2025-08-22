package assets

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

func GetAssestPath(mediaType string) string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Println("error reading random bytes")
		return ""
	}

	randString := base64.RawURLEncoding.EncodeToString(randomBytes)

	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ""
	}
	ext := "." + parts[1]
	return fmt.Sprintf("%s%s", randString, ext)

}
