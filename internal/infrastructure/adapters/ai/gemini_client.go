package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"email/internal/core/domain"
	
	// 1. Importar el SDK oficial de Google GenAI
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// 2. Modificar la estructura para incluir el cliente de Gemini del SDK
type geminiClient struct {
	client    *genai.Client // Cliente del SDK de Gemini
	modelName string
	promptDir string
}

// 3. Actualizar el constructor para inicializar el cliente de Gemini
func NewGeminiClient(apiKey, promptDir string) (*geminiClient, error) {
	// Crea el cliente de Gemini usando la API Key
	client, err := genai.NewClient(context.Background(), option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error al crear el cliente de Gemini: %w", err)
	}

	return &geminiClient{
		client:    client,
		modelName: "gemini-2.5-flash", // Modelo recomendado para tareas de extracción
		promptDir: promptDir,
	}, nil
}

func (c *geminiClient) ProcessInvoiceWithAI(ctx context.Context, xmlContent string) (*domain.InvoiceData, error) {
	response, err := c.AnalyzeXMLWithAI(ctx, xmlContent, map[string]string{
		"task": "invoice_processing",
	})

	if err != nil {
		return nil, err
	}

	return response.ExtractedData, nil
}

func (c *geminiClient) AnalyzeXMLWithAI(ctx context.Context, xmlContent string, context map[string]string) (*domain.AIResponse, error) {
	// Cargar prompt específico
	prompt, err := c.loadPrompt(domain.PromptTypeInvoiceProcessing)
	if err != nil {
		return nil, err
	}

	// Construir mensaje para Gemini
	message := c.buildMessage(prompt, xmlContent, context)

	// Llamar a Gemini API (implementado con el SDK)
	aiResponse, err := c.callGeminiAPI(ctx, message)
	if err != nil {
		return nil, err
	}

	// Parsear respuesta
	return c.parseAIResponse(aiResponse)
}

func (c *geminiClient) ValidateWithAI(ctx context.Context, invoiceData *domain.InvoiceData, xmlContent string) (bool, error) {
	// Implementar validación con AI
	return true, nil
}

func (c *geminiClient) loadPrompt(promptType string) (string, error) {
	// NOTE: Se asume que domain.PromptTypeInvoiceProcessing es la clave para el archivo
	promptFile := fmt.Sprintf("%s/%s.txt", c.promptDir, promptType)
	content, err := os.ReadFile(promptFile)
	if err != nil {
		return "", fmt.Errorf("error loading prompt: %w", err)
	}
	return string(content), nil
}

func (c *geminiClient) buildMessage(prompt, xmlContent string, context map[string]string) string {
	// La instrucción clave para obtener JSON estructurado se coloca en el mensaje
	return fmt.Sprintf(`%s

**Contenido XML a analizar:**
%s

**Contexto Adicional:**
%v

Instrucción: Analiza el contenido XML de la factura y extrae todos los campos relevantes en formato JSON. Asegúrate de que el resultado sea SOLO el objeto JSON, sin texto explicativo o formato markdown como \`\`\`.`, prompt, xmlContent, context)
}

// 4. Implementar la llamada real a la API de Gemini usando el SDK
func (c *geminiClient) callGeminiAPI(ctx context.Context, message string) (string, error) {
	fmt.Println("🔮 Llamando a la API de Gemini...")

	// Configurar la solicitud
	resp, err := c.client.GenerativeModel(c.modelName).GenerateContent(
		ctx,
		genai.Text(message),
		// Se usa un modo de respuesta JSON para garantizar el formato de salida
		genai.WithResponseMIMEType("application/json"),
	)
	
	if err != nil {
		return "", fmt.Errorf("error en la llamada a la API de Gemini: %w", err)
	}

	// Extraer el texto de la respuesta
	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("la respuesta de Gemini no contiene contenido")
	}

	// La respuesta suele estar en la primera parte del primer candidato
	generatedText := fmt.Sprint(resp.Candidates[0].Content.Parts[0])

	// El modo JSON de Gemini ya debería garantizar una salida limpia,
	// pero se pueden añadir pasos de limpieza si es necesario (ej: eliminar \`\`\`json)
	return strings.TrimSpace(generatedText), nil
}

func (c *geminiClient) parseAIResponse(aiResponse string) (*domain.AIResponse, error) {
	var invoiceData domain.InvoiceData
	// La respuesta de la API es un string que debería ser un JSON limpio
	if err := json.Unmarshal([]byte(aiResponse), &invoiceData); err != nil {
		return &domain.AIResponse{
			Content:     aiResponse,
			IsProcessed: false,
			Error:       fmt.Sprintf("Fallo al parsear la respuesta JSON: %v. Respuesta AI: %s", err, aiResponse),
		}, nil
	}

	return &domain.AIResponse{
		Content:     aiResponse,
		IsProcessed: true,
		Confidence:  0.85, // Nota: El SDK no proporciona un campo de confianza directo, esto es un valor de ejemplo.
		ExtractedData: &invoiceData,
	}, nil
}