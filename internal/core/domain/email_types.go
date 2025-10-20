package domain

import "time"

// EmailCriteria criterios de búsqueda para emails
type EmailCriteria struct {
	From    string
	Subject string
	Since   time.Time
	Unread  bool
	Label    string
}

// EmailFilter filtro para búsqueda de emails
type EmailFilter struct {
	From    string
	Subject string
	Since   time.Time
	Unread  bool
	Label    string
}

// Email representa un correo electrónico
type Email struct {
	ID           string
	From         string
	Subject      string
	Body         string
	ReceivedAt   time.Time
	InternalDate int64
	Attachments  []Attachment
	RawData      []byte
}

// Attachment representa un archivo adjunto
type Attachment struct {
	Filename string
	Content  []byte
	MIMEType string
}
