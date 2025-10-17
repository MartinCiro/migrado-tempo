package ports

import "email/internal/core/domain"

type EmailRepository interface {
	Connect() error
	Disconnect() error
	SearchEmails(criteria domain.EmailCriteria) ([]domain.Email, error)
}
