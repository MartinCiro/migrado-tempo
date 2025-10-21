package ports

import "context"

type WebhookService interface {
	ProcessXMLFromWebhook(ctx context.Context, request *WebhookRequest) (*WebhookResponse, error)
}

type WebhookRequest struct {
	XMLContent  string            `json:"xml_content"`
	XMLFileName string            `json:"xml_file_name"`
	PDFContent  string            `json:"pdf_content,omitempty"`
	PDFFileName string            `json:"pdf_file_name,omitempty"`
	Token       string            `json:"token"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type WebhookResponse struct {
	Success   bool   `json:"success"`
	InvoiceID string `json:"invoice_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}
