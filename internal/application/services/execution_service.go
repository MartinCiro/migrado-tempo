package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"email/config"
	"email/internal/core/domain"
	"email/internal/core/ports"
)

type executionService struct {
	emailService   ports.EmailService
	invoiceService ports.InvoiceService
	authService    ports.AuthService
	fileRepo       ports.FileRepository
	cryptoService  ports.CryptoService
	config         *config.AppConfig
	emailRepo      ports.EmailRepository
	aiService      ports.AIService
}

func NewExecutionService(
	emailService ports.EmailService,
	invoiceService ports.InvoiceService,
	authService ports.AuthService,
	fileRepo ports.FileRepository,
	cryptoService ports.CryptoService,
	config *config.AppConfig,
	emailRepo ports.EmailRepository,
) ports.ExecutionService {
	return &executionService{
		emailService:   emailService,
		invoiceService: invoiceService,
		authService:    authService,
		fileRepo:       fileRepo,
		cryptoService:  cryptoService,
		config:         config,
		emailRepo:      emailRepo,
	}
}

func (s *executionService) Run(ctx context.Context) error {
	// 1. Autenticación con la API
	userAPI := os.Getenv("USER_API")
	userPassAPI := os.Getenv("PASS_API")

	if userAPI == "" || userPassAPI == "" {
		return fmt.Errorf("USER_API and PASS_API environment variables are required")
	}

	token, err := s.authService.Login(ctx, map[string]string{
		"correo": userAPI,
		"passwd": userPassAPI,
	})
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// 2. Obtener NITs
	nits, err := s.getNITs(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get NITs: %w", err)
	}

	// 3. Procesar archivos existentes
	if err := s.ProcessExistingFiles(ctx, token); err != nil {
		return fmt.Errorf("failed to process existing files: %w", err)
	}

	// 4. Filtrar y procesar emails
	if err := s.ProcessEmails(ctx, nits, token); err != nil {
		return fmt.Errorf("failed to process emails: %w", err)
	}

	// 5. Limpiar archivos JSON temporal
	jsonPath := s.config.Paths.JSONFile
	if s.fileRepo.FileExists(jsonPath) {
		if err := os.Remove(jsonPath); err != nil {
			//fmt.Printf("Warning: could not delete JSON file: %v\n", err)
		}
	}

	return nil
}

func (s *executionService) getNITs(ctx context.Context, token string) ([]string, error) {
	// Intentar cargar desde JSON local primero
	var nitsData []domain.NIT
	if err := s.fileRepo.LoadJSON(s.config.Paths.JSONFile, &nitsData); err == nil {
		return s.extractNITsFromData(nitsData), nil
	}

	// Si no existe, obtener de la API
	nitsData, err := s.authService.GetNITs(ctx, token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get NITs from API: %w", err)
	}

	// Guardar localmente
	if err := s.fileRepo.SaveJSON(nitsData, s.config.Paths.JSONFile); err != nil {
		//fmt.Printf("Warning: could not save NITs to JSON: %v\n", err)
	}

	return s.extractNITsFromData(nitsData), nil
}

func (s *executionService) extractNITsFromData(nitsData []domain.NIT) []string {
	nits := make([]string, 0, len(nitsData))
	for _, item := range nitsData {
		if item.NIT != "" {
			nits = append(nits, item.NIT)
		}
	}
	return nits
}

func (s *executionService) ProcessExistingFiles(ctx context.Context, token string) error {
	zipFolder := s.config.Paths.ZipFolder

	// Contar archivos ZIP
	count, err := s.fileRepo.CountFiles(zipFolder)
	if err != nil {
		return fmt.Errorf("error counting ZIP files: %w", err)
	}

	if count == 0 {
		fmt.Println("No hay archivos ZIP en la carpeta")
		return nil
	}

	// Procesar cada archivo ZIP
	files, err := filepath.Glob(filepath.Join(zipFolder, "*.zip"))
	if err != nil {
		return fmt.Errorf("error listing ZIP files: %w", err)
	}

	for _, zipFile := range files {
		if err := s.processZipFile(ctx, zipFile, token); err != nil {
			//fmt.Printf("Error processing ZIP file %s: %v\n", zipFile, err)
			continue
		}
	}

	return nil
}

func (s *executionService) processZipFile(ctx context.Context, zipPath, token string) error {
	// Extraer ZIP
	if err := s.fileRepo.ExtractZip(zipPath, s.config.Paths.ZipFolder); err != nil {
		return fmt.Errorf("error extracting ZIP: %w", err)
	}

	// Mover archivos extraídos
	if err := s.moveExtractedFiles(); err != nil {
		return fmt.Errorf("error moving extracted files: %w", err)
	}

	// Procesar archivos XML
	xmlFiles, err := filepath.Glob(filepath.Join(s.config.Paths.XMLFolder, "*.xml"))
	if err != nil {
		return fmt.Errorf("error listing XML files: %w", err)
	}

	for _, xmlFile := range xmlFiles {
		if err := s.processXMLFile(ctx, xmlFile, token); err != nil {
			//fmt.Printf("Error processing XML file %s: %v\n", xmlFile, err)
			continue
		}
	}

	// Limpiar archivos procesados
	if err := s.cleanProcessedFiles(); err != nil {
		return fmt.Errorf("error cleaning processed files: %w", err)
	}

	return nil
}

func (s *executionService) moveExtractedFiles() error {
	// Mover PDFs
	pdfFiles, _ := filepath.Glob(filepath.Join(s.config.Paths.ZipFolder, "*.pdf"))
	for _, pdfFile := range pdfFiles {
		dest := filepath.Join(s.config.Paths.PDFFolder, filepath.Base(pdfFile))
		if err := s.fileRepo.MoveFile(pdfFile, dest); err != nil {
			return err
		}
	}

	// Mover XMLs
	xmlFiles, _ := filepath.Glob(filepath.Join(s.config.Paths.ZipFolder, "*.xml"))
	for _, xmlFile := range xmlFiles {
		dest := filepath.Join(s.config.Paths.XMLFolder, filepath.Base(xmlFile))
		if err := s.fileRepo.MoveFile(xmlFile, dest); err != nil {
			return err
		}
	}

	return nil
}

func (s *executionService) processXMLFile(ctx context.Context, xmlPath, token string) error {
    //xmlName := filepath.Base(xmlPath)
    
    invoiceData, err := s.invoiceService.ProcessInvoiceFromFile(ctx, xmlPath)
    if err != nil || invoiceData == nil || invoiceData.FEVIdFac == "" {
        fmt.Printf("  ⚠️  Traditional processing failed, using AI fallback: %v\n", err)
        
        xmlContent, err := os.ReadFile(xmlPath)
        if err != nil {
            return fmt.Errorf("error reading XML file: %w", err)
        }
        
        invoiceData, err = s.aiService.ProcessInvoiceWithAI(ctx, string(xmlContent))
        if err != nil {
            return fmt.Errorf("AI processing also failed: %w", err)
        }
        
        fmt.Printf("  ✅ AI successfully processed invoice: %s\n", invoiceData.FEVIdFac)
    }

    // Obtener XML en base64
    xmlBase64, err := s.fileRepo.FileToBase64(xmlPath)
    if err != nil {
        return fmt.Errorf("error encoding XML to base64: %w", err)
    }

    // ✅ BUSCAR PDF CORRESPONDIENTE
    pdfBase64, err := s.findAndConvertPDF(xmlPath)
    if err != nil {
        return fmt.Errorf("error finding PDF: %w", err)
    }

    // Completar datos para API
    invoiceData.XMLString = xmlBase64
    invoiceData.PDFBase64 = pdfBase64
    invoiceData.Email = os.Getenv("USER_API")

    fmt.Printf("  📤 Sending invoice to API: %s\n", invoiceData.FEVIdFac)
    
    // Enviar a API
    response, err := s.authService.SendInvoice(ctx, token, invoiceData)
    if err != nil {
        return fmt.Errorf("error sending invoice to API: %w", err)
    }

    if !response.OK {
        return fmt.Errorf("API returned error: %s", response.Message)
    }

    fmt.Printf("  ✅ Factura %s procesada exitosamente\n", invoiceData.FEVIdFac)
    return nil
}

 // findAndConvertPDF busca el PDF que corresponde al XML
func (s *executionService) findAndConvertPDF(xmlPath string) (string, error) {
    xmlName := filepath.Base(xmlPath)
    
    // Extraer el identificador único del XML (remover "ad" y extensión .xml)
    // XML: ad090021983400025030e2bb4.xml → 090021983400025030e2bb4
    xmlID := strings.TrimSuffix(strings.TrimPrefix(xmlName, "ad"), ".xml")
    
    //fmt.Printf("  🔍 Looking for PDF with ID: %s\n", xmlID)
    
    // Construir el nombre esperado del PDF
    expectedPDFName := "de" + xmlID + ".pdf"
    pdfPath := filepath.Join(s.config.Paths.PDFFolder, expectedPDFName)
    
    // Verificar si el PDF existe
    if !s.fileRepo.FileExists(pdfPath) {
        fmt.Printf("  ⚠️  PDF not found: %s\n", expectedPDFName)
        return "", nil
    }
    
    //fmt.Printf("  📄 Found matching PDF: %s\n", expectedPDFName)
    
    // Convertir a base64
    pdfBase64, err := s.fileRepo.FileToBase64(pdfPath)
    if err != nil {
        return "", fmt.Errorf("error encoding PDF to base64: %w", err)
    }
    
    //fmt.Printf("  ✅ PDF converted to base64: %d characters\n", len(pdfBase64))
    return pdfBase64, nil
}

func (s *executionService) cleanProcessedFiles() error {
	if err := s.fileRepo.DeleteFiles(s.config.Paths.PDFFolder, "pdf"); err != nil {
		return err
	}
	if err := s.fileRepo.DeleteFiles(s.config.Paths.XMLFolder, "xml"); err != nil {
		return err
	}
	return nil
}

func (s *executionService) ProcessEmails(ctx context.Context, nits []string, token string) error {
    fmt.Printf("📧 Processing emails from label 'No eliminar/pb scrapping' for %d NITs: %v\n", len(nits), nits)

    filter := domain.EmailFilter{
        Unread: true,
        Label:  "No eliminar/pb scrapping",
    }
    
    emails, err := s.emailRepo.GetEmails(ctx, filter)
    if err != nil {
        return fmt.Errorf("error fetching emails from label: %w", err)
    }

    fmt.Printf("📨 Found %d emails in label\n", len(emails))

    var adjuntos []string
    var correosSinAdjunto []map[string]string
    var emailsFiltrados int

    for _, email := range emails {
        if !s.emailMatchesNITs(email, nits) {
            continue
        }
        emailsFiltrados++

        fmt.Printf("\n✅ 📧 [%d/%d] Processing matching email: %s\n",
            emailsFiltrados, len(emails), email.Subject)

        // ✅ SOLO usar FindZipAttachments para emails que coinciden con NITs
        zipContents, zipFilenames, err := s.emailRepo.FindZipAttachments(ctx, email.ID)
        if err != nil {
            fmt.Printf("  ❌ Error searching for ZIP attachments: %v\n", err)
            continue
        }

        if len(zipContents) > 0 {
            fmt.Printf("  📦 Found %d ZIP attachments\n", len(zipContents))
            for j, zipData := range zipContents {
                filename := zipFilenames[j]
                fmt.Printf("  💾 Processing ZIP: %s (%d bytes)\n", filename, len(zipData))

                // ✅ Guardar el ZIP en disco
                savedPath := s.saveZipAttachment(filename, zipData)
                if savedPath != "" {
                    adjuntos = append(adjuntos, savedPath)
                    fmt.Printf("  ✅ ZIP saved to disk: %s\n", savedPath)
                }
            }
        } else {
            fmt.Printf("  📭 No ZIP attachments found\n")
            // Si no hay adjuntos, extraer datos del asunto
            correoData := s.extractDataFromSubject(email.Subject)
            if correoData != nil {
                correosSinAdjunto = append(correosSinAdjunto, correoData)
                fmt.Printf("  📝 Extracted data from subject: %s\n", correoData["numero_factura"])
            }
        }

        // Marcar como leído después de procesar
        if err := s.emailRepo.DeleteEmail(ctx, email.ID); err != nil {
            fmt.Printf("  ⚠️  Could not mark email as read: %v\n", err)
        }
    }

    // Procesar los ZIPs guardados
    for _, zipPath := range adjuntos {
        if err := s.processZipFile(ctx, zipPath, token); err != nil {
            fmt.Printf("❌ Error processing ZIP file %s: %v\n", zipPath, err)
        }
    }

    // Enviar alarmas para correos sin adjunto
    for _, correoData := range correosSinAdjunto {
        if err := s.SendAlarm(ctx, token, correoData); err != nil {
            fmt.Printf("❌ Error sending alarm for invoice %s: %v\n", 
                correoData["numero_factura"], err)
        }
    }

    fmt.Printf("✅ Processed %d emails with NITs, found %d ZIPs, %d without attachments\n",
        emailsFiltrados, len(adjuntos), len(correosSinAdjunto))

    return nil
}

// Nuevo método para guardar adjuntos ZIP
func (s *executionService) saveZipAttachment(filename string, content []byte) string {
	// Limpiar nombre de archivo
	cleanName := strings.ReplaceAll(filename, "/", "_")
	cleanName = strings.ReplaceAll(cleanName, "\\", "_")
	cleanName = strings.ReplaceAll(cleanName, ":", "_")

	timestamp := time.Now().Format("20060102_150405")
	finalFilename := fmt.Sprintf("%s_%s", timestamp, cleanName)
	filePath := filepath.Join(s.config.Paths.ZipFolder, finalFilename)

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		//fmt.Printf("  ❌ Error saving ZIP file: %v\n", err)
		return ""
	}

	// Verificar que se guardó correctamente
	if fileInfo, err := os.Stat(filePath); err == nil && fileInfo.Size() > 0 {
		return filePath
	}

	return ""
}

func (s *executionService) emailMatchesNITs(email domain.Email, nits []string) bool {
    subject := email.Subject
    
    // Dividir el subject por punto y coma
    parts := strings.Split(subject, ";")
    if len(parts) == 0 {
        return false
    }
    
    // El NIT debería estar en la primera parte
    nitFromSubject := strings.TrimSpace(parts[0])
    
    // Buscar si este NIT está en la lista de NITs buscados
    for _, nit := range nits {
        if nitFromSubject == nit {
            return true
        }
    }
    
    return false
}

func (s *executionService) extractZipAttachments(email domain.Email) []string {
	var zipFiles []string

	//fmt.Printf("  🔍 Searching for ZIP attachments in email %s\n", email.ID)
	//fmt.Printf("  📎 Total attachments: %d\n", len(email.Attachments))

	for _, attachment := range email.Attachments {
		//fmt.Printf("  📎 Attachment %d: %s (%d bytes, %s)\n", i+1, attachment.Filename, len(attachment.Content), attachment.MIMEType)

		if strings.HasSuffix(strings.ToLower(attachment.Filename), ".zip") {
			// Limpiar nombre de archivo de caracteres inválidos
			cleanName := strings.ReplaceAll(attachment.Filename, "/", "_")
			cleanName = strings.ReplaceAll(cleanName, "\\", "_")
			cleanName = strings.ReplaceAll(cleanName, ":", "_")

			timestamp := time.Now().Format("20060102_150405")
			filename := fmt.Sprintf("%s_%s", timestamp, cleanName)
			filePath := filepath.Join(s.config.Paths.ZipFolder, filename)

			//fmt.Printf("  💾 Saving ZIP attachment: %s (%d bytes)\n", filename, len(attachment.Content))

			if err := os.WriteFile(filePath, attachment.Content, 0644); err != nil {
				//fmt.Printf("  ❌ Error saving attachment %s: %v\n", attachment.Filename, err)
				continue
			}

			// Verificar que el archivo se guardó correctamente
			if _, err := os.Stat(filePath); err == nil {
				//fmt.Printf("  ✅ ZIP saved successfully: %s (%d bytes)\n", filename, fileInfo.Size())
				zipFiles = append(zipFiles, filePath)
			} else {
				//fmt.Printf("  ❌ Error verifying saved file: %v\n", err)
			}
		}
	}

	return zipFiles
}

func (s *executionService) extractDataFromSubject(subject string) map[string]string {
	// Replicar la lógica Python: partes_asunto = asunto.split(";")
	parts := strings.Split(subject, ";")
	if len(parts) >= 3 {
		return map[string]string{
			"nit":            strings.TrimSpace(parts[0]),
			"razon_social":   strings.TrimSpace(parts[1]),
			"numero_factura": strings.ToUpper(strings.TrimSpace(parts[2])),
		}
	}

	//fmt.Printf("⚠️  Malformed subject: %s\n", subject)
	return nil
}

func (s *executionService) processZipAttachment(zipPath string, token string) error {
	// Extraer y procesar el ZIP (similar a processZipFile)
	return s.processZipFile(context.Background(), zipPath, token)
}

func (s *executionService) SendAlarm(ctx context.Context, token string, alarmData map[string]string) error {
	currentDate := time.Now().Format("2006-01-02")

	alarm := &domain.InvoiceData{
		FEVIdFac:               alarmData["numero_factura"],
		ProductQuantity:        "0",
		ProductSubtotal:        "0",
		ProductTaxes:           "N/A",
		ProductIVA:             "0",
		ProductVlrTotalChecked: "0",
		ProductVlrTotal:        "0",
		DiscountGlobalApplied:  "0",
		XMLString:              "0",
		PDFBase64:              "0",
		IssueDate:              currentDate,
		PaymentDueDate:         currentDate,
		SellerNIT:              alarmData["nit"],
		BuyerNIT:               alarmData["nit"],
		SocialReasonSeller:     alarmData["razon_social"],
		SocialReasonBuyer:      alarmData["razon_social"],
		ProductDescription:     "N/A",
		ProductVlrUnit:         "0",
		NITSeller:              alarmData["nit"],
		NITBuyer:               alarmData["nit"],
		Email:                  os.Getenv("USER_API"),
		Estado:                 "Alarma",
	}

	_, err := s.authService.SendInvoice(ctx, token, alarm)
	return err
}
