package ports

import (
	"email/internal/core/domain"
)

// XMLProcessor define las operaciones para procesar XML de facturas
type XMLProcessor interface {
	ProcessXML(xmlData string) (*domain.XMLProcessingResult, error)
	ExtractInvoiceData(xmlData string) (*domain.XMLInvoice, error)
	ValidateXMLStructure(xmlData string) error
}
