package services

import (
    "context"
    "email/internal/core/domain"
    "email/internal/core/ports"
    "fmt"
)

type aiService struct {
    aiClient ports.AIService
}

func NewAIService(aiClient ports.AIService) ports.AIService {
    return &aiService{
        aiClient: aiClient,
    }
}

func (s *aiService) ProcessInvoiceWithAI(ctx context.Context, xmlContent string) (*domain.InvoiceData, error) {
    fmt.Println("🤖 Processing invoice with Gemini AI...")
    
    // Primero intentar análisis completo
    response, err := s.aiClient.AnalyzeXMLWithAI(ctx, xmlContent, map[string]string{
        "task": "invoice_processing",
    })
    
    if err != nil {
        return nil, fmt.Errorf("AI processing failed: %w", err)
    }
    
    if response.IsProcessed && response.ExtractedData != nil {
        fmt.Printf("✅ AI successfully processed invoice: %s\n", response.ExtractedData.FEVIdFac)
        return response.ExtractedData, nil
    }
    
    return nil, fmt.Errorf("AI could not process invoice: %s", response.Error)
}

func (s *aiService) AnalyzeXMLWithAI(ctx context.Context, xmlContent string, context map[string]string) (*domain.AIResponse, error) {
    return s.aiClient.AnalyzeXMLWithAI(ctx, xmlContent, context)
}

func (s *aiService) ValidateWithAI(ctx context.Context, invoiceData *domain.InvoiceData, xmlContent string) (bool, error) {
    return s.aiClient.ValidateWithAI(ctx, invoiceData, xmlContent)
}