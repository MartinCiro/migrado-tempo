package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"email/config"
	"email/internal/application/services"
	"email/internal/infrastructure/adapters/api"
	"email/internal/infrastructure/adapters/crypto"
	"email/internal/infrastructure/adapters/files"
	gmailapi "email/internal/infrastructure/adapters/gmail_api"
	"email/internal/infrastructure/adapters/oauth2"
	"email/internal/infrastructure/adapters/repositories"
	xmlprocessor "email/internal/infrastructure/adapters/xml_processor"
	"email/internal/infrastructure/delivery/cli"
)

func main() {
	ctx := context.Background()

	// =========================================================================
	// Cargar configuraciones
	// =========================================================================
	emailConfig, err := config.LoadEmailConfig()
	if err != nil {
		log.Fatal("Error loading email config:", err)
	}

	invoiceConfig, err := config.LoadInvoiceConfig("config.json")
	if err != nil {
		log.Fatal("Error loading invoice config:", err)
	}

	// Construir rutas y crear directorios
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting working directory:", err)
	}

	paths := invoiceConfig.BuildPaths(baseDir)

	if err := paths.EnsureDirectories(); err != nil {
		log.Fatal("Error creating directories:", err)
	}

	appConfig := config.NewAppConfig(emailConfig, invoiceConfig, paths)

	// =========================================================================
	// Configuración OAuth2 y Gmail
	// =========================================================================
	if _, err := os.Stat(emailConfig.OAuthClientSecret); os.IsNotExist(err) {
		log.Fatalf("OAuth client secret file not found: %s", emailConfig.OAuthClientSecret)
	}

	scopes := []string{"https://www.googleapis.com/auth/gmail.readonly"}
	oauthManager, err := oauth2.NewOAuth2Manager(emailConfig.OAuthClientSecret, emailConfig.OAuthTokenFile, scopes)
	if err != nil {
		log.Fatal("Error creating OAuth2 manager:", err)
	}

	gmailService, err := oauthManager.GetClient()
	if err != nil {
		log.Fatal("Error getting Gmail client:", err)
	}

	// =========================================================================
	// Inicializar dependencias - Infraestructura
	// =========================================================================

	// 1. Clientes externos
	gmailAdapter := gmailapi.NewGmailClient(gmailService)
	apiClient := api.NewAPIClient(invoiceConfig.API.Host, invoiceConfig.API.Endpoints)

	// 2. Repositorios
	emailRepo := repositories.NewEmailRepository(gmailAdapter)
	fileRepo := files.NewFileRepository(baseDir)

	// 3. Servicios de infraestructura
	cryptoService, err := crypto.NewCryptoService("ecTNu1JkrN8WOEZQ667dOGOqBcS9Peh0RShN83l1WK0=")
	if err != nil {
		log.Fatal("Error creating crypto service:", err)
	}

	xmlProcessor := xmlprocessor.NewXMLProcessor(invoiceConfig.Causacion)

	// =========================================================================
	// Inicializar servicios de aplicación
	// =========================================================================

	// Servicio de emails
	emailService := services.NewEmailService(emailRepo)

	// Servicio de facturas
	invoiceService := services.NewInvoiceService(xmlProcessor, fileRepo, invoiceConfig)

	// Servicio de autenticación API
	authService := services.NewAuthService(apiClient)

	// Servicio principal de ejecución (orquestador)
	executionService := services.NewExecutionService(
		emailService,
		invoiceService,
		authService,
		fileRepo,
		cryptoService,
		appConfig,
		emailRepo,
	)

	// =========================================================================
	// Inicializar delivery (CLI)
	// =========================================================================
	emailReader := cli.NewEmailReader(executionService, emailConfig)

	// =========================================================================
	// Conectar repositorios
	// =========================================================================
	if err := emailRepo.Connect(); err != nil {
		log.Fatal("Failed to connect to email repository:", err)
	}
	defer emailRepo.Disconnect()

	// =========================================================================
	// Ejecutar la aplicación
	// =========================================================================
	fmt.Println("🚀 Email Scrapper started successfully!")
	emailReader.Run(ctx)
}
