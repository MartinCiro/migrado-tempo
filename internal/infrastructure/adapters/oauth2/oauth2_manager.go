package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type OAuth2Manager struct {
	config    *oauth2.Config
	tokenFile string
	token     *oauth2.Token
}

func NewOAuth2Manager(clientSecretFile, tokenFile string, scopes []string) (*OAuth2Manager, error) {
	b, err := os.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	return &OAuth2Manager{
		config:    config,
		tokenFile: tokenFile,
	}, nil
}

func (o *OAuth2Manager) GetClient() (*gmail.Service, error) {
	if err := o.loadToken(); err != nil {
		if err := o.authenticate(); err != nil {
			return nil, err
		}
	}

	// Verificar si el token necesita refresh
	if !o.token.Valid() {
		if err := o.refreshToken(); err != nil {
			if err := o.authenticate(); err != nil {
				return nil, err
			}
		}
	}

	return gmail.NewService(context.Background(), option.WithTokenSource(o.config.TokenSource(context.Background(), o.token)))
}

func (o *OAuth2Manager) loadToken() error {
	if _, err := os.Stat(o.tokenFile); os.IsNotExist(err) {
		return fmt.Errorf("token file does not exist")
	}

	f, err := os.Open(o.tokenFile)
	if err != nil {
		return err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	if err != nil {
		return err
	}

	o.token = token
	return nil
}

func (o *OAuth2Manager) authenticate() error {
	authURL := o.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	fmt.Print("Enter authorization code: ")
	if _, err := fmt.Scan(&authCode); err != nil {
		return fmt.Errorf("unable to read authorization code: %v", err)
	}

	token, err := o.config.Exchange(context.TODO(), authCode)
	if err != nil {
		return fmt.Errorf("unable to retrieve token from web: %v", err)
	}

	o.token = token
	return o.saveToken()
}

func (o *OAuth2Manager) refreshToken() error {
	if o.token.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	token, err := o.config.TokenSource(context.TODO(), o.token).Token()
	if err != nil {
		return fmt.Errorf("unable to refresh token: %v", err)
	}

	o.token = token
	return o.saveToken()
}

func (o *OAuth2Manager) saveToken() error {
	fmt.Printf("Saving credential file to: %s\n", o.tokenFile)
	f, err := os.OpenFile(o.tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(o.token)
}

func (o *OAuth2Manager) IsAuthenticated() bool {
	return o.token != nil && o.token.Valid()
}
