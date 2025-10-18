package ports

import "email/internal/core/domain"

type AuthService interface {
	Authenticate() error
	IsAuthenticated() bool
	GetToken() (*domain.TokenInfo, error)
}
