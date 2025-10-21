package config

import (
	"strconv"
	"time"
)

type WebhookConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Enabled      bool
}

func LoadWebhookConfig() *WebhookConfig {
	// Usar variables de entorno con valores por defecto
	host := getEnv("WEBHOOK_HOST", "0.0.0.0")
	port := getEnv("WEBHOOK_PORT", "8080")

	// Timeouts en segundos
	readTimeout, _ := strconv.Atoi(getEnv("WEBHOOK_READ_TIMEOUT", "15"))
	writeTimeout, _ := strconv.Atoi(getEnv("WEBHOOK_WRITE_TIMEOUT", "15"))
	idleTimeout, _ := strconv.Atoi(getEnv("WEBHOOK_IDLE_TIMEOUT", "60"))

	enabled, _ := strconv.ParseBool(getEnv("WEBHOOK_ENABLED", "true"))

	return &WebhookConfig{
		Host:         host,
		Port:         port,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
		Enabled:      enabled,
	}
}

// Método helper para obtener la dirección completa
func (wc *WebhookConfig) GetAddress() string {
	return wc.Host + ":" + wc.Port
}

// Método para obtener URL base (útil para logs)
func (wc *WebhookConfig) GetBaseURL() string {
	return "http://" + wc.GetAddress()
}
