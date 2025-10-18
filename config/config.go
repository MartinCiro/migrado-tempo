package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              int
	Debug             bool
	OAuthClientSecret string
	OAuthTokenFile    string
}

func Load() (*Config, error) {
	godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		return nil, err
	}

	debug, err := strconv.ParseBool(getEnv("DEBUG", "true"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:              port,
		Debug:             debug,
		OAuthClientSecret: getEnv("OAUTH_CLIENT_SECRET", "credentials.json"),
		OAuthTokenFile:    getEnv("OAUTH_TOKEN_FILE", "token.json"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
