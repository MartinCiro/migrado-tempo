package services

import (
	"context"
	"fmt"

	"email/internal/core/domain"
	"email/internal/core/ports"

	"google.golang.org/api/gmail/v1"
)

type authService struct {
	apiClient ports.APIClient
}

func NewAuthService(apiClient ports.APIClient) ports.AuthService {
	return &authService{
		apiClient: apiClient,
	}
}

func (s *authService) Authenticate() error {
	// Para API auth, esto podría ser diferente
	return nil
}

func (s *authService) IsAuthenticated() bool {
	return true // Simplificado por ahora
}

func (s *authService) GetToken() (*domain.TokenInfo, error) {
	return &domain.TokenInfo{}, nil // Simplificado
}

func (s *authService) RefreshToken() error {
	return nil
}

func (s *authService) Login(ctx context.Context, credentials map[string]string) (string, error) {
	authResponse, err := s.apiClient.Login(ctx, credentials)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	if authResponse.Result.Tokens.Access == "" {
		return "", fmt.Errorf("empty access token received")
	}

	return authResponse.Result.Tokens.Access, nil
}

func (s *authService) GetNITs(ctx context.Context, token string, id *string) ([]domain.NIT, error) {
	nits, err := s.apiClient.GetNITs(ctx, token, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve NITs: %w", err)
	}

	if len(nits) == 0 {
		return nil, fmt.Errorf("no NITs available")
	}

	return nits, nil
}

func (s *authService) SendInvoice(ctx context.Context, token string, invoice *domain.InvoiceData) (*domain.APIResponse, error) {
	response, err := s.apiClient.SendInvoice(ctx, token, invoice)
	if err != nil {
		return nil, fmt.Errorf("failed to send invoice: %w", err)
	}

	if !response.OK && response.StatusCode != 200 {
		return nil, fmt.Errorf("API returned error: %s", response.Message)
	}

	return response, nil
}

func (s *authService) GetClient() (*gmail.Service, error) {
	// Para API auth, no manejamos cliente Gmail directamente
	return nil, fmt.Errorf("GetClient not implemented for API auth service")
}
