package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GoogleClientID    string
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
	AppConfig *Config = &Config{}
)

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
	MongoConfig := InitDB()
	ocr := InitOCR()

	AppConfig = &Config{
		GoogleClientID:    "974943437893-ohotc0roqcakgi26o33ubm45ng0fp08e.apps.googleusercontent.com",
		GoogleLoginConfig: googleLoginConfig,
		MongoDB:           MongoConfig,
		OCR:               ocr,
	}

}

func getRedirectURL(userAgent string) string {
	if strings.Contains(strings.ToLower(userAgent), "android") {
		return "http://10.0.2.2:5000/auth/google/callback"
	}
	return "http://localhost:5000/auth/google/callback"
}

func GoogleConfig() oauth2.Config {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	// 기본 리디렉션 URL 설정
	redirectURL := getRedirectURL("")

	googleLoginConfig := oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return googleLoginConfig
}

// UpdateRedirectURL은 사용자 에이전트에 따라 리디렉션 URL을 업데이트합니다
func (c *Config) UpdateRedirectURL(userAgent string) {
	config := c.GoogleLoginConfig
	config.RedirectURL = getRedirectURL(userAgent)
	c.GoogleLoginConfig = config
}
