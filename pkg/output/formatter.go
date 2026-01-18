// Package output provides formatting and display utilities for CLI output.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/term"
	yaml "gopkg.in/yaml.v3"
)

// Standard labels for boolean value formatting across all output formats.
const (
	yesLabel = "Yes" // Standard label for true boolean values
	noLabel  = "No"  // Standard label for false boolean values
)

// Format represents the available output format types for CLI responses.
// Each format provides different benefits: table for readability, JSON/YAML for automation,
// and CSV for spreadsheet integration.
type Format string

// Output format constants for different display types
const (
	FormatTable Format = "table" // Human-readable table with proper column alignment
	FormatJSON  Format = "json"  // Machine-readable JSON for API integration
	FormatYAML  Format = "yaml"  // Human-readable YAML for configuration
	FormatCSV   Format = "csv"   // Comma-separated values for spreadsheet import
	FormatPlain Format = "plain" // Borderless, fixed-width columns without truncation
)

// Formatter handles output formatting for CLI responses.
// It supports multiple output formats and provides consistent formatting
// across all CLI commands with proper terminal width detection and alignment.
type Formatter struct {
	format Format    // The output format to use for rendering
	writer io.Writer // The output destination (typically os.Stdout)
}

// NewFormatter creates a new output formatter with the specified format and writer.
// If writer is nil, it defaults to os.Stdout. The formatter will handle terminal
// width detection and proper alignment for table outputs automatically.
func NewFormatter(format Format, writer io.Writer) *Formatter {
	if writer == nil {
		writer = os.Stdout
	}
	return &Formatter{
		format: format,
		writer: writer,
	}
}

// Format renders the provided data in the formatter's configured output format.
// It handles type detection, proper formatting, and output generation for tables,
// JSON, YAML, and CSV formats. The data structure determines the specific formatting logic.
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
	case FormatPlain:
		return f.formatPlain(data)
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
	defer func() { _ = encoder.Close() }()
	return encoder.Encode(data)
}

// formatTable outputs data as a table using tablewriter
func (f *Formatter) formatTable(data interface{}) error {
	table := tablewriter.NewWriter(f.writer)

	// Get terminal width for column sizing
	terminalWidth := getTerminalWidth()

	switch v := data.(type) {
	case TableData:
		table.Header(convertToInterface(v.Headers)...)
		// Apply intelligent text wrapping for long content using current terminal width
		wrappedRows := f.wrapTableContentWithWidth(v.Rows, v.Headers, terminalWidth)
		for _, row := range wrappedRows {
			_ = table.Append(convertToInterface(row)...)
		}
	case *TableData:
		table.Header(convertToInterface(v.Headers)...)
		// Apply intelligent text wrapping for long content using current terminal width
		wrappedRows := f.wrapTableContentWithWidth(v.Rows, v.Headers, terminalWidth)
		for _, row := range wrappedRows {
			_ = table.Append(convertToInterface(row)...)
		}
	default:
		return fmt.Errorf("table format requires TableData struct, got %T", data)
	}

	return table.Render()
}

// wrapTableContent intelligently wraps long content in table cells (deprecated, use wrapTableContentWithWidth)
// Removed: wrapTableContent (deprecated)

// wrapTableContentWithWidth intelligently wraps long content in table cells using specified terminal width
func (f *Formatter) wrapTableContentWithWidth(rows [][]string, headers []string, terminalWidth int) [][]string {
	if len(rows) == 0 || len(headers) == 0 {
		return rows
	}

	wrappedRows := make([][]string, len(rows))

	// Calculate column widths to use full terminal width
	colWidths := f.calculateColumnWidthsForTerminal(headers, rows, terminalWidth)

	// Process each row
	for rowIdx, row := range rows {
		wrappedRow := make([]string, len(row))
		for colIdx, cell := range row {
			if colIdx < len(colWidths) {
				wrappedRow[colIdx] = f.wrapCellContent(cell, colWidths[colIdx])
			} else {
				wrappedRow[colIdx] = cell
			}
		}
		wrappedRows[rowIdx] = wrappedRow
	}

	return wrappedRows
}

// calculateColumnWidthsForTerminal distributes terminal width among columns
func (f *Formatter) calculateColumnWidthsForTerminal(headers []string, rows [][]string, terminalWidth int) []int {
	if terminalWidth <= 0 {
		terminalWidth = 80
	}

	numColumns := len(headers)
	if numColumns == 0 {
		return []int{}
	}

	// Account for table borders and padding:
	// Each column needs: | content | (3 chars: border + space + space)
	// Plus one final border
	overhead := (numColumns * 3) + 1
	availableWidth := terminalWidth - overhead

	if availableWidth < numColumns*5 { // Minimum 5 chars per column
		availableWidth = numColumns * 5
	}

	colWidths := make([]int, numColumns)

	// Calculate actual content widths
	for i, header := range headers {
		maxWidth := len(header)
		for _, row := range rows {
			if i < len(row) && len(row[i]) > maxWidth {
				maxWidth = len(row[i])
			}
		}
		colWidths[i] = maxWidth
	}

	// Distribute available width proportionally
	totalContentWidth := 0
	for _, width := range colWidths {
		totalContentWidth += width
	}

	if totalContentWidth <= availableWidth {
		// Content fits naturally, no need to truncate
		return colWidths
	}

	// Content doesn't fit, use intelligent width distribution
	colWidths = f.distributeWidthIntelligently(headers, colWidths, availableWidth)

	return colWidths
}

// distributeWidthIntelligently distributes available width with priority for certain column types
func (f *Formatter) distributeWidthIntelligently(headers []string, originalWidths []int, availableWidth int) []int {
	colWidths := make([]int, len(originalWidths))
	copy(colWidths, originalWidths)

	// Define priority columns that should get their natural width when possible
	priorityColumns := make([]bool, len(headers))
	contentColumns := make([]bool, len(headers))

	for i, header := range headers {
		headerLower := strings.ToLower(header)
		// Priority columns: short informational columns that should get natural width
		if headerLower == "domain" || headerLower == "enabled" || headerLower == "imap" ||
			headerLower == "created" || headerLower == "updated" || headerLower == "id" {
			priorityColumns[i] = true
		}
		// Content columns: potentially long text that can be wrapped effectively
		if headerLower == "name" || headerLower == "description" || headerLower == "labels" {
			contentColumns[i] = true
		}
	}

	remainingWidth := availableWidth

	// First pass: allocate natural width to priority columns
	for i := range colWidths {
		if priorityColumns[i] {
			// Give priority columns their natural width, but cap at reasonable limits
			maxAllowed := remainingWidth / 3 // Don't let any single priority column take more than 1/3
			if colWidths[i] > maxAllowed {
				colWidths[i] = maxAllowed
			}
			remainingWidth -= colWidths[i]
		}
	}

	// Second pass: handle content columns vs other columns differently
	contentIndices := make([]int, 0)
	otherIndices := make([]int, 0)

	for i := range colWidths {
		if !priorityColumns[i] {
			if contentColumns[i] {
				contentIndices = append(contentIndices, i)
			} else {
				otherIndices = append(otherIndices, i)
			}
		}
	}

	// Allocate reasonable space for non-content, non-priority columns first
	for _, i := range otherIndices {
		reasonableWidth := originalWidths[i]
		// Cap at sensible maximums based on column type
		headerLower := strings.ToLower(headers[i])
		switch headerLower {
		case "recipients":
			// Recipients need enough space for email addresses
			reasonableWidth = minInt(reasonableWidth, 30)
		default:
			reasonableWidth = minInt(reasonableWidth, 20)
		}

		if remainingWidth >= reasonableWidth {
			colWidths[i] = reasonableWidth
			remainingWidth -= reasonableWidth
		} else {
			// Give it what we can, but at least 8 chars
			colWidths[i] = maxInt(8, remainingWidth/maxInt(1, len(otherIndices)))
			remainingWidth = maxInt(0, remainingWidth-colWidths[i])
		}
	}

	// Third pass: distribute remaining width to content columns (these wrap well)
	if len(contentIndices) > 0 && remainingWidth > 0 {
		totalContentWeight := 0
		for _, i := range contentIndices {
			// Use a weighted approach - longer content gets more space, but with diminishing returns
			weight := int(math.Sqrt(float64(originalWidths[i])))
			totalContentWeight += weight
		}

		for _, i := range contentIndices {
			if totalContentWeight > 0 {
				weight := int(math.Sqrt(float64(originalWidths[i])))
				colWidths[i] = (weight * remainingWidth) / totalContentWeight
			} else {
				colWidths[i] = remainingWidth / len(contentIndices)
			}

			// Ensure minimum width for content columns
			if colWidths[i] < 12 {
				colWidths[i] = 12
			}
		}
	}

	// Final pass: ensure all columns have minimum width
	for i := range colWidths {
		if colWidths[i] < 5 {
			colWidths[i] = 5
		}
	}

	return colWidths
}

// Helper functions for min/max
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// wrapCellContent wraps individual cell content to fit within the specified width
func (f *Formatter) wrapCellContent(content string, maxWidth int) string {
	if len(content) <= maxWidth {
		return content
	}

	// Implement proper text wrapping by splitting on word boundaries
	return wrapText(content, maxWidth)
}

// wrapText wraps text to the specified width, breaking on word boundaries when possible
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	// For very small widths, just truncate
	if width < 10 {
		return TruncateString(text, width)
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return TruncateString(text, width)
	}

	var lines []string
	var currentLine strings.Builder

	// Reserve space for continuation indicator on wrapped lines
	continuationIndicator := "┗ "
	continuationSpace := len(continuationIndicator)

	for _, word := range words {
		// Determine the effective width for this line
		effectiveWidth := width
		if len(lines) > 0 {
			// This will be a continuation line, so reserve space for the indicator
			effectiveWidth = width - continuationSpace
		}

		// If adding this word would exceed the effective width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > effectiveWidth {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			// Recalculate effective width for the new line
			effectiveWidth = width - continuationSpace
		}

		// If the word itself is longer than the effective width, we need to break it
		if len(word) > effectiveWidth {
			// Finish current line if it has content
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}

			// Break the long word across multiple lines
			for len(word) > effectiveWidth {
				lines = append(lines, word[:effectiveWidth])
				word = word[effectiveWidth:]
				effectiveWidth = width - continuationSpace // Subsequent lines need indicator space
			}
			if len(word) > 0 {
				currentLine.WriteString(word)
			}
		} else {
			// Add word to current line
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	// Add the last line if it has content
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	// Add visual indicators for continuation lines
	if len(lines) > 1 {
		for i := 1; i < len(lines); i++ {
			lines[i] = continuationIndicator + lines[i]
		}
	}

	// Join lines with newlines for multi-line cell content
	return strings.Join(lines, "\n")
}

// calculateOptimalWidthForColumn: removed (was unused helper for column sizing)

// convertToInterface converts a slice of strings to a slice of interface{}
func convertToInterface(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// getTerminalWidth returns the current terminal width or a default value
func getTerminalWidth() int {
	// Try to get terminal width from various sources
	if width := getTerminalWidthFromTerm(); width > 0 {
		return width
	}

	// Default width for standard terminals
	return 80
}

// getTerminalWidthFromTerm attempts to get terminal width using golang.org/x/term
func getTerminalWidthFromTerm() int {
	// Check if stdout is a terminal
	if term.IsTerminal(int(os.Stdout.Fd())) {
		width, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && width > 0 {
			return width
		}
	}

	// Check common environment variables as fallback
	if cols := getEnvInt("COLUMNS"); cols > 0 {
		return cols
	}

	// Check TERM-specific environment variables
	if cols := getEnvInt("COLS"); cols > 0 {
		return cols
	}

	return 0 // Let caller use default
}

// getEnvInt gets an integer value from an environment variable
func getEnvInt(key string) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil && intValue > 0 {
			return intValue
		}
	}
	return 0
}

// formatCSV outputs data as CSV
func (f *Formatter) formatCSV(data interface{}) error {
	switch v := data.(type) {
	case TableData:
		// Write headers
		_, _ = fmt.Fprintln(f.writer, strings.Join(v.Headers, ","))
		// Write rows
		for _, row := range v.Rows {
			_, _ = fmt.Fprintln(f.writer, strings.Join(row, ","))
		}
	case *TableData:
		// Write headers
		_, _ = fmt.Fprintln(f.writer, strings.Join(v.Headers, ","))
		// Write rows
		for _, row := range v.Rows {
			_, _ = fmt.Fprintln(f.writer, strings.Join(row, ","))
		}
	default:
		return fmt.Errorf("CSV format requires TableData struct")
	}
	return nil
}

// formatPlain outputs data as plain text with fixed-width columns
// No borders, no truncation, just aligned columns separated by 2 spaces
func (f *Formatter) formatPlain(data interface{}) error {
	switch v := data.(type) {
	case TableData:
		return f.formatPlainTable(&v)
	case *TableData:
		return f.formatPlainTable(v)
	default:
		return fmt.Errorf("plain format requires TableData struct")
	}
}

// formatPlainTable formats a TableData as plain text with fixed-width columns
func (f *Formatter) formatPlainTable(td *TableData) error {
	if len(td.Headers) == 0 {
		return nil
	}

	// Calculate column widths based on actual content (no truncation)
	colWidths := make([]int, len(td.Headers))

	// Initialize with header widths
	for i, header := range td.Headers {
		colWidths[i] = len(header)
	}

	// Update with row content widths
	for _, row := range td.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Format and write headers
	headerParts := make([]string, len(td.Headers))
	for i, header := range td.Headers {
		// Don't pad the last column to avoid trailing spaces
		if i == len(td.Headers)-1 {
			headerParts[i] = header
		} else {
			headerParts[i] = padRight(header, colWidths[i])
		}
	}
	_, _ = fmt.Fprintln(f.writer, strings.Join(headerParts, "  "))

	// Format and write rows
	for _, row := range td.Rows {
		rowParts := make([]string, len(row))
		for i, cell := range row {
			// Don't pad the last column to avoid trailing spaces
			if i == len(row)-1 {
				rowParts[i] = cell
			} else if i < len(colWidths) {
				rowParts[i] = padRight(cell, colWidths[i])
			} else {
				rowParts[i] = cell
			}
		}
		_, _ = fmt.Fprintln(f.writer, strings.Join(rowParts, "  "))
	}

	return nil
}

// padRight pads a string with spaces on the right to reach the desired width
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
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
			return yesLabel
		}
		return noLabel
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
	const unit = 1024 // Use 1024 for binary units (KiB, MiB, GiB, etc.)
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	units := []string{"K", "M", "G", "T", "P", "E"}
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
		if exp >= len(units) {
			break
		}
	}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	return fmt.Sprintf("%.1f %sB", float64(bytes)/float64(div), units[exp])
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
	case "plain":
		return FormatPlain, nil
	default:
		return "", fmt.Errorf("unsupported format: %s", s)
	}
}
