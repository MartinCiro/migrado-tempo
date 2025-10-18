package main

import (
	"fmt"
	"log"
	"os"

	"email/config"
	"email/internal/application/services"
	gmailapi "email/internal/infrastructure/adapters/gmail_api"
	"email/internal/infrastructure/adapters/oauth2"
	"email/internal/infrastructure/adapters/repositories"
	"email/internal/infrastructure/delivery/cli"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	fmt.Printf("Starting email reader (Debug: %t)\n", cfg.Debug)

	// Verificar que el archivo de credenciales existe
	if _, err := os.Stat(cfg.OAuthClientSecret); os.IsNotExist(err) {
		log.Fatalf("OAuth client secret file not found: %s", cfg.OAuthClientSecret)
	}

	// Configurar OAuth2
	scopes := []string{"https://www.googleapis.com/auth/gmail.readonly"}
	oauthManager, err := oauth2.NewOAuth2Manager(cfg.OAuthClientSecret, cfg.OAuthTokenFile, scopes)
	if err != nil {
		log.Fatal("Error creating OAuth2 manager:", err)
	}

	// Obtener cliente Gmail
	gmailService, err := oauthManager.GetClient()
	if err != nil {
		log.Fatal("Error getting Gmail client:", err)
	}

	// Configurar dependencias
	gmailClient := gmailapi.NewGmailClient(gmailService)
	emailRepo := repositories.NewEmailRepository(gmailClient)
	emailService := services.NewEmailService(emailRepo)
	emailReader := cli.NewEmailReader(emailService)

	// Conectar al repositorio
	if err := emailRepo.Connect(); err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer emailRepo.Disconnect()

	// Ejecutar la aplicación
	emailReader.Run()
}
