package gmailapi

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"

	"email/internal/core/domain"
	"email/internal/core/ports"
)

type gmailClient struct {
	service *gmail.Service
}

func NewGmailClient(service *gmail.Service) ports.EmailRepository {
	return &gmailClient{
		service: service,
	}
}

// Connect implementa la conexión al servicio Gmail
func (c *gmailClient) Connect() error {
	// Verificar que el servicio esté funcionando
	_, err := c.service.Users.GetProfile("me").Do()
	if err != nil {
		return fmt.Errorf("failed to connect to Gmail: %w", err)
	}
	return nil
}

func (c *gmailClient) Disconnect() error {
	// Gmail API no requiere desconexión explícita
	return nil
}

func (c *gmailClient) SearchEmails(criteria domain.EmailCriteria) ([]domain.Email, error) {
    filter := domain.EmailFilter{
        From:    criteria.From,
        Subject: criteria.Subject,
        Since:   criteria.Since,
        Unread:  criteria.Unread, 
		Label:   criteria.Label,
    }
    return c.GetEmails(context.Background(), filter)
}

// GetEmailsByLabel obtiene emails de una etiqueta específica
// GetEmailsByLabel obtiene emails de una etiqueta específica con paginación
func (c *gmailClient) GetEmailsByLabel(ctx context.Context, labelName string) ([]domain.Email, error) {
	// Obtener ID de la etiqueta
	labelID, err := c.GetLabelID(ctx, labelName)
	if err != nil {
		return nil, fmt.Errorf("error getting label ID: %w", err)
	}

	var allEmails []domain.Email
	var pageToken string
	pageCount := 0

	for {
		pageCount++
		// Buscar mensajes en la etiqueta con paginación
		call := c.service.Users.Messages.List("me").LabelIds(labelID).MaxResults(500)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve messages from label: %v", err)
		}

		//fmt.Printf("📄 Page %d: %d messages, next page: %t\n", pageCount, len(response.Messages), response.NextPageToken != "")

		// Procesar mensajes de esta página
		for _, msg := range response.Messages {
			email, err := c.convertGmailMessageWithDetails(msg)
			if err != nil {
				log.Printf("Error converting message %s: %v", msg.Id, err)
				continue
			}
			allEmails = append(allEmails, *email)
		}

		// Verificar si hay más páginas
		if response.NextPageToken == "" {
			break
		}
		pageToken = response.NextPageToken

		// Pequeña pausa para no saturar la API
		time.Sleep(100 * time.Millisecond)
	}

	//fmt.Printf("📨 Total emails retrieved from label '%s': %d\n", labelName, len(allEmails))
	return allEmails, nil
}

// convertGmailMessageWithDetails obtiene detalles completos del mensaje
func (c *gmailClient) convertGmailMessageWithDetails(gmailMsg *gmail.Message) (*domain.Email, error) {
	// Obtener mensaje completo con adjuntos
	fullMessage, err := c.GetMessageDetails(gmailMsg.Id)
	if err != nil {
		return nil, fmt.Errorf("error getting full message: %w", err)
	}

	email := &domain.Email{
		ID:           fullMessage.Id,
		InternalDate: fullMessage.InternalDate,
	}

	// Extraer headers
	for _, header := range fullMessage.Payload.Headers {
		switch strings.ToLower(header.Name) {
		case "from":
			email.From = header.Value
		case "subject":
			email.Subject = header.Value
		case "date":
			if parsedDate, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
				email.ReceivedAt = parsedDate
			}
		}
	}

	// Extraer cuerpo y adjuntos
	if err := c.extractContent(fullMessage.Payload, email); err != nil {
		return nil, fmt.Errorf("failed to extract content: %w", err)
	}

	return email, nil
}

// GetEmails obtiene emails basado en el filtro
// GetEmails obtiene emails basado en el filtro (SOLO METADATOS)
func (c *gmailClient) GetEmails(ctx context.Context, filter domain.EmailFilter) ([]domain.Email, error) {
    query := c.buildSearchQuery(filter)
    
    fmt.Printf("📨 Gmail API Query: '%s'\n", query)
    
    // ✅ SOLO obtener metadatos, no contenido completo
    call := c.service.Users.Messages.List("me").Q(query)
    response, err := call.Do()
    if err != nil {
        return nil, fmt.Errorf("unable to retrieve messages: %w", err)
    }

    if response == nil {
        return []domain.Email{}, nil
    }

    var emails []domain.Email
    for _, msg := range response.Messages {
        if msg == nil {
            continue
        }

        // ✅ Obtener solo metadatos básicos (sin adjuntos)
        basicMsg, err := c.service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
        if err != nil {
            log.Printf("❌ Error getting message metadata %s: %v\n", msg.Id, err)
            continue
        }

        email, err := c.convertToBasicEmail(basicMsg)
        if err != nil {
            log.Printf("❌ Error converting message %s: %v\n", msg.Id, err)
            continue
        }

        if c.matchesFilter(email, filter) {
            emails = append(emails, *email)
        }
    }

    fmt.Printf("📨 Found %d emails with filter Unread=%t, Label='%s'\n", 
        len(emails), filter.Unread, filter.Label)
    return emails, nil
}

// convertToBasicEmail convierte solo metadatos básicos (sin adjuntos)
func (c *gmailClient) convertToBasicEmail(gmailMsg *gmail.Message) (*domain.Email, error) {
    if gmailMsg == nil {
        return nil, fmt.Errorf("gmail message is nil")
    }

    email := &domain.Email{
        ID:           gmailMsg.Id,
        InternalDate: gmailMsg.InternalDate,
    }

    if gmailMsg.Payload == nil || gmailMsg.Payload.Headers == nil {
        return email, nil
    }

    // Solo extraer headers básicos
    for _, header := range gmailMsg.Payload.Headers {
        switch strings.ToLower(header.Name) {
        case "from":
            email.From = header.Value
        case "subject":
            email.Subject = header.Value
        case "date":
            if parsedDate, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
                email.ReceivedAt = parsedDate
            }
        }
    }

    return email, nil
}

func (c *gmailClient) buildSearchQuery(filter domain.EmailFilter) string {
    var queryParts []string
    
    //fmt.Printf("🔍 Building query - Unread: %t, Label: '%s'\n", filter.Unread, filter.Label)
    
    // ✅ CORRECCIÓN: Siempre aplicar "is:unread" cuando filter.Unread es true
    if filter.Unread {
        queryParts = append(queryParts, "is:unread")
    }
    
    // Si se especifica un label, buscar en esa etiqueta
    if filter.Label != "" {
        queryParts = append(queryParts, fmt.Sprintf("label:%s", filter.Label))
    }
    
    // Filtros adicionales
    if filter.From != "" {
        queryParts = append(queryParts, fmt.Sprintf("from:%s", filter.From))
    }
    if filter.Subject != "" {
        queryParts = append(queryParts, fmt.Sprintf("subject:%s", filter.Subject))
    }
    if !filter.Since.IsZero() {
        queryParts = append(queryParts, fmt.Sprintf("after:%d", filter.Since.Unix()))
    }
    
    finalQuery := strings.Join(queryParts, " ")
    //fmt.Printf("📨 Final Gmail API Query: '%s'\n", finalQuery)
    
    return finalQuery
}

// GetLabelID obtiene el ID de una etiqueta por su nombre
func (c *gmailClient) GetLabelID(ctx context.Context, labelName string) (string, error) {
	labels, err := c.service.Users.Labels.List("me").Do()
	if err != nil {
		return "", fmt.Errorf("error listing labels: %w", err)
	}

	for _, label := range labels.Labels {
		if label.Name == labelName {
			return label.Id, nil
		}
	}

	return "", fmt.Errorf("label not found: %s", labelName)
}

// MoveToLabel mueve un email a una etiqueta específica
func (c *gmailClient) MoveToLabel(ctx context.Context, emailID, labelID string) error {
	_, err := c.service.Users.Messages.Modify("me", emailID, &gmail.ModifyMessageRequest{
		AddLabelIds: []string{labelID},
	}).Do()
	return err
}

/* // ListUnreadMessages - Tu implementación existente
// ListUnreadMessages - Método auxiliar específico para no leídos
func (c *gmailClient) ListUnreadMessages() ([]*gmail.Message, error) {
    call := c.service.Users.Messages.List("me").Q("is:unread")
    response, err := call.Do()
    if err != nil {
        return nil, fmt.Errorf("unable to retrieve messages: %v", err)
    }

    var messages []*gmail.Message
    for _, msg := range response.Messages {
        message, err := c.service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
        if err != nil {
            log.Printf("Error getting message %s: %v", msg.Id, err)
            continue
        }
        messages = append(messages, message)
    }

    return messages, nil
} */

// GetMessageDetails - Tu implementación existente
func (c *gmailClient) GetMessageDetails(messageId string) (*gmail.Message, error) {
	return c.service.Users.Messages.Get("me", messageId).Format("full").Do()
}

// SaveEmail - Implementación requerida por el puerto
func (c *gmailClient) SaveEmail(ctx context.Context, email *domain.Email) error {
	// Gmail API es principalmente de lectura para este caso
	return nil
}

// DeleteEmail - Marcar como leído después de procesar
func (c *gmailClient) DeleteEmail(ctx context.Context, emailID string) error {
	// Marcar como leído después de procesar
	_, err := c.service.Users.Messages.Modify("me", emailID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}).Do()
	return err
}

// convertGmailMessage convierte un mensaje de Gmail a nuestro dominio
func (c *gmailClient) convertGmailMessage(gmailMsg *gmail.Message) (*domain.Email, error) {
    // ✅ VALIDACIÓN: Verificar que el mensaje no sea nil
    if gmailMsg == nil {
        return nil, fmt.Errorf("gmail message is nil")
    }

    email := &domain.Email{
        ID:           gmailMsg.Id,
        InternalDate: gmailMsg.InternalDate,
        RawData:      []byte{},
    }

    // ✅ VALIDACIÓN: Verificar que Payload no sea nil
    if gmailMsg.Payload == nil {
        return email, nil // Retornar email básico sin headers
    }

    // Extraer información de los headers
    if gmailMsg.Payload.Headers != nil {
        for _, header := range gmailMsg.Payload.Headers {
            switch strings.ToLower(header.Name) {
            case "from":
                email.From = header.Value
            case "subject":
                email.Subject = header.Value
            case "date":
                if parsedDate, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
                    email.ReceivedAt = parsedDate
                }
            }
        }
    }

    // Si necesitamos detalles completos (cuerpo y adjuntos)
    if len(gmailMsg.Payload.Parts) == 0 {
        // Obtener mensaje completo
        fullMessage, err := c.GetMessageDetails(gmailMsg.Id)
        if err != nil {
            return nil, fmt.Errorf("error getting full message: %w", err)
        }
        return c.extractFullMessageDetails(fullMessage)
    }

    return email, nil
}

// extractFullMessageDetails extrae cuerpo y adjuntos del mensaje completo
func (c *gmailClient) extractFullMessageDetails(gmailMsg *gmail.Message) (*domain.Email, error) {
	email := &domain.Email{
		ID:           gmailMsg.Id,
		InternalDate: gmailMsg.InternalDate,
	}

	// Extraer headers
	for _, header := range gmailMsg.Payload.Headers {
		switch strings.ToLower(header.Name) {
		case "from":
			email.From = header.Value
		case "subject":
			email.Subject = header.Value
		case "date":
			if parsedDate, err := time.Parse(time.RFC1123Z, header.Value); err == nil {
				email.ReceivedAt = parsedDate
			}
		}
	}

	// Extraer cuerpo y adjuntos
	if err := c.extractContent(gmailMsg.Payload, email); err != nil {
		return nil, fmt.Errorf("failed to extract content: %w", err)
	}

	return email, nil
}

// extractContent extrae el contenido del mensaje (cuerpo y adjuntos) de forma recursiva
func (c *gmailClient) extractContent(part *gmail.MessagePart, email *domain.Email) error {
	if part == nil {
		return nil
	}

	// ✅ SOLO procesar archivos ZIP
	if part.Filename != "" && strings.HasSuffix(strings.ToLower(part.Filename), ".zip") {
		//fmt.Printf("  📎 Found ZIP file: %s (MIME: %s)\n", part.Filename, part.MimeType)

		var attachmentData []byte
		var err error

		// Caso 1: Datos directos en el body
		if part.Body != nil && part.Body.Data != "" {
			attachmentData, err = base64.URLEncoding.DecodeString(part.Body.Data)
			if err != nil {
				fmt.Printf("  ❌ Error decoding ZIP data: %v\n", err)
			} else {
				fmt.Printf("  💾 ZIP data from body: %d bytes\n", len(attachmentData))
			}
		}

		// Caso 2: Attachment ID (necesita descarga separada)
		if len(attachmentData) == 0 && part.Body != nil && part.Body.AttachmentId != "" {
			//fmt.Printf("  🔗 Downloading ZIP attachment with ID: %s\n", part.Body.AttachmentId)
			attachment, err := c.service.Users.Messages.Attachments.Get("me", email.ID, part.Body.AttachmentId).Do()
			if err != nil {
				fmt.Printf("  ❌ Error downloading ZIP: %v\n", err)
			} else {
				attachmentData, err = base64.URLEncoding.DecodeString(attachment.Data)
				if err != nil {
					fmt.Printf("  ❌ Error decoding downloaded ZIP: %v\n", err)
				} else {
					fmt.Printf("  💾 Downloaded ZIP: %d bytes\n", len(attachmentData))
				}
			}
		}

		// Solo agregar si tenemos datos
		if len(attachmentData) > 0 {
			email.Attachments = append(email.Attachments, domain.Attachment{
				Filename: part.Filename,
				Content:  attachmentData,
				MIMEType: part.MimeType,
			})
			//fmt.Printf("  ✅ Added ZIP attachment: %s (%d bytes)\n", part.Filename, len(attachmentData))
		}
	}

	// ❌ ELIMINAR: No procesar otros tipos de archivos (JPG, PDF, etc.)

	// Procesar partes hijas recursivamente SOLO para seguir buscando ZIPs
	if part.Parts != nil {
		for _, subpart := range part.Parts {
			//fmt.Printf("  🔄 Processing subpart %d/%d\n", i+1, len(part.Parts))
			if err := c.extractContent(subpart, email); err != nil {
				fmt.Printf("  ❌ Error processing subpart: %v\n", err)
			}
		}
	}

	return nil
}

// FindZipAttachments busca específicamente adjuntos ZIP en un mensaje
func (c *gmailClient) FindZipAttachments(ctx context.Context, messageID string) ([][]byte, []string, error) {
	//fmt.Printf("🔍 Searching for ZIP attachments in message: %s\n", messageID)

	// Obtener mensaje completo
	fullMessage, err := c.GetMessageDetails(messageID)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting message details: %w", err)
	}

	var zipContents [][]byte
	var zipFilenames []string

	// Función recursiva para buscar ZIPs
	var searchForZips func(part *gmail.MessagePart)
	searchForZips = func(part *gmail.MessagePart) {
		if part == nil {
			return
		}

		// Verificar si es un archivo ZIP
		if part.Filename != "" && strings.HasSuffix(strings.ToLower(part.Filename), ".zip") {
			//fmt.Printf("📎 Found ZIP file: %s\n", part.Filename)

			var fileData []byte

			// Intentar obtener datos del body
			if part.Body != nil && part.Body.Data != "" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil && len(data) > 0 {
					fileData = data
					//fmt.Printf("💾 ZIP data from body: %d bytes\n", len(fileData))
				}
			}

			// Si no hay datos en el body, intentar descargar por AttachmentId
			if len(fileData) == 0 && part.Body != nil && part.Body.AttachmentId != "" {
				//fmt.Printf("🔗 Downloading ZIP attachment: %s (ID: %s)\n", part.Filename, part.Body.AttachmentId)
				attachment, err := c.service.Users.Messages.Attachments.Get("me", messageID, part.Body.AttachmentId).Do()
				if err != nil {
					fmt.Printf("❌ Error downloading ZIP: %v\n", err)
				} else {
					data, err := base64.URLEncoding.DecodeString(attachment.Data)
					if err != nil {
						fmt.Printf("❌ Error decoding ZIP: %v\n", err)
					} else {
						fileData = data
						//fmt.Printf("💾 Downloaded ZIP: %d bytes\n", len(fileData))
					}
				}
			}

			// Si tenemos datos, agregarlos
			if len(fileData) > 0 {
				zipContents = append(zipContents, fileData)
				zipFilenames = append(zipFilenames, part.Filename)
				//fmt.Printf("✅ ZIP ready for processing: %s (%d bytes)\n", part.Filename, len(fileData))
			} else {
				fmt.Printf("❌ No data found for ZIP: %s\n", part.Filename)
			}
		}

		// Buscar recursivamente en partes hijas
		if part.Parts != nil {
			for _, subpart := range part.Parts {
				searchForZips(subpart)
			}
		}
	}

	// Iniciar búsqueda
	searchForZips(fullMessage.Payload)

	//fmt.Printf("📦 Found %d ZIP attachments in message %s\n", len(zipContents), messageID)
	return zipContents, zipFilenames, nil
}

// matchesFilter aplica filtros adicionales al email
func (c *gmailClient) matchesFilter(email *domain.Email, filter domain.EmailFilter) bool {
	if filter.From != "" && !strings.Contains(strings.ToLower(email.From), strings.ToLower(filter.From)) {
		return false
	}

	if filter.Subject != "" && !strings.Contains(strings.ToLower(email.Subject), strings.ToLower(filter.Subject)) {
		return false
	}

	if !filter.Since.IsZero() && email.ReceivedAt.Before(filter.Since) {
		return false
	}

	return true
}
