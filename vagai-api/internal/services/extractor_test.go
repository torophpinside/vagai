package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTextFromTXT(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "This is a test resume content.\nJohn Doe - Software Engineer"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ExtractTextFromFile(testFile)
	if err != nil {
		t.Fatalf("ExtractTextFromFile() error = %v", err)
	}
	if result != content {
		t.Errorf("ExtractTextFromFile() = %v, expected %v", result, content)
	}
}

func TestExtractTextFromTXT_NonExistent(t *testing.T) {
	_, err := ExtractTextFromFile("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestExtractTextFromUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.csv")

	err := os.WriteFile(testFile, []byte("data"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = ExtractTextFromFile(testFile)
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}

	expected := "formato não suportado: .csv"
	if err.Error() != expected {
		t.Errorf("ExtractTextFromFile() error = %v, expected %v", err.Error(), expected)
	}
}

func TestExtractTextFromFile_ExtensionCase(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.TXT")
	content := "Test content"
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ExtractTextFromFile(testFile)
	if err != nil {
		t.Fatalf("ExtractTextFromFile() error = %v", err)
	}
	if result != content {
		t.Errorf("ExtractTextFromFile() = %v, expected %v", result, content)
	}
}

func TestExtractFromTXT_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ExtractTextFromFile(testFile)
	if err != nil {
		t.Fatalf("ExtractTextFromFile() error = %v", err)
	}
	if result != "" {
		t.Errorf("ExtractTextFromFile() = %v, expected empty string", result)
	}
}

func TestExtractFromTXT_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	largeContent := ""
	for i := 0; i < 1000; i++ {
		largeContent += "Line " + string(rune(i)) + " of test content\n"
	}

	err := os.WriteFile(testFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	result, err := ExtractTextFromFile(testFile)
	if err != nil {
		t.Fatalf("ExtractTextFromFile() error = %v", err)
	}
	if result != largeContent {
		t.Error("ExtractTextFromFile() did not return full content")
	}
}
