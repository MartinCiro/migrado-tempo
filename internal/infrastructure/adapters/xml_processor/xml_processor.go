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
type AttachedDocument struct {
	XMLName          xml.Name       `xml:"AttachedDocument"`
	ProfileID        string         `xml:"ProfileID"`
	ID               string         `xml:"ID"`
	IssueDate        string         `xml:"IssueDate"`
	ParentDocumentID string         `xml:"ParentDocumentID"`
	SenderParty      Party          `xml:"SenderParty"`
	ReceiverParty    Party          `xml:"ReceiverParty"`
	Attachment       Attachment     `xml:"Attachment"`
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
	ID   string `xml:"ID"`
	Name string `xml:"Name"`
}

type Attachment struct {
	ExternalReference ExternalReference `xml:"ExternalReference"`
}

type ExternalReference struct {
	Description string `xml:"Description"`
}

// Para la factura embebida
type UBLInvoice struct {
	XMLName                 xml.Name           `xml:"Invoice"`
	ProfileID               string             `xml:"ProfileID"`
	ID                      string             `xml:"ID"`
	IssueDate               string             `xml:"IssueDate"`
	DueDate                 string             `xml:"DueDate"`
	InvoiceTypeCode         string             `xml:"InvoiceTypeCode"`
	AccountingSupplierParty AccountingParty    `xml:"AccountingSupplierParty"`
	AccountingCustomerParty AccountingParty    `xml:"AccountingCustomerParty"`
	LegalMonetaryTotal      LegalMonetaryTotal `xml:"LegalMonetaryTotal"`
	InvoiceLine             InvoiceLine        `xml:"InvoiceLine"`
	PaymentMeans            PaymentMeans       `xml:"PaymentMeans"`
	TaxTotal                TaxTotal           `xml:"TaxTotal"`
}

type PartyName struct {
	Name string `xml:"Name"`
}

type PartyTaxScheme struct {
	RegistrationName string    `xml:"RegistrationName"`
	CompanyID        CompanyID `xml:"CompanyID"`
	TaxScheme        TaxScheme `xml:"TaxScheme"`
}

type PartyLegalEntity struct {
	RegistrationName string    `xml:"RegistrationName"`
	CompanyID        CompanyID `xml:"CompanyID"`
}

type AccountingPartyDetails struct {
	PartyName        PartyName        `xml:"PartyName"`
	PartyTaxScheme   PartyTaxScheme   `xml:"PartyTaxScheme"`
	PartyLegalEntity PartyLegalEntity `xml:"PartyLegalEntity"`
}


type AccountingParty struct {
	Party AccountingPartyDetails `xml:"Party"`
}

type LegalMonetaryTotal struct {
	LineExtensionAmount  string `xml:"LineExtensionAmount"`
	AllowanceTotalAmount string `xml:"AllowanceTotalAmount"`
	PayableAmount        string `xml:"PayableAmount"`
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

type InvoiceLine struct {
	ID                 string `xml:"ID"`
	InvoicedQuantity   string `xml:"InvoicedQuantity"`
	LineExtensionAmount string `xml:"LineExtensionAmount"`
	Item               Item   `xml:"Item"`
	Price              Price  `xml:"Price"`
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
	// Primero parsear el AttachedDocument
	var attachedDoc AttachedDocument
	if err := xml.Unmarshal([]byte(xmlData), &attachedDoc); err != nil {
		return nil, fmt.Errorf("error unmarshaling AttachedDocument: %w", err)
	}

	// Extraer y parsear el XML embebido
	embeddedXML := p.extractEmbeddedXML(attachedDoc.Attachment.ExternalReference.Description)
	if embeddedXML == "" {
		return nil, fmt.Errorf("no embedded XML found")
	}

	// Parsear la factura UBL embebida
	var ublInvoice UBLInvoice
	if err := xml.Unmarshal([]byte(embeddedXML), &ublInvoice); err != nil {
		fmt.Printf("Error unmarshaling embedded UBL invoice: %v\n", err)
		return nil, fmt.Errorf("error unmarshaling embedded UBL invoice: %w", err)
	}

	// Procesar datos combinados
	return p.processInvoiceData(attachedDoc, ublInvoice)
}

func (p *xmlProcessor) extractEmbeddedXML(description string) string {
	if description == "" {
		return ""
	}

	xmlData := strings.TrimSpace(description)

	// Remover etiquetas CDATA si existen
	if strings.HasPrefix(xmlData, "<![CDATA[") && strings.HasSuffix(xmlData, "]]>") {
		xmlData = strings.TrimPrefix(xmlData, "<![CDATA[")
		xmlData = strings.TrimSuffix(xmlData, "]]>")
		xmlData = strings.TrimSpace(xmlData)
	}

	return xmlData
}

func (p *xmlProcessor) processInvoiceData(attachedDoc AttachedDocument, ublInvoice UBLInvoice) (*domain.XMLInvoice, error) {
	// Construir el objeto de dominio
	invoice := &domain.XMLInvoice{
		ProfileID:             attachedDoc.ProfileID,
		IDFactura:             p.extractInvoiceID(attachedDoc, ublInvoice),
		IssueDate:             p.parseDate(ublInvoice.IssueDate),
		PaymentDueDate:        p.extractPaymentDueDate(ublInvoice),
		DiscountAmountApplied: p.cleanDecimalValue(ublInvoice.LegalMonetaryTotal.AllowanceTotalAmount),
		InvoicedQuantity:      p.extractInvoicedQuantity(ublInvoice),
		TaxableAmount:         p.extractTaxableAmount(ublInvoice),
		PayableAmount:         p.cleanDecimalValue(ublInvoice.LegalMonetaryTotal.PayableAmount),
		Seller:                p.extractSellerInfo(attachedDoc.SenderParty, ublInvoice.AccountingSupplierParty.Party),
		Buyer:                 p.extractBuyerInfo(attachedDoc.ReceiverParty, ublInvoice.AccountingCustomerParty.Party),
		BilledTo:              p.extractBilledTo(attachedDoc.ReceiverParty, ublInvoice.AccountingCustomerParty.Party),
		InvoiceDetails:        p.extractInvoiceDetails(ublInvoice),
	}

	return invoice, nil
}

func (p *xmlProcessor) extractInvoiceID(attachedDoc AttachedDocument, ublInvoice UBLInvoice) string {
	// Priorizar ParentDocumentID del documento adjunto
	if attachedDoc.ParentDocumentID != "" {
		return attachedDoc.ParentDocumentID
	}

	// Si no hay ParentDocumentID, usar el ID de la factura UBL
	if ublInvoice.ID != "" {
		return ublInvoice.ID
	}

	return attachedDoc.ID
}

func (p *xmlProcessor) extractSellerInfo(attachedParty Party, ublParty AccountingPartyDetails) domain.CompanyInfo {
	// Priorizar información del UBL Invoice
	var name, nit, schemeID string
	
	// Intentar obtener de PartyLegalEntity primero
	if ublParty.PartyLegalEntity.RegistrationName != "" {
		name = ublParty.PartyLegalEntity.RegistrationName
		nit = ublParty.PartyLegalEntity.CompanyID.Value
		schemeID = ublParty.PartyLegalEntity.CompanyID.SchemeID
	}
	
	// Si no hay datos en PartyLegalEntity, intentar con PartyTaxScheme
	if name == "" && ublParty.PartyTaxScheme.RegistrationName != "" {
		name = ublParty.PartyTaxScheme.RegistrationName
		nit = ublParty.PartyTaxScheme.CompanyID.Value
		schemeID = ublParty.PartyTaxScheme.CompanyID.SchemeID
	}
	
	// Si no hay datos en UBL, usar attachedParty
	if name == "" {
		name = attachedParty.RegistrationName
		nit = attachedParty.CompanyID.Value
		schemeID = attachedParty.CompanyID.SchemeID
	}

	return domain.CompanyInfo{
		Name:                name,
		NIT:                 nit,
		Tributacion:         schemeID,
		TaxResponsibilities: p.formatTaxResponsibilities([]string{ublParty.PartyTaxScheme.TaxScheme.Name}),
	}
}

func (p *xmlProcessor) extractBuyerInfo(attachedParty Party, ublParty AccountingPartyDetails) domain.CompanyInfo {
	// Priorizar información del UBL Invoice
	var name, nit, schemeID string
	
	// Intentar obtener de PartyLegalEntity primero
	if ublParty.PartyLegalEntity.RegistrationName != "" {
		name = ublParty.PartyLegalEntity.RegistrationName
		nit = ublParty.PartyLegalEntity.CompanyID.Value
		schemeID = ublParty.PartyLegalEntity.CompanyID.SchemeID
	}
	
	// Si no hay datos en PartyLegalEntity, intentar con PartyTaxScheme
	if name == "" && ublParty.PartyTaxScheme.RegistrationName != "" {
		name = ublParty.PartyTaxScheme.RegistrationName
		nit = ublParty.PartyTaxScheme.CompanyID.Value
		schemeID = ublParty.PartyTaxScheme.CompanyID.SchemeID
	}
	
	// Si no hay datos en UBL, usar attachedParty
	if name == "" {
		name = attachedParty.RegistrationName
		nit = attachedParty.CompanyID.Value
		schemeID = attachedParty.CompanyID.SchemeID
	}


	return domain.CompanyInfo{
		Name:                name,
		NIT:                 nit,
		Tributacion:         schemeID,
		TaxResponsibilities: p.formatTaxResponsibilities([]string{ublParty.PartyTaxScheme.TaxScheme.Name}),
	}
}

func (p *xmlProcessor) extractBilledTo(attachedParty Party, ublParty AccountingPartyDetails) domain.BilledTo {
	// Priorizar información del UBL Invoice
	var name, companyID string
	
	// Intentar obtener de PartyLegalEntity primero
	if ublParty.PartyLegalEntity.RegistrationName != "" {
		name = ublParty.PartyLegalEntity.RegistrationName
		companyID = ublParty.PartyLegalEntity.CompanyID.Value
	}
	
	// Si no hay datos en PartyLegalEntity, intentar con PartyTaxScheme
	if name == "" && ublParty.PartyTaxScheme.RegistrationName != "" {
		name = ublParty.PartyTaxScheme.RegistrationName
		companyID = ublParty.PartyTaxScheme.CompanyID.Value
	}
	
	// Si no hay datos en UBL, usar attachedParty
	if name == "" {
		name = attachedParty.RegistrationName
		companyID = attachedParty.CompanyID.Value
	}

	return domain.BilledTo{
		Name:      name,
		CompanyID: companyID,
	}
}

func (p *xmlProcessor) extractInvoiceDetails(ublInvoice UBLInvoice) domain.InvoiceDetails {
    details := domain.InvoiceDetails{
        ProductDescription: ublInvoice.InvoiceLine.Item.Description,
        Quantity:           p.cleanDecimalValue(ublInvoice.InvoiceLine.InvoicedQuantity),
        UnitPrice:          p.calculateUnitPrice(ublInvoice.InvoiceLine),
        Subtotal:           p.cleanDecimalValue(ublInvoice.LegalMonetaryTotal.LineExtensionAmount),
        TaxPercentageIVA:   p.cleanDecimalValue(ublInvoice.TaxTotal.TaxSubtotal.TaxCategory.Percent), // Esto capturará "0.00"
        Taxes:              p.cleanDecimalValue(ublInvoice.TaxTotal.TaxAmount), // Esto capturará "0.00"
    }

    return details
}

func (p *xmlProcessor) extractInvoicedQuantity(ublInvoice UBLInvoice) string {
	if ublInvoice.InvoiceLine.InvoicedQuantity != "" {
		return p.cleanDecimalValue(ublInvoice.InvoiceLine.InvoicedQuantity)
	}
	return "N/A"
}

func (p *xmlProcessor) extractTaxableAmount(ublInvoice UBLInvoice) string {
    // Priorizar LineExtensionAmount sobre TaxableAmount del TaxSubtotal
    if ublInvoice.LegalMonetaryTotal.LineExtensionAmount != "" {
        return p.cleanDecimalValue(ublInvoice.LegalMonetaryTotal.LineExtensionAmount)
    }
    if ublInvoice.TaxTotal.TaxSubtotal.TaxableAmount != "" {
        return p.cleanDecimalValue(ublInvoice.TaxTotal.TaxSubtotal.TaxableAmount)
    }
    return "N/A"
}

func (p *xmlProcessor) extractPaymentDueDate(ublInvoice UBLInvoice) *time.Time {
	// Priorizar DueDate de la factura UBL
	if ublInvoice.DueDate != "" {
		return p.parseDate(ublInvoice.DueDate)
	}
	// Luego PaymentDueDate de PaymentMeans
	if ublInvoice.PaymentMeans.PaymentDueDate != "" {
		return p.parseDate(ublInvoice.PaymentMeans.PaymentDueDate)
	}
	return nil
}

func (p *xmlProcessor) calculateUnitPrice(line InvoiceLine) string {
	if line.LineExtensionAmount != "" && line.InvoicedQuantity != "" {
		total, err1 := strconv.ParseFloat(p.cleanDecimalValue(line.LineExtensionAmount), 64)
		quantity, err2 := strconv.ParseFloat(p.cleanDecimalValue(line.InvoicedQuantity), 64)

		if err1 == nil && err2 == nil && quantity != 0 {
			unitPrice := total / quantity
			return fmt.Sprintf("%.2f", unitPrice)
		}
		return "Error en el cálculo"
	}

	// Si no se puede calcular, usar el precio directo si está disponible
	if line.Price.PriceAmount != "" {
		return p.cleanDecimalValue(line.Price.PriceAmount)
	}

	return "Datos incompletos"
}

func (p *xmlProcessor) parseDate(dateStr string) *time.Time {
	if dateStr == "" {
		return nil
	}

	// Intentar diferentes formatos de fecha
	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z",
	}

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

	// Remover caracteres no numéricos excepto punto y signo negativo
	re := regexp.MustCompile(`[^\d.-]`)
	cleaned := re.ReplaceAllString(value, "")

	// Limitar a 2 decimales
	re = regexp.MustCompile(`(\.\d{2})\d+`)
	cleaned = re.ReplaceAllString(cleaned, "$1")

	// Verificar que sea un número válido
	if _, err := strconv.ParseFloat(cleaned, 64); err != nil {
		return "0.00" // Retornar cero si no es número válido
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
	// Primero validar como AttachedDocument
	var attachedDoc AttachedDocument
	if err := xml.Unmarshal([]byte(xmlData), &attachedDoc); err != nil {
		return fmt.Errorf("XML inválido: %w", err)
	}

	// Validaciones básicas del documento adjunto
	if attachedDoc.ParentDocumentID == "" {
		return fmt.Errorf("ParentDocumentID requerido")
	}

	if attachedDoc.IssueDate == "" {
		return fmt.Errorf("fecha de emisión requerida")
	}

	// Validar que tenga attachment con XML embebido
	if attachedDoc.Attachment.ExternalReference.Description == "" {
		return fmt.Errorf("XML embebido requerido en el attachment")
	}

	return nil
}