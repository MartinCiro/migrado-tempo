package config

import (
	"strconv"

	"github.com/joho/godotenv"
)

type EmailConfig struct {
	Port              int
	Debug             bool
	OAuthClientSecret string
	OAuthTokenFile    string
	ScrapingInterval  int // Intervalo en minutos para revisar emails
	MaxEmailsPerBatch int // Límite de emails por procesamiento
}

func LoadEmailConfig() (*EmailConfig, error) {
	godotenv.Load()

	port, err := strconv.Atoi(getEnv("EMAIL_PORT", "8080"))
	if err != nil {
		return nil, err
	}

	debug, err := strconv.ParseBool(getEnv("EMAIL_DEBUG", "true"))
	if err != nil {
		return nil, err
	}

	scrapingInterval, _ := strconv.Atoi(getEnv("SCRAPING_INTERVAL", "5"))
	maxEmails, _ := strconv.Atoi(getEnv("MAX_EMAILS_PER_BATCH", "50"))

	return &EmailConfig{
		Port:              port,
		Debug:             debug,
		OAuthClientSecret: getEnv("OAUTH_CLIENT_SECRET", "credentials.json"),
		OAuthTokenFile:    getEnv("OAUTH_TOKEN_FILE", "token.json"),
		ScrapingInterval:  scrapingInterval,
		MaxEmailsPerBatch: maxEmails,
	}, nil
}
