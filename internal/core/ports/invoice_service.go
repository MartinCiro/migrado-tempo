package ports

import (
	"context"
	"email/internal/core/domain"
)

// InvoiceService define las operaciones de procesamiento de facturas
type InvoiceService interface {
	ProcessInvoiceFromXML(ctx context.Context, xmlData string) (*domain.InvoiceData, error)
	ProcessInvoiceFromFile(ctx context.Context, xmlFilePath string) (*domain.InvoiceData, error)
	ValidateInvoice(ctx context.Context, invoice *domain.InvoiceData) error
	GenerateInvoiceData(xmlInvoice *domain.XMLInvoice) *domain.InvoiceData
}
