package config

import (
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	SessionStore      *sessions.CookieStore
	GoogleLoginConfig oauth2.Config
	MongoDB           MongoConfig
	OCR               OCRConfig
}

type MongoConfig struct {
	URI      string
	Database string
}

type OCRConfig struct {
	URL       string
	SecretKey string
}

var (
	AppConfig *Config
)

func GoogleConfig() oauth2.Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	googleLoginConfig := oauth2.Config{
		RedirectURL:  "http://localhost:5000/auth/google/callback",
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint: google.Endpoint,
	}

	return googleLoginConfig
}

func InitSession() *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte("super-secret-key"))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	return store
}

func InitDB() MongoConfig {
	mongoConfig := MongoConfig{
		URI:      "mongodb+srv://dbedit:AyochsheJ1@fullstackprogramming.mjgyo.mongodb.net/?retryWrites=true&w=majority&appName=FullStackProgramming",
		Database: "FullStackProgramming",
	}

	return mongoConfig
}

func InitOCR() OCRConfig {
	ocrConfig := OCRConfig{
		URL:       "https://3lw4f4mamp.apigw.ntruss.com/custom/v1/35733/81998f2d759c60f8772617c8d9589f4f1f3e83a4f7fca03370e66ebc35487f3a/document/receipt",
		SecretKey: "d2xXSGVlTElyaGV6VGVEcUFIeXh6d09DTWpOVUdaS0s=",
	}

	return ocrConfig
}

func Init() {
	googleLoginConfig := GoogleConfig()
	store := InitSession()
	MongoConfig := InitDB()
	ocr := InitOCR()

	AppConfig = &Config{
		SessionStore:      store,
		GoogleLoginConfig: googleLoginConfig,
		MongoDB:           MongoConfig,
		OCR:               ocr,
	}

}
