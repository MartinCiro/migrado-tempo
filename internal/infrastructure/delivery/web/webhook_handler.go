package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"email/internal/core/ports"
)

type WebhookHandler struct {
	webhookService ports.WebhookService
}

func NewWebhookHandler(webhookService ports.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

func (h *WebhookHandler) HandleProcessXML(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req ports.WebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Error decodificando request", err)
		return
	}

	// Validar campos requeridos
	if req.XMLContent == "" || req.XMLFileName == "" || req.Token == "" {
		h.sendErrorResponse(w, "Campos requeridos faltantes: xml_content, xml_file_name, token", nil)
		return
	}

	// Procesar el XML
	result, err := h.webhookService.ProcessXMLFromWebhook(r.Context(), &req)
	if err != nil {
		h.sendErrorResponse(w, "Error procesando XML", err)
		return
	}

	h.sendResponse(w, result)
}

func (h *WebhookHandler) sendResponse(w http.ResponseWriter, result *ports.WebhookResponse) {
	w.Header().Set("Content-Type", "application/json")
	if result.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(result)
}

func (h *WebhookHandler) sendErrorResponse(w http.ResponseWriter, message string, err error) {
	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(ports.WebhookResponse{
		Success: false,
		Error:   errorMsg,
	})
}
