package domain

import "time"

// XMLInvoice representa los datos extraídos del XML de factura
type XMLInvoice struct {
	ProfileID             string         `json:"profile_id"`
	IDFactura             string         `json:"id_factura"`
	IssueDate             *time.Time     `json:"issue_date"`
	PaymentDueDate        *time.Time     `json:"payment_due_date"`
	InvoiceDetails        InvoiceDetails `json:"invoice_details"`
	DiscountAmountApplied string         `json:"discount_amount_applied"`
	InvoicedQuantity      string         `json:"invoiced_quantity"`
	TaxableAmount         string         `json:"taxable_amount"`
	PayableAmount         string         `json:"payable_amount"`
	Seller                CompanyInfo    `json:"seller"`
	Buyer                 CompanyInfo    `json:"buyer"`
	BilledTo              BilledTo       `json:"billed_to"`
}

type InvoiceDetails struct {
	Taxes              string `json:"taxes"`
	Subtotal           string `json:"subtotal"`
	Quantity           string `json:"quantity"`
	UnitPrice          string `json:"unit_price"`
	TaxPercentageIVA   string `json:"tax_percentage_iva"`
	ProductDescription string `json:"product_description"`
}

type CompanyInfo struct {
	Name                string `json:"name"`
	NIT                 string `json:"nit"`
	Tributacion         string `json:"tributacion"`
	TaxResponsibilities string `json:"tax_responsibilities"`
}

type BilledTo struct {
	Name      string `json:"name"`
	CompanyID string `json:"company_id"`
}

// PartyInfo representa la información de una empresa (vendedor/comprador)
type PartyInfo struct {
	RegistrationName string   `json:"registration_name"`
	CompanyID        string   `json:"company_id"`
	SchemeID         string   `json:"scheme_id"`
	TaxLevelCode     []string `json:"tax_level_code"`
}

// InvoiceLine representa la línea de detalle de la factura
type InvoiceLine struct {
	Description string `json:"description"`
	Quantity    string `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
}

// XMLProcessingResult resultado del procesamiento del XML
type XMLProcessingResult struct {
	Data  *XMLInvoice `json:"data"`
	Error bool        `json:"error"`
	Msg   string      `json:"msg,omitempty"`
}
