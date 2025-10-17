package cli

import (
	"email/internal/core/ports"
	"fmt"
	"log"
)

type EmailReader struct {
	emailService ports.EmailService
}

func NewEmailReader(emailService ports.EmailService) *EmailReader {
	return &EmailReader{
		emailService: emailService,
	}
}

func (er *EmailReader) Run() {
	emails, err := er.emailService.GetUnreadEmails()
	if err != nil {
		log.Fatal("Error getting unread emails:", err)
	}

	if len(emails) == 0 {
		fmt.Println("No hay mensajes no leídos")
		return
	}

	fmt.Printf("Mensajes no leídos: %d\n", len(emails))
	for _, email := range emails {
		fmt.Printf("Asunto: %s\n", email.Subject)
		fmt.Printf("De: %s\n", email.From)
		fmt.Printf("Fecha: %s\n", email.Date.Format("2006-01-02 15:04:05"))
		fmt.Println("---")
	}
}
