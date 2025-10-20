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

	// IMPRIMIR DATOS EXTRAÍDOS DEL XML
	/* fmt.Println("=== DATOS EXTRAÍDOS DEL XML ===")
	fmt.Printf("ID Factura: %s\n", result.Data.IDFactura)
	fmt.Printf("ProfileID: %s\n", result.Data.ProfileID)
	fmt.Printf("Fecha Emisión: %v\n", result.Data.IssueDate)
	fmt.Printf("Fecha Vencimiento: %v\n", result.Data.PaymentDueDate)
	fmt.Printf("Monto Total: %s\n", result.Data.PayableAmount)
	fmt.Printf("Monto Gravable: %s\n", result.Data.TaxableAmount)
	fmt.Printf("Descuento: %s\n", result.Data.DiscountAmountApplied)
	fmt.Printf("Cantidad: %s\n", result.Data.InvoicedQuantity)
	fmt.Printf("Vendedor: %s (NIT: %s)\n", result.Data.Seller.Name, result.Data.Seller.NIT)
	fmt.Printf("Comprador: %s (NIT: %s)\n", result.Data.Buyer.Name, result.Data.Buyer.NIT)
	fmt.Printf("Descripción Producto: %s\n", result.Data.InvoiceDetails.ProductDescription)
	fmt.Printf("Cantidad Detalle: %s\n", result.Data.InvoiceDetails.Quantity)
	fmt.Printf("Precio Unitario: %s\n", result.Data.InvoiceDetails.UnitPrice)
	fmt.Printf("Subtotal: %s\n", result.Data.InvoiceDetails.Subtotal)
	fmt.Printf("IVA: %s%%\n", result.Data.InvoiceDetails.TaxPercentageIVA)
	fmt.Printf("Impuestos: %s\n", result.Data.InvoiceDetails.Taxes)
	fmt.Println("=================================")
 */
	// Convertir a estructura de datos para API
	invoiceData := s.convertToInvoiceData(result.Data)

	// IMPRIMIR DATOS CONVERTIDOS PARA API
	/* fmt.Println("=== DATOS CONVERTIDOS PARA API ===")
	fmt.Printf("FEVIdFac: %s\n", invoiceData.FEVIdFac)
	fmt.Printf("ProductQuantity: %s\n", invoiceData.ProductQuantity)
	fmt.Printf("ProductSubtotal: %s\n", invoiceData.ProductSubtotal)
	fmt.Printf("ProductTaxes: %s\n", invoiceData.ProductTaxes)
	fmt.Printf("ProductIVA: %s\n", invoiceData.ProductIVA)
	fmt.Printf("ProductVlrTotalChecked: %s\n", invoiceData.ProductVlrTotalChecked)
	fmt.Printf("ProductVlrTotal: %s\n", invoiceData.ProductVlrTotal)
	fmt.Printf("DiscountGlobalApplied: %s\n", invoiceData.DiscountGlobalApplied)
	fmt.Printf("IssueDate: %s\n", invoiceData.IssueDate)
	fmt.Printf("PaymentDueDate: %s\n", invoiceData.PaymentDueDate)
	fmt.Printf("SellerNIT: %s\n", invoiceData.SellerNIT)
	fmt.Printf("BuyerNIT: %s\n", invoiceData.BuyerNIT)
	fmt.Printf("SocialReasonSeller: %s\n", invoiceData.SocialReasonSeller)
	fmt.Printf("SocialReasonBuyer: %s\n", invoiceData.SocialReasonBuyer)
	fmt.Printf("ProductDescription: %s\n", invoiceData.ProductDescription)
	fmt.Printf("ProductVlrUnit: %s\n", invoiceData.ProductVlrUnit)
	fmt.Println("===================================") */

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
