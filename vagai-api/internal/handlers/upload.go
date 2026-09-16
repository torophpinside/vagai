package handlers

import (
	"bytes"
	"errors"
	"io"
	"path/filepath"
	"strings"
)

const maxUploadSize int64 = 10 << 20 // 10 MB

var (
	errUploadTooLarge    = errors.New("arquivo excede o limite de 10MB")
	errUploadUnsupported = errors.New("tipo de arquivo não suportado")
)

// validateUpload valida tamanho e magic bytes antes de persistir o arquivo,
// evitando uploads de binários maliciosos ou excessivamente grandes.
func validateUpload(fileName string, size int64, content io.Reader) error {
	if size > maxUploadSize {
		return errUploadTooLarge
	}

	header := make([]byte, 512)
	n, _ := io.ReadFull(content, header)
	header = header[:n]

	if !validFileMagic(fileName, header) {
		return errUploadUnsupported
	}
	return nil
}

// validFileMagic valida magic bytes conhecidos, incluindo DOC antigo (OLE),
// DOCX/XLSX (PK zip), PDF e .txt (rejeita binário contendo null byte).
func validFileMagic(fileName string, header []byte) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".pdf":
		return bytes.HasPrefix(header, []byte("%PDF-"))
	case ".docx", ".doc":
		if len(header) >= 8 {
			// DOCX/XLSX (zip) ou DOC (OLE Compound File Binary Format)
			if bytes.HasPrefix(header, []byte{0x50, 0x4B, 0x03, 0x04}) {
				return true
			}
			if bytes.Equal(header[:8], []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}) {
				return true
			}
		}
		return false
	case ".txt":
		// Arquivo texto é aceito desde que não contenha null byte (binário).
		return bytes.IndexByte(header, 0) == -1
	default:
		return false
	}
}

// sanitizeFilename retira path components e caracteres de controle do nome
// original do arquivo, evitando path traversal na criação do filepath.
func sanitizeFilename(name string) string {
	base := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	return strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		return r
	}, base)
}

// truncateName limita o tamanho do nome para 255 caracteres UTF-8, evitando
// DoS via coluna VARCHAR no banco de dados.
func truncateName(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
