package handlers

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidateUpload_Oversize(t *testing.T) {
	err := validateUpload("resume.pdf", maxUploadSize+1, strings.NewReader(""))
	if err == nil || err != errUploadTooLarge {
		t.Errorf("validateUpload(oversize) = %v, want %v", err, errUploadTooLarge)
	}
}

func TestValidateUpload_UnsupportedExtension(t *testing.T) {
	err := validateUpload("script.exe", 100, bytes.NewReader([]byte("MZ\x90\x00")))
	if err == nil || err != errUploadUnsupported {
		t.Errorf("validateUpload(.exe) = %v, want %v", err, errUploadUnsupported)
	}
}

func TestValidateUpload_PDF(t *testing.T) {
	// Magic bytes do PDF: %PDF-1.7...
	body := bytes.NewReader(append([]byte("%PDF-1.7 test"), make([]byte, 100)...))
	if err := validateUpload("curriculo.pdf", int64(body.Len()), body); err != nil {
		t.Errorf("validateUpload(.pdf) = %v, want nil", err)
	}
}

func TestValidateUpload_DOCX(t *testing.T) {
	docxMagic := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}
	body := bytes.NewReader(append(docxMagic, make([]byte, 200)...))
	if err := validateUpload("resume.docx", int64(body.Len()), body); err != nil {
		t.Errorf("validateUpload(.docx) = %v, want nil", err)
	}
}

func TestValidateUpload_DOC_OLE(t *testing.T) {
	oleMagic := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	body := bytes.NewReader(append(oleMagic, make([]byte, 200)...))
	if err := validateUpload("resume.doc", int64(body.Len()), body); err != nil {
		t.Errorf("validateUpload(.doc) = %v, want nil", err)
	}
}

func TestValidateUpload_TXT_Text(t *testing.T) {
	body := bytes.NewReader([]byte("Hello world sem null bytes"))
	if err := validateUpload("notes.txt", int64(body.Len()), body); err != nil {
		t.Errorf("validateUpload(.txt text) = %v, want nil", err)
	}
}

func TestValidateUpload_TXT_Binary(t *testing.T) {
	// TXT com null byte → rejeitar
	body := bytes.NewReader([]byte{0x00, 0x01, 0x02, 0x03, 0x04})
	err := validateUpload("malicious.txt", int64(body.Len()), body)
	if err == nil || err != errUploadUnsupported {
		t.Errorf("validateUpload(.txt binary) = %v, want %v", err, errUploadUnsupported)
	}
}

func TestValidateUpload_EmptyFile(t *testing.T) {
	err := validateUpload("empty.pdf", 0, bytes.NewReader(nil))
	// PDF magic bytes não presentes → unsupported
	if err == nil || err != errUploadUnsupported {
		t.Errorf("validateUpload(empty) = %v, want %v", err, errUploadUnsupported)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"resume.pdf", "resume.pdf"},
		{"../../etc/passwd", "passwd"},
		{"../../../tmp/evil.pdf", "evil.pdf"},
		{"with spaces (1).pdf", "with spaces (1).pdf"},
		{"./foo/bar.txt", "bar.txt"},
		{"a\x00b.pdf", "ab.pdf"},
		{"normal-doc_v2.pdf", "normal-doc_v2.pdf"},
	}
	for _, tt := range tests {
		got := sanitizeFilename(tt.in)
		if got != tt.want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTruncateName(t *testing.T) {
	short := "Curriculo"
	if got := truncateName(short, 255); got != short {
		t.Errorf("truncateName(short) = %q, want %q", got, short)
	}
	long := strings.Repeat("A", 300)
	got := truncateName(long, 255)
	if len(got) != 255 || got != long[:255] {
		t.Errorf("truncateName(long) = %d chars, want 255", len(got))
	}
}
