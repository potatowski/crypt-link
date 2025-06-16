package config

import "os"

type Config struct {
	MongoURI string
}

func Load() Config {
	return Config{
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
