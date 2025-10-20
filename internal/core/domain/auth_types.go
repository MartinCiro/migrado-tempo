package domain

import "time"

// TokenInfo representa la información del token OAuth2
type TokenInfo struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
	Scopes       []string  `json:"scopes"`
}

// OAuth2Config configuración para OAuth2
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_uri"`
	Scopes       []string `json:"scopes"`
}

// OAuthConfig configuración alternativa (si es necesaria para compatibilidad)
type OAuthConfig struct {
	ClientSecretFile string   `json:"client_secret_file"`
	TokenFile        string   `json:"token_file"`
	Scopes           []string `json:"scopes"`
}
