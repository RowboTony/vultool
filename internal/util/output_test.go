package util

import (
	"bytes"
	"strings"
	"testing"
)

// TestData represents test data for serialization
type TestData struct {
	Name    string `json:"name" yaml:"name"`
	Value   int    `json:"value" yaml:"value"`
	Enabled bool   `json:"enabled" yaml:"enabled"`
}

func TestOutputResult_JSON(t *testing.T) {
	data := TestData{
		Name:    "test",
		Value:   42,
		Enabled: true,
	}

	var buf bytes.Buffer
	err := OutputResult(data, "json", &buf)
	if err != nil {
		t.Fatalf("Expected no error for JSON output, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `"name": "test"`) {
		t.Errorf("Expected JSON output to contain name field, got: %s", output)
	}
	if !strings.Contains(output, `"value": 42`) {
		t.Errorf("Expected JSON output to contain value field, got: %s", output)
	}
}

func TestOutputResult_YAML(t *testing.T) {
	data := TestData{
		Name:    "test",
		Value:   42,
		Enabled: true,
	}

	var buf bytes.Buffer
	err := OutputResult(data, "yaml", &buf)
	if err != nil {
		t.Fatalf("Expected no error for YAML output, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "name: test") {
		t.Errorf("Expected YAML output to contain name field, got: %s", output)
	}
	if !strings.Contains(output, "value: 42") {
		t.Errorf("Expected YAML output to contain value field, got: %s", output)
	}
}

func TestOutputResult_TOML_NotSupported(t *testing.T) {
	data := TestData{
		Name:    "test",
		Value:   42,
		Enabled: true,
	}

	var buf bytes.Buffer
	err := OutputResult(data, "toml", &buf)
	if err == nil {
		t.Fatal("Expected error for TOML output, got nil")
	}

	expectedError := "TOML not yet supported; TODO implement"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got: %s", expectedError, err.Error())
	}

	// Buffer should be empty since TOML is not supported
	if buf.Len() > 0 {
		t.Errorf("Expected empty buffer for unsupported TOML, got: %s", buf.String())
	}
}

func TestOutputResult_UnsupportedFormat(t *testing.T) {
	data := TestData{
		Name:    "test",
		Value:   42,
		Enabled: true,
	}

	var buf bytes.Buffer
	err := OutputResult(data, "xml", &buf)
	if err == nil {
		t.Fatal("Expected error for unsupported format, got nil")
	}

	expectedError := "unsupported format: xml"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got: %s", expectedError, err.Error())
	}
}

func TestOutputResult_ErrorPropagation(t *testing.T) {
	// Test with invalid data that should cause JSON marshaling to fail
	invalidData := make(chan int) // channels cannot be marshaled to JSON

	var buf bytes.Buffer
	err := OutputResult(invalidData, "json", &buf)
	if err == nil {
		t.Fatal("Expected error for invalid JSON data, got nil")
	}

	if !strings.Contains(err.Error(), "failed to serialize to JSON") {
		t.Errorf("Expected JSON serialization error, got: %s", err.Error())
	}
}
