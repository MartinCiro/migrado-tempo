package config

import "os"

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// AppConfig contiene ambas configuraciones para inyección de dependencias
type AppConfig struct {
	Email   *EmailConfig
	Invoice *InvoiceConfig
	Paths   *InvoicePaths
	Webhook *WebhookConfig
}

func NewAppConfig(emailConfig *EmailConfig, invoiceConfig *InvoiceConfig, paths *InvoicePaths, webhookConfig *WebhookConfig) *AppConfig {
	return &AppConfig{
		Email:   emailConfig,
		Invoice: invoiceConfig,
		Paths:   paths,
		Webhook: LoadWebhookConfig(),
	}
}
