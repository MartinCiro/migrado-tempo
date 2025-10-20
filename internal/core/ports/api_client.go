package ports

import (
	"context"
	"email/internal/core/domain" // Cambiar email por email_scrapper
)

// APIClient define las operaciones de comunicación con APIs externas
type APIClient interface {
	Login(ctx context.Context, credentials map[string]string) (*domain.AuthResponse, error)
	GetNITs(ctx context.Context, token string, id *string) ([]domain.NIT, error)
	SendInvoice(ctx context.Context, token string, invoice *domain.InvoiceData) (*domain.APIResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
}
