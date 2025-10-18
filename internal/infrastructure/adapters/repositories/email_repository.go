package repositories

import (
	"email/internal/core/domain"
	gmailapi "email/internal/infrastructure/adapters/gmail_api"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

type EmailRepository struct {
	gmailClient *gmailapi.GmailClient
}

func NewEmailRepository(gmailClient *gmailapi.GmailClient) *EmailRepository {
	return &EmailRepository{
		gmailClient: gmailClient,
	}
}

func (r *EmailRepository) Connect() error {
	return nil
}

func (r *EmailRepository) Disconnect() error {
	return nil
}

func (r *EmailRepository) SearchEmails(criteria domain.EmailCriteria) ([]domain.Email, error) {
	messages, err := r.gmailClient.ListUnreadMessages()
	if err != nil {
		return nil, err
	}

	var emails []domain.Email
	for i, msg := range messages {
		email := domain.Email{
			// Solución 1: Usar el índice como ID temporal
			ID: uint32(i),
			// Solución 2: Convertir el InternalDate a uint32 (si es seguro)
			// ID: uint32(msg.InternalDate), // CUIDADO: Puede haber overflow
			Subject: r.extractHeader(msg.Payload.Headers, "Subject"),
			From:    r.extractHeader(msg.Payload.Headers, "From"),
			Date:    time.Unix(msg.InternalDate/1000, 0),
			Read:    !containsLabel(msg.LabelIds, "UNREAD"),
		}

		emails = append(emails, email)
	}

	return emails, nil
}

// Método alternativo si quieres usar el ID del mensaje de Gmail
func (r *EmailRepository) SearchEmailsWithMessageId(criteria domain.EmailCriteria) ([]domain.Email, error) {
	messages, err := r.gmailClient.ListUnreadMessages()
	if err != nil {
		return nil, err
	}

	var emails []domain.Email
	for _, msg := range messages {
		// Convertir el ID del mensaje (string) a un número
		// Esto es un ejemplo - necesitarías una forma de convertir string a uint32
		id, _ := strconv.ParseUint(msg.Id, 10, 32)

		email := domain.Email{
			ID:      uint32(id),
			Subject: r.extractHeader(msg.Payload.Headers, "Subject"),
			From:    r.extractHeader(msg.Payload.Headers, "From"),
			Date:    time.Unix(msg.InternalDate/1000, 0),
			Read:    !containsLabel(msg.LabelIds, "UNREAD"),
		}

		emails = append(emails, email)
	}

	return emails, nil
}

func (r *EmailRepository) extractHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, header := range headers {
		if strings.EqualFold(header.Name, name) {
			return header.Value
		}
	}
	return ""
}

func containsLabel(labels []string, target string) bool {
	for _, label := range labels {
		if label == target {
			return true
		}
	}
	return false
}
