package ports

import (
    "context"
    "email/internal/core/domain"
)

type AIService interface {
    ProcessInvoiceWithAI(ctx context.Context, xmlContent string) (*domain.InvoiceData, error)
    AnalyzeXMLWithAI(ctx context.Context, xmlContent string, context map[string]string) (*domain.AIResponse, error)
    ValidateWithAI(ctx context.Context, invoiceData *domain.InvoiceData, xmlContent string) (bool, error)
}