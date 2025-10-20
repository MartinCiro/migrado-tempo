package files

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"email/internal/core/ports"
)

type fileRepository struct {
	baseDir string
}

func NewFileRepository(baseDir string) ports.FileRepository {
	return &fileRepository{
		baseDir: baseDir,
	}
}

func (f *fileRepository) CountFiles(directory string) (int, error) {
	// directory ya viene con el path completo, no necesitamos unirlo con baseDir
	dirPath := directory

	fmt.Printf("📁 Counting files in: %s\n", dirPath)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, fmt.Errorf("error reading directory: %w", err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	fmt.Printf("📊 Found %d files in directory\n", count)
	return count, nil
}

func (f *fileRepository) ExtractZip(zipPath, destPath string) error {
	fmt.Printf("📦 Extracting ZIP: %s to %s\n", zipPath, destPath)

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("error opening zip file: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		filePath := filepath.Join(destPath, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return fmt.Errorf("error creating directory: %w", err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("error creating file: %w", err)
		}

		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return fmt.Errorf("error opening zipped file: %w", err)
		}

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			dstFile.Close()
			srcFile.Close()
			return fmt.Errorf("error copying file content: %w", err)
		}

		dstFile.Close()
		srcFile.Close()
	}

	// Eliminar el archivo ZIP después de extraer
	if err := os.Remove(zipPath); err != nil {
		return fmt.Errorf("error deleting zip file: %w", err)
	}

	return nil
}

func (f *fileRepository) MoveFile(source, destination string) error {
	fmt.Printf("📁 Moving file: %s to %s\n", source, destination)

	// Crear directorio de destino si no existe
	if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
		return fmt.Errorf("error creating destination directory: %w", err)
	}

	if err := os.Rename(source, destination); err != nil {
		return fmt.Errorf("error moving file: %w", err)
	}

	return nil
}

func (f *fileRepository) DeleteFiles(directory, extension string) error {
	fmt.Printf("🗑️  Deleting files in: %s with extension: %s\n", directory, extension)

	files, err := filepath.Glob(filepath.Join(directory, "*."+extension))
	if err != nil {
		return fmt.Errorf("error listing files: %w", err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("error deleting file %s: %w", file, err)
		}
	}

	return nil
}

func (f *fileRepository) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (f *fileRepository) WriteFile(filePath string, data []byte) error {
	fmt.Printf("📝 Writing file: %s\n", filePath)

	// Crear directorio si no existe
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

func (f *fileRepository) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Métodos adicionales específicos para el procesamiento
func (f *fileRepository) ReadXMLFile(xmlPath string) (string, error) {
	data, err := f.ReadFile(xmlPath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func (f *fileRepository) FileToBase64(filePath string) (string, error) {
	data, err := f.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (f *fileRepository) SaveJSON(data interface{}, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}
	return f.WriteFile(filePath, jsonData)
}

func (f *fileRepository) LoadJSON(filePath string, v interface{}) error {
	data, err := f.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %w", err)
	}
	return json.Unmarshal(data, v)
}
