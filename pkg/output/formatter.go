package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// Format represents the output format
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
)

// Formatter handles output formatting
type Formatter struct {
	format Format
	writer io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(format Format, writer io.Writer) *Formatter {
	if writer == nil {
		writer = os.Stdout
	}
	return &Formatter{
		format: format,
		writer: writer,
	}
}

// Format outputs data in the specified format
func (f *Formatter) Format(data interface{}) error {
	switch f.format {
	case FormatTable:
		return f.formatTable(data)
	case FormatJSON:
		return f.formatJSON(data)
	case FormatYAML:
		return f.formatYAML(data)
	case FormatCSV:
		return f.formatCSV(data)
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}
}

// formatJSON outputs data as JSON
func (f *Formatter) formatJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// formatYAML outputs data as YAML
func (f *Formatter) formatYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

// formatTable outputs data as a table using tablewriter
func (f *Formatter) formatTable(data interface{}) error {
	table := tablewriter.NewWriter(f.writer)

	switch v := data.(type) {
	case TableData:
		table.Header(convertToInterface(v.Headers)...)
		for _, row := range v.Rows {
			table.Append(convertToInterface(row)...)
		}
	case *TableData:
		table.Header(convertToInterface(v.Headers)...)
		for _, row := range v.Rows {
			table.Append(convertToInterface(row)...)
		}
	default:
		return fmt.Errorf("table format requires TableData struct, got %T", data)
	}

	return table.Render()
}

// convertToInterface converts a slice of strings to a slice of interface{}
func convertToInterface(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, s := range strings {
		result[i] = s
	}
	return result
}

// formatCSV outputs data as CSV
func (f *Formatter) formatCSV(data interface{}) error {
	switch v := data.(type) {
	case TableData:
		// Write headers
		fmt.Fprintln(f.writer, strings.Join(v.Headers, ","))
		// Write rows
		for _, row := range v.Rows {
			fmt.Fprintln(f.writer, strings.Join(row, ","))
		}
	case *TableData:
		// Write headers
		fmt.Fprintln(f.writer, strings.Join(v.Headers, ","))
		// Write rows
		for _, row := range v.Rows {
			fmt.Fprintln(f.writer, strings.Join(row, ","))
		}
	default:
		return fmt.Errorf("CSV format requires TableData struct")
	}
	return nil
}

// TableData represents tabular data
type TableData struct {
	Headers []string
	Rows    [][]string
}

// NewTableData creates a new TableData instance
func NewTableData(headers []string) *TableData {
	return &TableData{
		Headers: headers,
		Rows:    make([][]string, 0),
	}
}

// AddRow adds a row to the table data
func (td *TableData) AddRow(row []string) {
	td.Rows = append(td.Rows, row)
}

// FormatValue formats a value for display
func FormatValue(value interface{}) string {
	if value == nil {
		return "-"
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return "-"
		}
		return v
	case bool:
		if v {
			return "✓"
		}
		return "✗"
	case int, int64, int32:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatBytes formats bytes as human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatPercentage formats a percentage value
func FormatPercentage(value, total int64) string {
	if total == 0 {
		return "0%"
	}
	pct := float64(value) / float64(total) * 100
	return fmt.Sprintf("%.1f%%", pct)
}

// ParseFormat parses a format string
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "table":
		return FormatTable, nil
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	case "csv":
		return FormatCSV, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", s)
	}
}
