package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"email/config"
	"email/internal/core/ports"
)

type webhookService struct {
	invoiceService ports.InvoiceService
	authService    ports.AuthService
	fileRepo       ports.FileRepository
	config         *config.AppConfig
}

func NewWebhookService(
	invoiceService ports.InvoiceService,
	authService ports.AuthService,
	fileRepo ports.FileRepository,
	config *config.AppConfig,
) ports.WebhookService {
	return &webhookService{
		invoiceService: invoiceService,
		authService:    authService,
		fileRepo:       fileRepo,
		config:         config,
	}
}

func (s *webhookService) ProcessXMLFromWebhook(ctx context.Context, request *ports.WebhookRequest) (*ports.WebhookResponse, error) {
	fmt.Printf("🔗 Procesando XML via webhook: %s\n", request.XMLFileName)

	// 1. Guardar archivos temporales
	xmlPath, err := s.saveTempFile(request.XMLContent, request.XMLFileName, "xml")
	if err != nil {
		return s.errorResponse("error guardando XML", err), nil
	}
	defer os.Remove(xmlPath)

	var pdfPath string
	if request.PDFContent != "" {
		pdfPath, err = s.saveTempFile(request.PDFContent, request.PDFFileName, "pdf")
		if err != nil {
			return s.errorResponse("error guardando PDF", err), nil
		}
		defer os.Remove(pdfPath)
	}

	// 2. Procesar XML
	invoiceData, err := s.invoiceService.ProcessInvoiceFromFile(ctx, xmlPath)
	if err != nil {
		return s.errorResponse("error procesando XML", err), nil
	}

	// 3. Convertir a base64
	xmlBase64, err := s.fileRepo.FileToBase64(xmlPath)
	if err != nil {
		return s.errorResponse("error codificando XML", err), nil
	}

	var pdfBase64 string
	if pdfPath != "" {
		pdfBase64, err = s.fileRepo.FileToBase64(pdfPath)
		if err != nil {
			return s.errorResponse("error codificando PDF", err), nil
		}
	}

	// 4. Completar datos
	invoiceData.XMLString = xmlBase64
	invoiceData.PDFBase64 = pdfBase64
	invoiceData.Email = "webhook@system"

	// 5. Enviar a API
	response, err := s.authService.SendInvoice(ctx, request.Token, invoiceData)
	if err != nil {
		return s.errorResponse("error enviando a API", err), nil
	}

	if !response.OK {
		return s.errorResponse("API retornó error", fmt.Errorf(response.Message)), nil
	}

	return &ports.WebhookResponse{
		Success:   true,
		InvoiceID: invoiceData.FEVIdFac,
		Message:   "Factura procesada exitosamente via webhook",
	}, nil
}

func (s *webhookService) saveTempFile(content, filename, fileType string) (string, error) {
	tempDir := filepath.Join(s.config.Paths.ZipFolder, "temp_webhook")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	// Decodificar base64 si es necesario
	var fileData []byte
	if fileType == "xml" || fileType == "pdf" {
		var err error
		fileData, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			// Si falla la decodificación, asumir que es texto plano
			fileData = []byte(content)
		}
	} else {
		fileData = []byte(content)
	}

	filePath := filepath.Join(tempDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), filename))
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *webhookService) errorResponse(message string, err error) *ports.WebhookResponse {
	return &ports.WebhookResponse{
		Success: false,
		Error:   fmt.Sprintf("%s: %v", message, err),
	}
}
