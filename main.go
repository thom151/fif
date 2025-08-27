package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"github.com/thom151/fif/internal/database"
	"github.com/thom151/fif/internal/httpapi"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type apiConfig struct {
	db                *database.Queries
	jwtSecret         string
	s3Client          *s3.Client
	s3Region          string
	s3Bucket          string
	heygenApiKey      string
	openaiClient      *openai.Client
	openaiAssistantID string
	tempDir           string
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
		db:                dbQueries,
		jwtSecret:         secret,
		s3Client:          client,
		s3Bucket:          s3Bucket,
		s3Region:          s3Region,
		heygenApiKey:      heygenApiKey,
		openaiClient:      openai.NewClient(openaiApiKey),
		openaiAssistantID: assistantID,
		tempDir:           tempBase,
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
	mux.HandleFunc("POST /api/delete_broll/{brollID}", apiCfg.handlerDeleteBroll)
	mux.HandleFunc("POST /api/set_user_avatar_id", apiCfg.handlerSetUserAvatarAndVoiceID)
	mux.HandleFunc("POST /api/create_broll_meta", apiCfg.handlerCreateBrollMeta)
	mux.HandleFunc("POST /api/fif_meta", apiCfg.handlerFifMeta)
	mux.HandleFunc("POST /api/create_fif_video/{fifID}", apiCfg.handlerCreateFifVideo)

	//GET API
	mux.HandleFunc("GET /api/brolls", apiCfg.handlerGetBrolls)

	//FRONTEND
	mux.HandleFunc("GET /upload_broll", apiCfg.handlerGetUploadBrollPage)
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
