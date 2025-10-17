package main

import (
	"fmt"
	"log"

	"email/config"
	"email/internal/application/services"
	"email/internal/infrastructure/adapters/imap"
	"email/internal/infrastructure/adapters/repositories"
	"email/internal/infrastructure/delivery/cli"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	fmt.Printf("Server running on port %d (Debug: %t)\n", cfg.Port, cfg.Debug)

	// Configurar dependencias
	imapClient := imap.NewIMAPClient(
		"imap.gmail.com:993",
		cfg.GmailEmail,
		cfg.GmailAppPassword,
	)

	emailRepo := repositories.NewEmailRepository(imapClient)
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
