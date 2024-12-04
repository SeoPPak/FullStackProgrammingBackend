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

func Init() {
	MongoConfig := InitDB()

	AppConfig = &Config{
		MongoDB: MongoConfig,
	}

}
