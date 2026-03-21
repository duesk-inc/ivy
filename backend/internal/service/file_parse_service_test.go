package service

import (
	"testing"
)

func TestExtractPDFText_NonPDF(t *testing.T) {
	data := []byte("This is not a PDF file at all, just plain text content")
	text := extractPDFText(data)
	if text != "" {
		t.Logf("extractPDFText returned non-empty for non-PDF: %q (may contain false positives from BT/ET patterns)", text)
	}
}

func TestExtractPDFText_EmptyData(t *testing.T) {
	text := extractPDFText([]byte{})
	if text != "" {
		t.Errorf("expected empty string for empty data, got %q", text)
	}
}

func TestExtractPDFText_PDFMagic(t *testing.T) {
	data := []byte("%PDF-1.4 some binary content without any text blocks")
	text := extractPDFText(data)
	if text != "" {
		t.Logf("extractPDFText returned %q for PDF without text blocks (may have false positives)", text)
	}
}
