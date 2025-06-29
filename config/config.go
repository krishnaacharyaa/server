package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port           string
	VerifyToken    string
	PhoneNumberID  string
	AccessToken    string
	GeminiAPIKey   string
	PropertiesFile string
}

func LoadConfig() *AppConfig {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file (this is fine if using system env vars): %v", err)
	}

	return &AppConfig{
		Port:           getEnv("PORT", "8080"),
		VerifyToken:    getEnv("WEBHOOK_VERIFY_TOKEN", ""),
		PhoneNumberID:  getEnv("PHONE_NUMBER_ID", ""),
		AccessToken:    getEnv("ACCESS_TOKEN", ""),
		GeminiAPIKey:   getEnv("GEMINI_API_KEY", ""),
		PropertiesFile: getEnv("PROPERTIES_FILE", "properties.json"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
