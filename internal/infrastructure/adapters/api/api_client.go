package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"email/internal/core/domain"
	"email/internal/core/ports"
)

type apiClient struct {
	baseURL    string
	endpoints  map[string]string
	httpClient *http.Client
}

func NewAPIClient(baseURL string, endpoints map[string]string) ports.APIClient {
	return &apiClient{
		baseURL:   baseURL,
		endpoints: endpoints,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *apiClient) Login(ctx context.Context, credentials map[string]string) (*domain.AuthResponse, error) {
	endpoint := c.getEndpoint("login")
	url := c.buildURL(endpoint)

	loginData := map[string]string{
		"correo": credentials["correo"],
		"passwd": credentials["passwd"],
	}

	responseBody, err := c.doRequest(ctx, "POST", url, loginData, "")
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}

	var authResponse domain.AuthResponse
	if err := json.Unmarshal(responseBody, &authResponse); err != nil {
		return nil, fmt.Errorf("error parsing auth response: %w", err)
	}

	if !authResponse.OK || authResponse.Result.Tokens.Access == "" {
		return nil, fmt.Errorf("login failed: %s", authResponse.Message)
	}

	////fmt.Printf("✅ Login successful\n")
	return &authResponse, nil
}

func (c *apiClient) GetNITs(ctx context.Context, token string, id *string) ([]domain.NIT, error) {
	endpoint := c.getEndpoint("nits")
	url := c.buildURL(endpoint)

	if id != nil {
		url = fmt.Sprintf("%s/%s", url, *id)
	}

	response, err := c.doRequest(ctx, "GET", url, nil, token)
	if err != nil {
		return nil, fmt.Errorf("get NITs request failed: %w", err)
	}

	var apiResponse domain.NITResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, fmt.Errorf("error parsing NITs response: %w", err)
	}

	if !apiResponse.OK {
		return nil, fmt.Errorf("API error: %s", apiResponse.Message)
	}

	return apiResponse.Result, nil
}

func (c *apiClient) SendInvoice(ctx context.Context, token string, invoice *domain.InvoiceData) (*domain.APIResponse, error) {
	endpoint := c.getEndpoint("facturas")
	url := c.buildURL(endpoint)

	response, err := c.doRequest(ctx, "POST", url, invoice, token)
	if err != nil {
		return nil, fmt.Errorf("send invoice request failed: %w", err)
	}

	var apiResponse domain.APIResponse
	if err := json.Unmarshal(response, &apiResponse); err != nil {
		return nil, fmt.Errorf("error parsing invoice response: %w", err)
	}

	// Debug: mostrar la estructura de la respuesta
	//fmt.Printf("📥 Invoice API Response - OK: %t, Status: %d, Result type: %T\n", apiResponse.OK, apiResponse.StatusCode, apiResponse.Result)

	// Manejar caso especial de factura duplicada
	if apiResponse.StatusCode == 409 || apiResponse.StatusCode == 400 {
		errorMsg := ""

		// Extraer mensaje de error dependiendo del tipo de Result
		switch result := apiResponse.Result.(type) {
		case string:
			errorMsg = result
		case map[string]interface{}:
			if msg, ok := result["message"].(string); ok {
				errorMsg = msg
			}
		}

		if strings.Contains(errorMsg, "ya existe") || strings.Contains(errorMsg, "Duplicate entry") {
			fmt.Printf("ℹ️  Factura ya existe en el sistema: %s - %s\n", invoice.FEVIdFac, errorMsg)
			// Considerar como éxito (no es un error)
			apiResponse.OK = true
			apiResponse.StatusCode = 200
			apiResponse.Message = fmt.Sprintf("Factura ya existe: %s", invoice.FEVIdFac)
			apiResponse.Duplicate = true
		}
	}

	return &apiResponse, nil
}

func (c *apiClient) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	return nil, fmt.Errorf("refresh token not implemented")
}

func (c *apiClient) doRequest(ctx context.Context, method, url string, data interface{}, token string) ([]byte, error) {
	var body io.Reader

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("error marshaling request data: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Host", "localhost")

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errorResponse map[string]interface{}
		if err := json.Unmarshal(responseBody, &errorResponse); err == nil {
			return nil, fmt.Errorf("API error %d: %v", resp.StatusCode, errorResponse)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(responseBody))
	}

	return responseBody, nil
}

func (c *apiClient) getEndpoint(key string) string {
	if endpoint, exists := c.endpoints[key]; exists {
		return endpoint
	}
	return key
}

func (c *apiClient) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, endpoint)
}
