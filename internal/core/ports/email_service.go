package ports

import "email/internal/core/domain"

type EmailService interface {
	GetUnreadEmails() ([]domain.Email, error)
}
