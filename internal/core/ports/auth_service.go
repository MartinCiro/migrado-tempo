package ports

import (
	"context"
	"email/internal/core/domain"

	"google.golang.org/api/gmail/v1"
)

type AuthService interface {
	Authenticate() error
	IsAuthenticated() bool
	GetToken() (*domain.TokenInfo, error)
	RefreshToken() error
	// Métodos para API
	Login(ctx context.Context, credentials map[string]string) (string, error)
	GetNITs(ctx context.Context, token string, id *string) ([]domain.NIT, error)
	SendInvoice(ctx context.Context, token string, invoice *domain.InvoiceData) (*domain.APIResponse, error)
	// Método para obtener cliente Gmail
	GetClient() (*gmail.Service, error)
}
