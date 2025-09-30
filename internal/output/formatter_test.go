package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"ollamacli/internal/client"
)

func TestFormatModelsText(t *testing.T) {
	var buf bytes.Buffer
	formatter := New(Options{
		Format: FormatText,
		Quiet:  false,
		Writer: &buf,
	})

	models := []client.Model{
		{
			Name:       "llama2",
			Size:       4096000000,
			ModifiedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Digest:     "sha256:1234567890abcdef",
		},
		{
			Name:       "codellama",
			Size:       7000000000,
			ModifiedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			Digest:     "sha256:fedcba0987654321",
		},
	}

	err := formatter.FormatModels(models)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Check for header
	if !strings.Contains(output, "NAME") {
		t.Error("Output should contain header with NAME")
	}

	// Check for model names
	if !strings.Contains(output, "llama2") {
		t.Error("Output should contain llama2")
	}
	if !strings.Contains(output, "codellama") {
		t.Error("Output should contain codellama")
	}

	// Check for formatted sizes (allow some tolerance in formatting)
	if !strings.Contains(output, "3.8") && !strings.Contains(output, "3.9") {
		t.Errorf("Output should contain formatted size for llama2, got: %s", output)
	}
	if !strings.Contains(output, "6.5") && !strings.Contains(output, "6.6") {
		t.Errorf("Output should contain formatted size for codellama, got: %s", output)
	}
}

func TestFormatModelsQuiet(t *testing.T) {
	var buf bytes.Buffer
	formatter := New(Options{
		Format: FormatText,
		Quiet:  true,
		Writer: &buf,
	})

	models := []client.Model{
		{Name: "llama2"},
		{Name: "codellama"},
	}

	err := formatter.FormatModels(models)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in quiet mode, got %d", len(lines))
	}

	if lines[0] != "llama2" {
		t.Errorf("Expected first line to be 'llama2', got '%s'", lines[0])
	}
	if lines[1] != "codellama" {
		t.Errorf("Expected second line to be 'codellama', got '%s'", lines[1])
	}
}

func TestFormatModelsJSON(t *testing.T) {
	var buf bytes.Buffer
	formatter := New(Options{
		Format: FormatJSON,
		Quiet:  false,
		Writer: &buf,
	})

	models := []client.Model{
		{Name: "llama2"},
	}

	err := formatter.FormatModels(models)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, `"models"`) {
		t.Error("JSON output should contain 'models' key")
	}
	if !strings.Contains(output, `"llama2"`) {
		t.Error("JSON output should contain model name")
	}
}

func TestFormatSize(t *testing.T) {
	formatter := &formatter{}

	tests := []struct {
		bytes    int64
		expected string
	}{
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
		{4096000000, "3.8 GB"},
	}

	for _, test := range tests {
		result := formatter.formatSize(test.bytes)
		if result != test.expected {
			t.Errorf("Expected %s for %d bytes, got %s", test.expected, test.bytes, result)
		}
	}
}

func TestTruncateString(t *testing.T) {
	formatter := &formatter{}

	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"exactly10c", 10, "exactly10c"},
		{"", 5, ""},
	}

	for _, test := range tests {
		result := formatter.truncateString(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("Expected '%s' for input '%s' with maxLen %d, got '%s'",
				test.expected, test.input, test.maxLen, result)
		}
	}
}

func TestStreamFormatter(t *testing.T) {
	var buf bytes.Buffer
	sf := NewStream(Options{
		Format: FormatText,
		Writer: &buf,
	})

	err := sf.WriteChunk("Hello ")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	err = sf.WriteChunk("World")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if sf.GetBuffer() != "Hello World" {
		t.Errorf("Expected buffer to contain 'Hello World', got '%s'", sf.GetBuffer())
	}

	if buf.String() != "Hello World" {
		t.Errorf("Expected output to be 'Hello World', got '%s'", buf.String())
	}

	sf.ClearBuffer()
	if sf.GetBuffer() != "" {
		t.Errorf("Expected buffer to be empty after clear, got '%s'", sf.GetBuffer())
	}
}