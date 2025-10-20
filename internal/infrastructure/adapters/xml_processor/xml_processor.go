package xmlprocessor

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"email/internal/core/domain"
	"email/internal/core/ports"
)

type xmlProcessor struct {
	causacionMap map[string]string
	namespaces   map[string]string
}

func NewXMLProcessor(causacionMap map[string]string) ports.XMLProcessor {
	return &xmlProcessor{
		causacionMap: causacionMap,
		namespaces: map[string]string{
			"cbc": "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",
			"cac": "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
			"ext": "urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2",
			"sts": "dian:gov:co:facturaelectronica:Structures-2-1",
		},
	}
}

// XML Structures para parsing
type UBLInvoice struct {
	XMLName            xml.Name           `xml:"Invoice"`
	ProfileID          string             `xml:"ProfileID"`
	ID                 string             `xml:"ID"`
	IssueDate          string             `xml:"IssueDate"`
	SenderParty        Party              `xml:"SenderParty"`
	ReceiverParty      Party              `xml:"ReceiverParty"`
	Attachment         Attachment         `xml:"Attachment"`
	TaxTotal           TaxTotal           `xml:"TaxTotal"`
	LegalMonetaryTotal LegalMonetaryTotal `xml:"LegalMonetaryTotal"`
}

type Party struct {
	RegistrationName string    `xml:"RegistrationName"`
	CompanyID        CompanyID `xml:"CompanyID"`
	TaxScheme        TaxScheme `xml:"PartyTaxScheme>TaxScheme"`
}

type CompanyID struct {
	SchemeID string `xml:"schemeID,attr"`
	Value    string `xml:",chardata"`
}

type TaxScheme struct {
	TaxLevelCode string `xml:"TaxLevelCode"`
}

type Attachment struct {
	ExternalReference ExternalReference `xml:"ExternalReference"`
}

type ExternalReference struct {
	Description string `xml:"Description"`
}

type TaxTotal struct {
	TaxAmount   string      `xml:"TaxAmount"`
	TaxSubtotal TaxSubtotal `xml:"TaxSubtotal"`
}

type TaxSubtotal struct {
	TaxableAmount string      `xml:"TaxableAmount"`
	TaxAmount     string      `xml:"TaxAmount"`
	TaxCategory   TaxCategory `xml:"TaxCategory"`
}

type TaxCategory struct {
	Percent string `xml:"Percent"`
}

type LegalMonetaryTotal struct {
	LineExtensionAmount  string `xml:"LineExtensionAmount"`
	AllowanceTotalAmount string `xml:"AllowanceTotalAmount"`
	PayableAmount        string `xml:"PayableAmount"`
}

// Embedded Invoice Structures
type EmbeddedInvoice struct {
	XMLName      xml.Name     `xml:"Invoice"`
	InvoiceLine  InvoiceLine  `xml:"InvoiceLine"`
	PaymentMeans PaymentMeans `xml:"PaymentMeans"`
}

type InvoiceLine struct {
	InvoicedQuantity    string `xml:"InvoicedQuantity"`
	LineExtensionAmount string `xml:"LineExtensionAmount"`
	Item                Item   `xml:"Item"`
	Price               Price  `xml:"Price"`
}

type Item struct {
	Description string `xml:"Description"`
}

type Price struct {
	PriceAmount string `xml:"PriceAmount"`
}

type PaymentMeans struct {
	PaymentDueDate string `xml:"PaymentDueDate"`
}

func (p *xmlProcessor) ProcessXML(xmlData string) (*domain.XMLProcessingResult, error) {
	invoice, err := p.ExtractInvoiceData(xmlData)
	if err != nil {
		return &domain.XMLProcessingResult{
			Error: true,
			Msg:   fmt.Sprintf("Error procesando XML: %v", err),
		}, nil
	}

	return &domain.XMLProcessingResult{
		Data:  invoice,
		Error: false,
	}, nil
}

func (p *xmlProcessor) ExtractInvoiceData(xmlData string) (*domain.XMLInvoice, error) {
	var ublInvoice UBLInvoice
	if err := xml.Unmarshal([]byte(xmlData), &ublInvoice); err != nil {
		return nil, fmt.Errorf("error unmarshaling XML: %w", err)
	}

	// Procesar XML embebido si existe
	embeddedInvoice := p.extractEmbeddedInvoice(ublInvoice.Attachment.ExternalReference.Description)

	// Extraer datos básicos
	invoice := &domain.XMLInvoice{
		ProfileID:             ublInvoice.ProfileID,
		IDFactura:             p.extractInvoiceID(ublInvoice),
		IssueDate:             p.parseDate(ublInvoice.IssueDate),
		PaymentDueDate:        p.extractPaymentDueDate(embeddedInvoice),
		DiscountAmountApplied: p.cleanDecimalValue(ublInvoice.LegalMonetaryTotal.AllowanceTotalAmount),
		InvoicedQuantity:      p.extractInvoicedQuantity(embeddedInvoice),
		TaxableAmount:         p.extractTaxableAmount(ublInvoice, embeddedInvoice),
		PayableAmount:         p.cleanDecimalValue(ublInvoice.LegalMonetaryTotal.PayableAmount),
		Seller:                p.extractSellerInfo(ublInvoice.SenderParty),
		Buyer:                 p.extractBuyerInfo(ublInvoice.ReceiverParty),
		BilledTo:              p.extractBilledTo(ublInvoice.ReceiverParty),
		InvoiceDetails:        p.extractInvoiceDetails(ublInvoice, embeddedInvoice),
	}

	return invoice, nil
}

func (p *xmlProcessor) extractEmbeddedInvoice(description string) *EmbeddedInvoice {
	if description == "" {
		return nil
	}

	// Limpiar y extraer XML embebido
	xmlData := strings.TrimSpace(description)
	var embeddedInvoice EmbeddedInvoice
	if err := xml.Unmarshal([]byte(xmlData), &embeddedInvoice); err != nil {
		return nil
	}

	return &embeddedInvoice
}

func (p *xmlProcessor) extractInvoiceID(invoice UBLInvoice) string {
	if invoice.ID != "" && len(invoice.ID) < 16 {
		return invoice.ID
	}

	// Buscar ParentDocumentID en XML embebido si es necesario
	// Esta lógica puede expandirse según sea necesario
	return invoice.ID
}

func (p *xmlProcessor) extractSellerInfo(party Party) domain.CompanyInfo {
	return domain.CompanyInfo{
		Name:                party.RegistrationName,
		NIT:                 party.CompanyID.Value,
		Tributacion:         party.CompanyID.SchemeID,
		TaxResponsibilities: p.formatTaxResponsibilities(strings.Split(party.TaxScheme.TaxLevelCode, ";")),
	}
}

func (p *xmlProcessor) extractBuyerInfo(party Party) domain.CompanyInfo {
	return domain.CompanyInfo{
		Name:                party.RegistrationName,
		NIT:                 party.CompanyID.Value,
		Tributacion:         party.CompanyID.SchemeID,
		TaxResponsibilities: p.formatTaxResponsibilities(strings.Split(party.TaxScheme.TaxLevelCode, ";")),
	}
}

func (p *xmlProcessor) extractBilledTo(party Party) domain.BilledTo {
	return domain.BilledTo{
		Name:      party.RegistrationName,
		CompanyID: party.CompanyID.Value,
	}
}

func (p *xmlProcessor) extractInvoiceDetails(ublInvoice UBLInvoice, embedded *EmbeddedInvoice) domain.InvoiceDetails {
	var details domain.InvoiceDetails

	if embedded != nil {
		details = domain.InvoiceDetails{
			ProductDescription: embedded.InvoiceLine.Item.Description, // Cambiar Description por ProductDescription
			Quantity:           p.cleanDecimalValue(embedded.InvoiceLine.InvoicedQuantity),
			UnitPrice:          p.calculateUnitPrice(embedded.InvoiceLine),
			TaxPercentageIVA:   p.cleanDecimalValue(ublInvoice.TaxTotal.TaxSubtotal.TaxCategory.Percent),
			Subtotal:           p.cleanDecimalValue(ublInvoice.TaxTotal.TaxSubtotal.TaxableAmount),
			Taxes:              p.cleanDecimalValue(ublInvoice.TaxTotal.TaxAmount),
		}
	}

	return details
}

func (p *xmlProcessor) extractInvoicedQuantity(embedded *EmbeddedInvoice) string {
	if embedded != nil {
		return p.cleanDecimalValue(embedded.InvoiceLine.InvoicedQuantity)
	}
	return "N/A"
}

func (p *xmlProcessor) extractTaxableAmount(ublInvoice UBLInvoice, embedded *EmbeddedInvoice) string {
	if ublInvoice.TaxTotal.TaxSubtotal.TaxableAmount != "" {
		return p.cleanDecimalValue(ublInvoice.TaxTotal.TaxSubtotal.TaxableAmount)
	}
	if embedded != nil {
		return p.cleanDecimalValue(embedded.InvoiceLine.LineExtensionAmount)
	}
	return "N/A"
}

func (p *xmlProcessor) extractPaymentDueDate(embedded *EmbeddedInvoice) *time.Time {
	if embedded != nil && embedded.PaymentMeans.PaymentDueDate != "" {
		return p.parseDate(embedded.PaymentMeans.PaymentDueDate)
	}
	return nil
}

func (p *xmlProcessor) calculateUnitPrice(line InvoiceLine) string {
	if line.LineExtensionAmount != "" && line.InvoicedQuantity != "" {
		total, err1 := strconv.ParseFloat(line.LineExtensionAmount, 64)
		quantity, err2 := strconv.ParseFloat(line.InvoicedQuantity, 64)

		if err1 == nil && err2 == nil && quantity != 0 {
			unitPrice := total / quantity
			return fmt.Sprintf("%.6f", unitPrice)
		}
		return "Error en el cálculo"
	}
	return "Datos incompletos"
}

func (p *xmlProcessor) parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}

	// Intentar diferentes formatos de fecha
	formats := []string{"2006-01-02", "02/01/2006", "2006-01-02T15:04:05"}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t
		}
	}

	return nil
}

func (p *xmlProcessor) cleanDecimalValue(value string) string {
	if value == "" {
		return "0.00"
	}

	// Limitar a 2 decimales
	re := regexp.MustCompile(`(\.\d{2})\d+`)
	cleaned := re.ReplaceAllString(value, "$1")

	// Verificar que sea un número válido
	if _, err := strconv.ParseFloat(cleaned, 64); err != nil {
		return value // Retornar original si no es número
	}

	return cleaned
}

func (p *xmlProcessor) formatTaxResponsibilities(codes []string) string {
	var responsibilities []string
	for _, code := range codes {
		if code == "" {
			continue
		}
		description := p.causacionMap[code]
		if description == "" {
			description = "No definido"
		}
		responsibilities = append(responsibilities, fmt.Sprintf("%s: %s", code, description))
	}

	if len(responsibilities) == 0 {
		return "N/A"
	}

	return strings.Join(responsibilities, "\n")
}

func (p *xmlProcessor) ValidateXMLStructure(xmlData string) error {
	var invoice UBLInvoice
	if err := xml.Unmarshal([]byte(xmlData), &invoice); err != nil {
		return fmt.Errorf("XML inválido: %w", err)
	}

	// Validaciones básicas
	if invoice.ID == "" {
		return fmt.Errorf("ID de factura requerido")
	}

	if invoice.IssueDate == "" {
		return fmt.Errorf("fecha de emisión requerida")
	}

	return nil
}
