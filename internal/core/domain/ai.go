package domain

// AIPromptRequest solicitud para procesamiento con AI
type AIPromptRequest struct {
    XMLContent  string
    PromptType  string
    Context     map[string]string
}

// AIResponse respuesta de Gemini
type AIResponse struct {
    Content     string
    IsProcessed bool
    Confidence  float64
    ExtractedData *InvoiceData
    Error       string
}

// PromptType tipos de prompts disponibles
const (
    PromptTypeInvoiceProcessing = "invoice_processing"
    PromptTypeXMLAnalysis       = "xml_analysis"
)