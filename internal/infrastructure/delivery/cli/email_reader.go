package cli

import (
	"context"
	"email/config"
	"email/internal/core/ports"
	"fmt"
	"log"
)

type EmailReader struct {
	emailService     ports.EmailService
	executionService ports.ExecutionService
	config           *config.EmailConfig
}

func NewEmailReader(executionService ports.ExecutionService, config *config.EmailConfig) *EmailReader {
	return &EmailReader{
		executionService: executionService,
		config:           config,
	}
}

func (er *EmailReader) Run(ctx context.Context) {
	if er.executionService == nil {
		log.Println("❌ ExecutionService no está inicializado")
		return
	}

	if er.config != nil && er.config.Debug {
		fmt.Println("🔧 Running in DEBUG mode")
	}

	// Ejecución inmediata
	fmt.Println("🔄 Starting initial execution...")
	if err := er.executionService.Run(ctx); err != nil {
		log.Printf("❌ Initial execution failed: %v", err)
	} else {
		fmt.Println("✅ Initial execution completed successfully")
	}

	// ✅ TERMINAR INMEDIATAMENTE después de la ejecución
	fmt.Println("🏁 Application finished")
}

/* func (er *EmailReader) runPeriodically(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(er.config.ScrapingInterval) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//fmt.Printf("\n🔄 Periodic execution at %s...\n", time.Now().Format("2006-01-02 15:04:05"))

			if err := er.executionService.Run(ctx); err != nil {
				log.Printf("❌ Periodic execution failed: %v", err)
			} else {
				fmt.Println("✅ Periodic execution completed successfully")
			}

		case <-ctx.Done():
			fmt.Println("🛑 Stopping email reader...")
			return
		}
	}
} */
