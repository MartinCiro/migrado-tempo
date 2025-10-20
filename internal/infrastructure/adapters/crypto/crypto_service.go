package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"email/internal/core/ports"
)

type cryptoService struct {
	key []byte
}

func NewCryptoService(encryptionKey string) (ports.CryptoService, error) {
	var key []byte
	var err error

	// Intentar decodificar como base64 primero
	if strings.Contains(encryptionKey, "=") {
		key, err = base64.StdEncoding.DecodeString(encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("error decoding base64 key: %w", err)
		}
	} else {
		// Usar como string directo
		key = []byte(encryptionKey)
	}

	// La clave debe ser de 16, 24 o 32 bytes para AES
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 16, 24 or 32 bytes, got %d bytes", len(key))
	}

	return &cryptoService{key: key}, nil
}

func (c *cryptoService) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error creating GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("error generating nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func (c *cryptoService) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, fmt.Errorf("error creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("error decrypting: %w", err)
	}

	return plaintext, nil
}

func (c *cryptoService) EncryptString(data string) (string, error) {
	encrypted, err := c.Encrypt([]byte(data))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(encrypted), nil
}

func (c *cryptoService) DecryptString(data string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %w", err)
	}

	decrypted, err := c.Decrypt(decoded)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
