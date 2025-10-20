package repositories

import (
	"context"

	"email/internal/core/domain"
	"email/internal/core/ports"
)

type emailRepository struct {
	gmailClient ports.EmailRepository
}

func NewEmailRepository(gmailClient ports.EmailRepository) ports.EmailRepository {
	return &emailRepository{
		gmailClient: gmailClient,
	}
}

func (r *emailRepository) Connect() error {
	return r.gmailClient.Connect()
}

func (r *emailRepository) Disconnect() error {
	return r.gmailClient.Disconnect()
}

func (r *emailRepository) GetLabelID(ctx context.Context, labelName string) (string, error) {
	return r.gmailClient.GetLabelID(ctx, labelName)
}

func (r *emailRepository) GetEmailsByLabel(ctx context.Context, labelName string) ([]domain.Email, error) {
	return r.gmailClient.GetEmailsByLabel(ctx, labelName)
}

func (r *emailRepository) FindZipAttachments(ctx context.Context, messageID string) ([][]byte, []string, error) {
	return r.gmailClient.FindZipAttachments(ctx, messageID)
}

func (r *emailRepository) SearchEmails(criteria domain.EmailCriteria) ([]domain.Email, error) {
	filter := domain.EmailFilter{
		From:    criteria.From,
		Subject: criteria.Subject,
		Since:   criteria.Since,
	}
	return r.gmailClient.GetEmails(context.Background(), filter)
}

func (r *emailRepository) GetEmails(ctx context.Context, filter domain.EmailFilter) ([]domain.Email, error) {
	return r.gmailClient.GetEmails(ctx, filter)
}

func (r *emailRepository) SaveEmail(ctx context.Context, email *domain.Email) error {
	return r.gmailClient.SaveEmail(ctx, email)
}

func (r *emailRepository) DeleteEmail(ctx context.Context, emailID string) error {
	return r.gmailClient.DeleteEmail(ctx, emailID)
}
