package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	"email/internal/core/domain"
	"email/internal/core/ports"
)

type oauth2Manager struct {
	config     *oauth2.Config
	tokenFile  string
	httpClient *http.Client
}

func NewOAuth2Manager(credentialsFile, tokenFile string, scopes []string) (ports.AuthService, error) {

	// Leer credenciales
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	// Configurar OAuth2
	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	manager := &oauth2Manager{
		config:     config,
		tokenFile:  tokenFile,
		httpClient: &http.Client{},
	}

	return manager, nil
}

func (m *oauth2Manager) Authenticate() error {
	ctx := context.Background()

	// Cargar token existente
	token, err := m.tokenFromFile()
	if err != nil {
		fmt.Println("🔑 No se encontró token existente, iniciando autenticación...")
		// Si no hay token, obtener uno nuevo
		token, err = m.getTokenFromWeb(ctx)
		if err != nil {
			return fmt.Errorf("unable to get token from web: %w", err)
		}

		// Guardar token
		if err := m.saveToken(token); err != nil {
			return fmt.Errorf("unable to save token: %w", err)
		}
		fmt.Println("✅ Token guardado exitosamente")
	}

	// ⚠️ ELIMINAR la lógica de refresh automático
	// Solo verificar si el token es válido, pero NO refrescar
	/* if m.tokenNeedsRefresh(token) {
		fmt.Println("⚠️  Token expirado o próximo a expirar, pero el refresh automático está deshabilitado")
		fmt.Println("ℹ️   Si tienes problemas de autenticación, elimina token.json para generar uno nuevo")
		// No hacemos refresh, continuamos con el token actual
	} */

	// Configurar cliente HTTP
	m.httpClient = m.config.Client(ctx, token)
	//fmt.Println("✅ Autenticación completada exitosamente")
	return nil
}

func (m *oauth2Manager) IsAuthenticated() bool {
	token, err := m.tokenFromFile()
	if err != nil {
		return false
	}

	return !m.tokenNeedsRefresh(token)
}

func (m *oauth2Manager) GetToken() (*domain.TokenInfo, error) {
	oauthToken, err := m.tokenFromFile()
	if err != nil {
		return nil, fmt.Errorf("unable to get token: %w", err)
	}

	tokenInfo := &domain.TokenInfo{
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
		Expiry:       oauthToken.Expiry,
		TokenType:    oauthToken.TokenType,
	}

	return tokenInfo, nil
}

func (m *oauth2Manager) RefreshToken() error {
	ctx := context.Background()

	token, err := m.tokenFromFile()
	if err != nil {
		return fmt.Errorf("unable to load token for refresh: %w", err)
	}

	tokenSource := m.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("unable to refresh token: %w", err)
	}

	if newToken.AccessToken != token.AccessToken {
		if err := m.saveToken(newToken); err != nil {
			return fmt.Errorf("unable to save refreshed token: %w", err)
		}
	}

	return nil
}

func (m *oauth2Manager) GetClient() (*gmail.Service, error) {
	if !m.IsAuthenticated() {
		if err := m.Authenticate(); err != nil {
			return nil, fmt.Errorf("authentication required: %w", err)
		}
	}

	ctx := context.Background()
	return gmail.NewService(ctx, option.WithHTTPClient(m.httpClient))
}

// Helpers internos
func (m *oauth2Manager) tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(m.tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (m *oauth2Manager) saveToken(token *oauth2.Token) error {
	//fmt.Printf("Saving token file to: %s\n", m.tokenFile)

	f, err := os.OpenFile(m.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %w", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

func (m *oauth2Manager) getTokenFromWeb(ctx context.Context) (*oauth2.Token, error) {
	//authURL := m.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	//fmt.Printf("Go to the following link in your browser then type the "+ "authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	tok, err := m.config.Exchange(ctx, authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}
	return tok, nil
}

func (m *oauth2Manager) tokenNeedsRefresh(token *oauth2.Token) bool {
	if token == nil {
		return true
	}

	// Solo considerar expirado si ya pasó la fecha de expiración
	return token.Expiry.Before(time.Now())
}

func (m *oauth2Manager) Login(ctx context.Context, credentials map[string]string) (string, error) {
	// Para OAuth2, el login se maneja de forma diferente
	// Puedes implementar esto si necesitas integración con tu API
	return "", fmt.Errorf("OAuth2 login not implemented - use Authenticate() instead")
}

func (m *oauth2Manager) GetNITs(ctx context.Context, token string, id *string) ([]domain.NIT, error) {
	// OAuth2 manager no maneja NITs
	return nil, fmt.Errorf("GetNITs not implemented in OAuth2 manager")
}

func (m *oauth2Manager) SendInvoice(ctx context.Context, token string, invoice *domain.InvoiceData) (*domain.APIResponse, error) {
	// OAuth2 manager no envía facturas
	return nil, fmt.Errorf("SendInvoice not implemented in OAuth2 manager")
}
