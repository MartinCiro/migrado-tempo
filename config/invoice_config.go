package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type InvoiceConfig struct {
	Folders struct {
		Zip  string `json:"zip"`
		XML  string `json:"xml"`
		PDF  string `json:"pdf"`
		JSON string `json:"json"`
	} `json:"folders"`
	Credentials string            `json:"credentials"`
	Token       string            `json:"token"`
	Causacion   map[string]string `json:"causacion"`
	API         struct {
		Host      string            `json:"host"`
		Endpoints map[string]string `json:"endpoints"`
	} `json:"api"`
	Example string `json:"example"`
}

func LoadInvoiceConfig(configPath string) (*InvoiceConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	var config InvoiceConfig
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	return &config, nil
}

type InvoicePaths struct {
	BaseDir     string
	ZipFolder   string
	XMLFolder   string
	PDFFolder   string
	JSONFile    string
	Credentials string
	Token       string
	Example     string
}

func (c *InvoiceConfig) BuildPaths(baseDir string) *InvoicePaths {
	return &InvoicePaths{
		BaseDir:     baseDir,
		ZipFolder:   filepath.Join(baseDir, c.Folders.Zip),
		XMLFolder:   filepath.Join(baseDir, c.Folders.XML),
		PDFFolder:   filepath.Join(baseDir, c.Folders.PDF),
		JSONFile:    filepath.Join(baseDir, c.Folders.JSON),
		Credentials: filepath.Join(baseDir, c.Credentials),
		Token:       filepath.Join(baseDir, c.Token),
		Example:     filepath.Join(baseDir, c.Example),
	}
}

func (p *InvoicePaths) EnsureDirectories() error {
	dirs := []string{
		p.ZipFolder,
		p.XMLFolder,
		p.PDFFolder,
		filepath.Dir(p.JSONFile),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}
	return nil
}
