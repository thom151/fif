package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	//	"github.com/sashabaranov/go-openai"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/dropbox"
	"github.com/thom151/fif/internal/httpapi"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type apiConfig struct {
	db                       *database.Queries
	jwtSecret                string
	s3Client                 *s3.Client
	s3Region                 string
	s3Bucket                 string
	s3CfDistribution         string
	dropboxAccToken          string
	dropboxAccTokenExpiresAt time.Time
	dropboxRefreshToken      string
	dropboxClientID          string
	dropboxClientSecret      string
	heygenApiKey             string
	deepgramApiKey           string
	openaiClient             openai.Client
	openaiAssistantID        string
	tempDir                  string
}

func main() {

	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT must be set")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL must be set")
	}

	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET must be set")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET environment variable is not set")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		log.Fatal("S3_REGION environment variable is not set")
	}

	s3CfDistribution := os.Getenv("S3_CF_DISTRO")
	if s3CfDistribution == "" {
		log.Fatal("S3_CF_DISTRO environment variable is not set")
	}
	dropboxRefreshToken := os.Getenv("DROPBOX_REFRESH_TOKEN")
	if dropboxRefreshToken == "" {
		log.Fatal("DROPBOX_REFRESH_TOKEN environment variable is not set")
	}

	dropboxClientID := os.Getenv("DROPBOX_CLIENT_ID")
	if dropboxClientID == "" {
		log.Fatal("DROPBOX_CLIENT_ID environment variable is not set")
	}

	dropboxClientSecret := os.Getenv("DROPBOX_CLIENT_SECRET")
	if dropboxClientSecret == "" {
		log.Fatal("DROPBOX_CLIENT_SECRET environment variable is not set")
	}

	dropboxAccToken, err := dropbox.GetNewAccessToken(dropboxRefreshToken, dropboxClientID, dropboxClientSecret)
	if err != nil {
		log.Fatal("DROPBOX_ACC_TOKEN cannot get")
	}

	fmt.Printf("ACC TOKEN: %s\n", dropboxAccToken.AccessToken)

	/*
		dropboxAccToken := os.Getenv("DROPBOX_ACC_TOKEN")
		if dropboxAccToken == "" {
			log.Fatal("DROPBOX_ACC_TOKEN environment variable is not set")
		}*/

	heygenApiKey := os.Getenv("HEYGEN_API_KEY")
	if heygenApiKey == "" {
		log.Fatal("HEYGEN_API_KEY not set")
	}

	openaiApiKey := os.Getenv("OPENAI_API_KEY")
	if openaiApiKey == "" {
		log.Fatal("OPENAI_API_KEY not set")
	}

	assistantID := os.Getenv("ASSISTANT_ID")
	if assistantID == "" {
		log.Fatal("ASSISTANT_ID not set")
	}

	deepgramApiKey := os.Getenv("DEEPGRAM_API_KEY")
	if assistantID == "" {
		log.Fatal("DEEPGRAM_API_KEY not set")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(s3Region))
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(awsCfg)

	db, err := sql.Open("libsql", dbURL)
	dbQueries := database.New(db)

	tempBase := filepath.Join(os.TempDir(), "fif")
	if err := os.MkdirAll(tempBase, 0o755); err != nil {
		log.Fatal(err)
	}
	apiCfg := apiConfig{
		db:                       dbQueries,
		jwtSecret:                secret,
		s3Client:                 client,
		s3Bucket:                 s3Bucket,
		s3Region:                 s3Region,
		s3CfDistribution:         s3CfDistribution,
		dropboxAccToken:          dropboxAccToken.AccessToken,
		dropboxAccTokenExpiresAt: time.Now().Add(time.Duration(dropboxAccToken.ExpiresIn) * time.Second),
		dropboxRefreshToken:      dropboxRefreshToken,
		dropboxClientID:          dropboxClientID,
		dropboxClientSecret:      dropboxClientSecret,
		heygenApiKey:             heygenApiKey,
		deepgramApiKey:           deepgramApiKey,
		openaiClient:             openai.NewClient(option.WithAPIKey(openaiApiKey)),
		openaiAssistantID:        assistantID,
		tempDir:                  tempBase,
	}

	if err != nil {
		log.Fatal("Cannot open db" + err.Error())
	}

	mux := http.NewServeMux()

	const filepathRoot = "./web"
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	mux.HandleFunc("GET /home", apiCfg.handlerHome)

	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	//API
	mux.HandleFunc("POST /api/upload_broll/{brollID}", apiCfg.handlerUploadBroll)
	mux.HandleFunc("POST /api/upload_music/{musicID}", apiCfg.handlerUploadMusic)
	mux.HandleFunc("POST /api/delete_broll/{brollID}", apiCfg.handlerDeleteBroll)
	mux.HandleFunc("POST /api/set_user_avatar_id", apiCfg.handlerSetUserAvatarAndVoiceID)
	mux.HandleFunc("POST /api/create_broll_meta", apiCfg.handlerCreateBrollMeta)
	mux.HandleFunc("POST /api/create_music_meta", apiCfg.handlerCreateMusicMeta)
	mux.HandleFunc("POST /api/fif_meta", apiCfg.handlerFifMeta)
	mux.HandleFunc("POST /api/create_fif_video/{fifID}", apiCfg.handlerCreateFifVideo)

	//GET API
	mux.HandleFunc("GET /api/brolls", apiCfg.handlerGetBrolls)

	// TEST ENDPOINTS
	mux.HandleFunc("POST /api/test_upload", apiCfg.handlerTestUploadV2)

	//FRONTEND
	mux.HandleFunc("GET /upload_broll", apiCfg.handlerGetUploadBrollPage)
	mux.HandleFunc("GET /upload_music", apiCfg.handlerGetUploadMusicPage)
	mux.HandleFunc("GET /login", apiCfg.handlerGetLoginPage)

	mux.HandleFunc("/healthz", handlerReadiness)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerHome(w http.ResponseWriter, r *http.Request) {
	httpapi.RenderTemplate(w, "home", nil)
}
