package domain

// AuthResponse representa la respuesta de autenticación
type AuthResponse struct {
	OK         bool `json:"ok"`
	StatusCode int  `json:"statusCode"`
	Result     struct {
		Tokens struct {
			Access  string `json:"access"`
			Refresh string `json:"refresh"`
		} `json:"tokens"`
	} `json:"result"`
	Message string `json:"message,omitempty"`
}

// APIResponse representa una respuesta genérica de API
type APIResponse struct {
	OK         bool        `json:"ok"`
	StatusCode int         `json:"statusCode"`
	Result     interface{} `json:"result,omitempty"`
	Message    string      `json:"message,omitempty"`
	Error      string      `json:"error,omitempty"`
	Duplicate  bool        `json:"duplicate,omitempty"`
}

// NITResponse representa la respuesta para NITs
type NITResponse struct {
	OK         bool   `json:"ok"`
	StatusCode int    `json:"statusCode"`
	Result     []NIT  `json:"result"`
	Message    string `json:"message,omitempty"`
}

// NIT representa la estructura de datos de NITs desde la API
type NIT struct {
	ID          int    `json:"id"`
	NIT         string `json:"nit"`
	RazonSocial string `json:"razon_social"`
	Email       string `json:"email"`
	Estado      string `json:"estado"`
}

// InvoiceData representa los datos de factura para enviar a la API
type InvoiceData struct {
	FEVIdFac               string `json:"fev_id_fac"`
	ProductQuantity        string `json:"product_quantity"`
	ProductSubtotal        string `json:"product_subtotal"`
	ProductTaxes           string `json:"product_taxes"`
	ProductIVA             string `json:"product_iva"`
	ProductVlrTotalChecked string `json:"product_vlr_total_checked"`
	ProductVlrTotal        string `json:"product_vlr_total"`
	DiscountGlobalApplied  string `json:"dicount_global_applied"`
	XMLString              string `json:"xml_string"`
	PDFBase64              string `json:"pdf_base64"`
	IssueDate              string `json:"issue_date"`
	PaymentDueDate         string `json:"payment_due_date"`
	SellerNIT              string `json:"seller_nit"`
	BuyerNIT               string `json:"buyer_nit"`
	SocialReasonSeller     string `json:"social_reason_seller"`
	SocialReasonBuyer      string `json:"social_reason_buyer"`
	ProductDescription     string `json:"product_description"`
	ProductVlrUnit         string `json:"product_vlr_unit"`
	NITSeller              string `json:"nit_seller"`
	NITBuyer               string `json:"nit_buyer"`
	Email                  string `json:"email"`
	Estado                 string `json:"estado"`
}
