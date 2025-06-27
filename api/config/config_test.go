package config

import (
	"os"
	"testing"
)

// Helper to set up a temporary .env file for testing
func setEnvFile(content string) {
	envFile = []byte(content)
}

func TestGetEnv_WithValue(t *testing.T) {
	setEnvFile("DATABASE_URL=mongodb://test:27017\n")
	val := getEnv("DATABASE_URL", "fallback")
	if val != "mongodb://test:27017" {
		t.Errorf("expected 'mongodb://test:27017', got '%s'", val)
	}
}

func TestGetEnv_WithFallback(t *testing.T) {
	setEnvFile("")
	val := getEnv("DATABASE_URL", "fallback")
	if val != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", val)
	}
}

func TestLoad_SetsDatabaseUrl(t *testing.T) {
	setEnvFile("DATABASE_URL=mongodb://prod:27017\n")
	Load()
	if DatabaseUrl != "mongodb://prod:27017" {
		t.Errorf("expected 'mongodb://prod:27017', got '%s'", DatabaseUrl)
	}
}

func TestLoad_UsesFallback(t *testing.T) {
	setEnvFile("")
	Load()
	if DatabaseUrl != "mongodb://localhost:27017" {
		t.Errorf("expected fallback 'mongodb://localhost:27017', got '%s'", DatabaseUrl)
	}
}

// Restore envFile after tests
func TestMain(m *testing.M) {
	orig := envFile
	code := m.Run()
	envFile = orig
	os.Exit(code)
}
