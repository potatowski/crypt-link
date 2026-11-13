package config

import (
	"errors"
	"io/fs"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	DatabaseUrl  = ""
	DatabaseName = ""
	Port         = ""
)

// envFiles are optional dotenv files loaded at startup, if present.
// Variables already set in the environment take precedence.
var envFiles = []string{".env", "config/.env"}

func Load() {
	for _, f := range envFiles {
		if err := godotenv.Load(f); err != nil && !errors.Is(err, fs.ErrNotExist) {
			log.Printf("config: failed to load %s: %v", f, err)
		}
	}

	DatabaseUrl = getEnv("DATABASE_URL", "mongodb://localhost:27017")
	DatabaseName = getEnv("DATABASE_NAME", "onetime")
	Port = getEnv("PORT", "8080")
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
