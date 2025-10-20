package services

import (
	"context"
	"fmt"
	"time"

	"email/config"
	"email/internal/core/domain"
	"email/internal/core/ports"
)

type invoiceService struct {
	xmlProcessor ports.XMLProcessor
	fileRepo     ports.FileRepository
	config       *config.InvoiceConfig
}

func NewInvoiceService(
	xmlProcessor ports.XMLProcessor,
	fileRepo ports.FileRepository,
	config *config.InvoiceConfig,
) ports.InvoiceService {
	return &invoiceService{
		xmlProcessor: xmlProcessor,
		fileRepo:     fileRepo,
		config:       config,
	}
}

func (s *invoiceService) GenerateInvoiceData(xmlInvoice *domain.XMLInvoice) *domain.InvoiceData {
	return s.convertToInvoiceData(xmlInvoice)
}

func (s *invoiceService) ProcessInvoiceFromXML(ctx context.Context, xmlData string) (*domain.InvoiceData, error) {
	// Procesar XML
	result, err := s.xmlProcessor.ProcessXML(xmlData)
	if err != nil {
		return nil, fmt.Errorf("error processing XML: %w", err)
	}

	if result.Error {
		return nil, fmt.Errorf("XML processing error: %s", result.Msg)
	}

	// Convertir a estructura de datos para API
	invoiceData := s.convertToInvoiceData(result.Data)

	return invoiceData, nil
}

func (s *invoiceService) ProcessInvoiceFromFile(ctx context.Context, xmlFilePath string) (*domain.InvoiceData, error) {
	// Leer archivo XML
	xmlData, err := s.fileRepo.ReadFile(xmlFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading XML file: %w", err)
	}

	return s.ProcessInvoiceFromXML(ctx, string(xmlData))
}

func (s *invoiceService) convertToInvoiceData(xmlInvoice *domain.XMLInvoice) *domain.InvoiceData {
	invoiceData := &domain.InvoiceData{
		FEVIdFac:               xmlInvoice.IDFactura,
		ProductQuantity:        xmlInvoice.InvoiceDetails.Quantity,
		ProductSubtotal:        xmlInvoice.InvoiceDetails.Subtotal,
		ProductTaxes:           xmlInvoice.InvoiceDetails.Taxes,
		ProductIVA:             xmlInvoice.InvoiceDetails.TaxPercentageIVA,
		ProductVlrTotalChecked: xmlInvoice.PayableAmount,
		ProductVlrTotal:        xmlInvoice.TaxableAmount,
		DiscountGlobalApplied:  xmlInvoice.DiscountAmountApplied,
		IssueDate:              s.formatDate(xmlInvoice.IssueDate),
		PaymentDueDate:         s.formatDate(xmlInvoice.PaymentDueDate),
		SellerNIT:              xmlInvoice.Seller.NIT,
		BuyerNIT:               xmlInvoice.Buyer.NIT,
		SocialReasonSeller:     xmlInvoice.Seller.Name,
		SocialReasonBuyer:      xmlInvoice.Buyer.Name,
		ProductDescription:     xmlInvoice.InvoiceDetails.ProductDescription,
		ProductVlrUnit:         xmlInvoice.InvoiceDetails.UnitPrice,
		NITSeller:              xmlInvoice.Seller.NIT,
		NITBuyer:               xmlInvoice.Buyer.NIT,
		Estado:                 "Sin procesar",
	}

	return invoiceData
}

func (s *invoiceService) formatDate(date *time.Time) string {
	if date == nil {
		return ""
	}
	return date.Format("2006-01-02")
}

// Implementación de ports.InvoiceService
func (s *invoiceService) ValidateInvoice(ctx context.Context, invoice *domain.InvoiceData) error {
	// Validaciones de negocio
	if invoice.FEVIdFac == "" {
		return fmt.Errorf("ID de factura requerido")
	}

	if invoice.SellerNIT == "" {
		return fmt.Errorf("NIT del vendedor requerido")
	}

	if invoice.BuyerNIT == "" {
		return fmt.Errorf("NIT del comprador requerido")
	}

	return nil
}
