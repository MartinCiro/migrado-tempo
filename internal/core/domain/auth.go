package domain

type OAuthConfig struct {
	ClientSecretFile string
	TokenFile        string
	Scopes           []string
}

type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expiry       string `json:"expiry"`
}
