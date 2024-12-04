package config

type Config struct {
	MongoDB MongoConfig
	OCR     OCRConfig
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
	MongoConfig := InitDB()
	ocr := InitOCR()

	AppConfig = &Config{
		MongoDB: MongoConfig,
		OCR:     ocr,
	}

}
