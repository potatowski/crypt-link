package config

import (
	_ "embed"
	"log"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile []byte

var (
	DatabaseUrl = ""
)

func Load() {
	DatabaseUrl = getEnv("DATABASE_URL", "mongodb://localhost:27017")
}

func getEnv(key, fallback string) string {
	env, err := godotenv.Unmarshal(string(envFile))
	if err != nil {
		log.Fatalf("Error when read file.env: %s", err)
	}

	if val := env[key]; val != "" {
		return val
	}

	return fallback
}
