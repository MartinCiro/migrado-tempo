package ports

import (
	"context"
	"email/internal/core/domain"
)

// EmailRepository define las operaciones de almacenamiento de emails
type EmailRepository interface {
	Connect() error
	Disconnect() error
	SearchEmails(criteria domain.EmailCriteria) ([]domain.Email, error)
	GetEmails(ctx context.Context, filter domain.EmailFilter) ([]domain.Email, error)
	GetEmailsByLabel(ctx context.Context, labelName string) ([]domain.Email, error)
	SaveEmail(ctx context.Context, email *domain.Email) error
	DeleteEmail(ctx context.Context, emailID string) error
	GetLabelID(ctx context.Context, labelName string) (string, error)
	FindZipAttachments(ctx context.Context, messageID string) ([][]byte, []string, error)
}

// FileRepository define las operaciones de archivos
// FileRepository define las operaciones de archivos
type FileRepository interface {
	ExtractZip(zipPath, destPath string) error
	MoveFile(source, destination string) error
	DeleteFiles(directory, extension string) error
	CountFiles(directory string) (int, error)
	ReadFile(filePath string) ([]byte, error)
	WriteFile(filePath string, data []byte) error
	FileExists(filePath string) bool
	// Agregar estos métodos:
	LoadJSON(filePath string, v interface{}) error
	SaveJSON(data interface{}, filePath string) error
	FileToBase64(filePath string) (string, error)
	ReadXMLFile(xmlPath string) (string, error)
}

// CryptoService define las operaciones de cifrado/descifrado
type CryptoService interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
	EncryptString(data string) (string, error)
	DecryptString(data string) (string, error)
}
