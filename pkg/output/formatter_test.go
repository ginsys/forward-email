package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name   string
		format Format
		writer *bytes.Buffer
	}{
		{"JSON formatter", FormatJSON, &bytes.Buffer{}},
		{"YAML formatter", FormatYAML, &bytes.Buffer{}},
		{"Table formatter", FormatTable, &bytes.Buffer{}},
		{"CSV formatter", FormatCSV, &bytes.Buffer{}},
		{"Plain formatter", FormatPlain, &bytes.Buffer{}},
		{"With nil writer", FormatJSON, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format, tt.writer)

			if formatter == nil {
				t.Fatal("Expected formatter but got nil")
			}

			if formatter.format != tt.format {
				t.Errorf("Expected format %s, got %s", tt.format, formatter.format)
			}

			if tt.writer == nil && formatter.writer == nil {
				t.Error("Expected default writer when nil provided")
			}
		})
	}
}

func TestFormatter_FormatJSON(t *testing.T) {
	testData := map[string]interface{}{
		"name":   "test",
		"count":  42,
		"active": true,
	}

	var buf bytes.Buffer
	formatter := NewFormatter(FormatJSON, &buf)

	err := formatter.Format(testData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify content
	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", result["name"])
	}
	if result["count"].(float64) != 42 {
		t.Errorf("Expected count 42, got %v", result["count"])
	}
	if result["active"] != true {
		t.Errorf("Expected active true, got %v", result["active"])
	}
}

func TestFormatter_FormatYAML(t *testing.T) {
	testData := map[string]interface{}{
		"name":   "test",
		"count":  42,
		"active": true,
	}

	var buf bytes.Buffer
	formatter := NewFormatter(FormatYAML, &buf)

	err := formatter.Format(testData)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify it's valid YAML
	var result map[string]interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Invalid YAML output: %v", err)
	}

	// Verify content
	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", result["name"])
	}
	if result["count"] != 42 {
		t.Errorf("Expected count 42, got %v", result["count"])
	}
	if result["active"] != true {
		t.Errorf("Expected active true, got %v", result["active"])
	}
}

func TestFormatter_FormatTable(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		shouldError bool
		expectError string
	}{
		{
			name: "TableData struct",
			data: TableData{
				Headers: []string{"Name", "Count", "Active"},
				Rows: [][]string{
					{"test1", "10", "true"},
					{"test2", "20", "false"},
				},
			},
			shouldError: false,
		},
		{
			name: "TableData pointer",
			data: &TableData{
				Headers: []string{"Name", "Count"},
				Rows: [][]string{
					{"test1", "10"},
				},
			},
			shouldError: false,
		},
		{
			name:        "invalid data type",
			data:        map[string]string{"key": "value"},
			shouldError: true,
			expectError: "table format requires TableData struct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(FormatTable, &buf)

			err := formatter.Format(tt.data)

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				if !strings.Contains(err.Error(), tt.expectError) {
					t.Errorf("Expected error to contain %q, got %q", tt.expectError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := buf.String()
			if output == "" {
				t.Error("Expected table output but got empty string")
			}

			// Basic verification - tablewriter uses Unicode box-drawing characters
			if !strings.Contains(output, "│") && !strings.Contains(output, "┌") {
				t.Error("Expected table output to contain Unicode table formatting characters")
			}

			// Verify headers are present (they should be in uppercase by tablewriter)
			switch data := tt.data.(type) {
			case TableData:
				for _, header := range data.Headers {
					upperHeader := strings.ToUpper(header)
					if !strings.Contains(output, upperHeader) {
						t.Errorf("Expected output to contain header %q", upperHeader)
					}
				}
			case *TableData:
				for _, header := range data.Headers {
					upperHeader := strings.ToUpper(header)
					if !strings.Contains(output, upperHeader) {
						t.Errorf("Expected output to contain header %q", upperHeader)
					}
				}
			}
		})
	}
}

func TestFormatter_FormatCSV(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		shouldError bool
		expected    []string
	}{
		{
			name: "TableData struct",
			data: TableData{
				Headers: []string{"Name", "Count", "Active"},
				Rows: [][]string{
					{"test1", "10", "true"},
					{"test2", "20", "false"},
				},
			},
			shouldError: false,
			expected:    []string{"Name,Count,Active", "test1,10,true", "test2,20,false"},
		},
		{
			name: "TableData pointer",
			data: &TableData{
				Headers: []string{"ID", "Name"},
				Rows: [][]string{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
			shouldError: false,
			expected:    []string{"ID,Name", "1,Alice", "2,Bob"},
		},
		{
			name:        "invalid data type",
			data:        "invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(FormatCSV, &buf)

			err := formatter.Format(tt.data)

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := strings.TrimSpace(buf.String())
			lines := strings.Split(output, "\n")

			if len(lines) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(lines))
			}

			for i, expectedLine := range tt.expected {
				if i < len(lines) && strings.TrimSpace(lines[i]) != expectedLine {
					t.Errorf("Line %d: expected %q, got %q", i, expectedLine, strings.TrimSpace(lines[i]))
				}
			}
		})
	}
}

func TestFormatter_FormatPlain(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		shouldError bool
		expected    []string
	}{
		{
			name: "TableData struct",
			data: TableData{
				Headers: []string{"Name", "Count", "Active"},
				Rows: [][]string{
					{"test1", "10", "true"},
					{"test2", "20", "false"},
				},
			},
			shouldError: false,
			expected: []string{
				"Name   Count  Active",
				"test1  10     true",
				"test2  20     false",
			},
		},
		{
			name: "TableData pointer",
			data: &TableData{
				Headers: []string{"ID", "Name"},
				Rows: [][]string{
					{"1", "Alice"},
					{"2", "Bob"},
				},
			},
			shouldError: false,
			expected: []string{
				"ID  Name",
				"1   Alice",
				"2   Bob",
			},
		},
		{
			name: "varying column widths",
			data: &TableData{
				Headers: []string{"Short", "VeryLongHeader"},
				Rows: [][]string{
					{"A", "B"},
					{"ShortValue", "VeryVeryLongValue"},
				},
			},
			shouldError: false,
			expected: []string{
				"Short       VeryLongHeader",
				"A           B",
				"ShortValue  VeryVeryLongValue",
			},
		},
		{
			name:        "invalid data type",
			data:        "invalid",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewFormatter(FormatPlain, &buf)

			err := formatter.Format(tt.data)

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := strings.TrimSpace(buf.String())
			lines := strings.Split(output, "\n")

			if len(lines) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(lines))
			}

			for i, expectedLine := range tt.expected {
				if i < len(lines) && lines[i] != expectedLine {
					t.Errorf("Line %d:\nexpected: %q\ngot:      %q", i, expectedLine, lines[i])
				}
			}
		})
	}
}

func TestFormatter_UnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewFormatter("unsupported", &buf)

	err := formatter.Format(map[string]string{"key": "value"})
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}

	expectedError := "unsupported format"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
	}
}

func TestTableData_NewTableData(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	tableData := NewTableData(headers)

	if tableData == nil {
		t.Fatal("Expected TableData but got nil")
	}

	if len(tableData.Headers) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(tableData.Headers))
	}

	for i, header := range headers {
		if tableData.Headers[i] != header {
			t.Errorf("Header %d: expected %q, got %q", i, header, tableData.Headers[i])
		}
	}

	if tableData.Rows == nil {
		t.Error("Expected non-nil Rows slice")
	}

	if len(tableData.Rows) != 0 {
		t.Errorf("Expected empty Rows slice, got %d rows", len(tableData.Rows))
	}
}

func TestTableData_AddRow(t *testing.T) {
	tableData := NewTableData([]string{"Name", "Age"})

	// Add first row
	row1 := []string{"Alice", "30"}
	tableData.AddRow(row1)

	if len(tableData.Rows) != 1 {
		t.Errorf("Expected 1 row, got %d", len(tableData.Rows))
	}

	for i, value := range row1 {
		if tableData.Rows[0][i] != value {
			t.Errorf("Row 0, Column %d: expected %q, got %q", i, value, tableData.Rows[0][i])
		}
	}

	// Add second row
	row2 := []string{"Bob", "25"}
	tableData.AddRow(row2)

	if len(tableData.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(tableData.Rows))
	}

	for i, value := range row2 {
		if tableData.Rows[1][i] != value {
			t.Errorf("Row 1, Column %d: expected %q, got %q", i, value, tableData.Rows[1][i])
		}
	}
}

// Note: Tests for FormatValue, TruncateString, FormatBytes, and FormatPercentage
// are already covered in domain_test.go

func TestParseFormat(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Format
		shouldError bool
	}{
		{"table", "table", FormatTable, false},
		{"TABLE", "TABLE", FormatTable, false},
		{"json", "json", FormatJSON, false},
		{"JSON", "JSON", FormatJSON, false},
		{"yaml", "yaml", FormatYAML, false},
		{"yml", "yml", FormatYAML, false},
		{"YAML", "YAML", FormatYAML, false},
		{"csv", "csv", FormatCSV, false},
		{"CSV", "CSV", FormatCSV, false},
		{"plain", "plain", FormatPlain, false},
		{"PLAIN", "PLAIN", FormatPlain, false},
		{"invalid", "invalid", "", true},
		{"empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFormat(tt.input)

			if tt.shouldError {
				if err == nil {
					t.Fatal("Expected error but got success")
				}
				expectedError := "unsupported format"
				if !strings.Contains(err.Error(), expectedError) {
					t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func Test_convertToInterface(t *testing.T) {
	input := []string{"a", "b", "c"}
	result := convertToInterface(input)

	if len(result) != len(input) {
		t.Errorf("Expected length %d, got %d", len(input), len(result))
	}

	for i, value := range input {
		if result[i] != value {
			t.Errorf("Index %d: expected %q, got %v", i, value, result[i])
		}
	}
}

func Test_convertToInterface_Empty(t *testing.T) {
	input := []string{}
	result := convertToInterface(input)

	if len(result) != 0 {
		t.Errorf("Expected empty slice, got length %d", len(result))
	}
}
