package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/gmail/v1"

	"email/config"
	"email/internal/application/services"
	"email/internal/core/ports"
	"email/internal/infrastructure/adapters/api"
	"email/internal/infrastructure/adapters/crypto"
	"email/internal/infrastructure/adapters/files"
	gmailapi "email/internal/infrastructure/adapters/gmail_api"
	"email/internal/infrastructure/adapters/oauth2"
	"email/internal/infrastructure/adapters/repositories"
	xmlprocessor "email/internal/infrastructure/adapters/xml_processor"
	"email/internal/infrastructure/delivery/cli"
	"email/internal/infrastructure/delivery/web"
)

func main() {
	// =========================================================================
	// Parsear flags de modo de operación
	// =========================================================================
	mode := flag.String("mode", "", "Modo de operación: vacío para proceso completo, 'webhook' para servidor webhook")
	flag.Parse()

	ctx := context.Background()

	// =========================================================================
	// Cargar configuraciones (común a todos los modos)
	// =========================================================================
	emailConfig, err := config.LoadEmailConfig()
	if err != nil {
		log.Fatal("Error loading email config:", err)
	}

	invoiceConfig, err := config.LoadInvoiceConfig("config.json")
	if err != nil {
		log.Fatal("Error loading invoice config:", err)
	}

	// Cargar configuración webhook
	webhookConfig := config.LoadWebhookConfig()

	// Construir rutas y crear directorios
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting working directory:", err)
	}

	paths := invoiceConfig.BuildPaths(baseDir)
	if err := paths.EnsureDirectories(); err != nil {
		log.Fatal("Error creating directories:", err)
	}

	// AppConfig actualizado con webhook
	appConfig := config.NewAppConfig(emailConfig, invoiceConfig, paths, webhookConfig)

	// =========================================================================
	// Configuración OAuth2 y Gmail (solo necesario para modo email)
	// =========================================================================
	var gmailService interface{}
	var emailRepo ports.EmailRepository

	if *mode != "webhook" {
		if _, err := os.Stat(emailConfig.OAuthClientSecret); os.IsNotExist(err) {
			log.Fatalf("OAuth client secret file not found: %s", emailConfig.OAuthClientSecret)
		}

		scopes := []string{"https://www.googleapis.com/auth/gmail.readonly"}
		oauthManager, err := oauth2.NewOAuth2Manager(emailConfig.OAuthClientSecret, emailConfig.OAuthTokenFile, scopes)
		if err != nil {
			log.Fatal("Error creating OAuth2 manager:", err)
		}

		gmailService, err = oauthManager.GetClient()
		if err != nil {
			log.Fatal("Error getting Gmail client:", err)
		}

		// Inicializar adaptador Gmail
		gmailAdapter := gmailapi.NewGmailClient(gmailService.(*gmail.Service))
		emailRepo = repositories.NewEmailRepository(gmailAdapter)
	}

	// =========================================================================
	// Inicializar dependencias comunes - Infraestructura
	// =========================================================================

	// 1. Clientes externos
	apiClient := api.NewAPIClient(invoiceConfig.API.Host, invoiceConfig.API.Endpoints)

	// 2. Repositorios
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

	// Servicio de facturas
	invoiceService := services.NewInvoiceService(xmlProcessor, fileRepo, invoiceConfig)

	// Servicio de autenticación API
	authService := services.NewAuthService(apiClient)

	// =========================================================================
	// Ejecutar según modo
	// =========================================================================
	switch *mode {
	case "webhook":
		runWebhookMode(ctx, invoiceService, authService, fileRepo, appConfig)
	case "": // Modo por defecto - proceso completo
		runFullMode(ctx, emailConfig, invoiceConfig, appConfig, gmailService, emailRepo, fileRepo, cryptoService, xmlProcessor, apiClient)
	default:
		log.Fatalf("Modo no válido: '%s'. Use vacío para proceso completo o 'webhook'", *mode)
	}
}

// runFullMode ejecuta el proceso completo de emails
func runFullMode(
	ctx context.Context,
	emailConfig *config.EmailConfig,
	invoiceConfig *config.InvoiceConfig,
	appConfig *config.AppConfig,
	gmailService interface{},
	emailRepo ports.EmailRepository,
	fileRepo ports.FileRepository,
	cryptoService ports.CryptoService,
	xmlProcessor ports.XMLProcessor,
	apiClient ports.APIClient,
) {
	fmt.Println("🚀 Email Scrapper started successfully!")

	if emailConfig != nil && emailConfig.Debug {
		fmt.Println("🔧 Running in DEBUG mode")
	}

	// =========================================================================
	// Inicializar servicios de aplicación para modo completo
	// =========================================================================

	// Servicio de emails
	emailService := services.NewEmailService(emailRepo)

	// Servicio de facturas (ya inicializado en main)
	invoiceService := services.NewInvoiceService(xmlProcessor, fileRepo, invoiceConfig)

	// Servicio de autenticación API (ya inicializado en main)
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
	emailReader.Run(ctx)
}

// runWebhookMode inicia el servidor webhook
func runWebhookMode(
	ctx context.Context,
	invoiceService ports.InvoiceService,
	authService ports.AuthService,
	fileRepo ports.FileRepository,
	appConfig *config.AppConfig,
) {
	// Verificar si el webhook está habilitado
	if !appConfig.Webhook.Enabled {
		log.Println("ℹ️  Webhook mode is disabled in configuration")
		return
	}

	fmt.Printf("🌐 Iniciando servidor webhook\n")

	// Crear servicio webhook
	webhookService := services.NewWebhookService(
		invoiceService,
		authService,
		fileRepo,
		appConfig,
	)

	webhookHandler := web.NewWebhookHandler(webhookService)

	// Usar la configuración del webhook desde appConfig
	server := web.NewServer(webhookHandler, appConfig.Webhook)

	fmt.Printf("✅ Servidor webhook inicializado en: %s\n", appConfig.Webhook.GetBaseURL())
	fmt.Println("📋 Endpoints disponibles:")
	fmt.Printf("   POST %s/webhook/process-xml\n", appConfig.Webhook.GetBaseURL())
	fmt.Printf("   GET  %s/\n", appConfig.Webhook.GetBaseURL())
	/*
		// Mostrar información de configuración
		fmt.Printf("⚙️  Configuración del servidor:\n")
		fmt.Printf("   Host: %s\n", appConfig.Webhook.Host)
		fmt.Printf("   Port: %s\n", appConfig.Webhook.Port)
		fmt.Printf("   Read Timeout: %v\n", appConfig.Webhook.ReadTimeout)
		fmt.Printf("   Write Timeout: %v\n", appConfig.Webhook.WriteTimeout)
		fmt.Printf("   Idle Timeout: %v\n", appConfig.Webhook.IdleTimeout) */

	if err := server.Start(); err != nil {
		log.Fatal("Error starting webhook server:", err)
	}
}
