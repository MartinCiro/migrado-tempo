package services

import (
	"email/internal/core/domain"
	"email/internal/core/ports"
)

type emailService struct {
	emailRepo ports.EmailRepository
}

func NewEmailService(emailRepo ports.EmailRepository) ports.EmailService {
	return &emailService{
		emailRepo: emailRepo,
	}
}

func (s *emailService) GetUnreadEmails() ([]domain.Email, error) {
	criteria := domain.EmailCriteria{
		Unread: true,
	}

	return s.emailRepo.SearchEmails(criteria)
}
