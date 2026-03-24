package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	MongoURI              string
	MongoDatabase         string
	JWTSecret             string
	AllowedOrigin         string
	AccessTokenTTLMinutes int
	RefreshTokenTTLDays   int
	FrontendURL           string
	SMTPHost              string
	SMTPPort              string
	SMTPUsername          string
	SMTPPassword          string
	SMTPFromEmail         string
	GoogleClientID        string
	GoogleClientSecret    string
	GoogleRedirectURL     string
	GithubClientID        string
	GithubClientSecret    string
	GithubRedirectURL     string
	UploadStorage         string
	UploadLocalPath       string
	S3Endpoint            string
	S3Region              string
	S3Bucket              string
	S3AccessKeyID         string
	S3SecretAccessKey     string
	S3UsePathStyle        string
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		Port:                  get("PORT", "8080"),
		MongoURI:              get("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:         get("MONGO_DATABASE", "stratyx"),
		JWTSecret:             get("JWT_SECRET", "change-me"),
		AllowedOrigin:         get("ALLOWED_ORIGIN", "http://localhost:3000"),
		AccessTokenTTLMinutes: 30,
		RefreshTokenTTLDays:   14,
		FrontendURL:           get("FRONTEND_URL", "http://localhost:3000"),
		SMTPHost:              get("SMTP_HOST", ""),
		SMTPPort:              get("SMTP_PORT", "587"),
		SMTPUsername:          get("SMTP_USERNAME", ""),
		SMTPPassword:          get("SMTP_PASSWORD", ""),
		SMTPFromEmail:         get("SMTP_FROM_EMAIL", ""),
		GoogleClientID:        get("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    get("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:     get("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/oauth/google/callback"),
		GithubClientID:        get("GITHUB_CLIENT_ID", ""),
		GithubClientSecret:    get("GITHUB_CLIENT_SECRET", ""),
		GithubRedirectURL:     get("GITHUB_REDIRECT_URL", "http://localhost:8080/api/auth/oauth/github/callback"),
		UploadStorage:         get("UPLOAD_STORAGE", "local"),
		UploadLocalPath:       get("UPLOAD_LOCAL_PATH", "./uploads"),
		S3Endpoint:            get("S3_ENDPOINT", ""),
		S3Region:              get("S3_REGION", ""),
		S3Bucket:              get("S3_BUCKET", ""),
		S3AccessKeyID:         get("S3_ACCESS_KEY_ID", ""),
		S3SecretAccessKey:     get("S3_SECRET_ACCESS_KEY", ""),
		S3UsePathStyle:        get("S3_USE_PATH_STYLE", "true"),
	}
}

func get(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
