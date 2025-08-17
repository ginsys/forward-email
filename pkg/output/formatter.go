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

	// Get terminal width for column sizing
	terminalWidth := getTerminalWidth()

	switch v := data.(type) {
	case TableData:
		table.Header(convertToInterface(v.Headers)...)
		// Apply intelligent text wrapping for long content using current terminal width
		wrappedRows := f.wrapTableContentWithWidth(v.Rows, v.Headers, terminalWidth)
		for _, row := range wrappedRows {
			table.Append(convertToInterface(row)...)
		}
	case *TableData:
		table.Header(convertToInterface(v.Headers)...)
		// Apply intelligent text wrapping for long content using current terminal width
		wrappedRows := f.wrapTableContentWithWidth(v.Rows, v.Headers, terminalWidth)
		for _, row := range wrappedRows {
			table.Append(convertToInterface(row)...)
		}
	default:
		return fmt.Errorf("table format requires TableData struct, got %T", data)
	}

	return table.Render()
}

// wrapTableContent intelligently wraps long content in table cells (deprecated, use wrapTableContentWithWidth)
func (f *Formatter) wrapTableContent(rows [][]string, headers []string) [][]string {
	return f.wrapTableContentWithWidth(rows, headers, 80) // Default width
}

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
			reasonableWidth = min(reasonableWidth, 30)
		default:
			reasonableWidth = min(reasonableWidth, 20)
		}

		if remainingWidth >= reasonableWidth {
			colWidths[i] = reasonableWidth
			remainingWidth -= reasonableWidth
		} else {
			// Give it what we can, but at least 8 chars
			colWidths[i] = max(8, remainingWidth/max(1, len(otherIndices)))
			remainingWidth = max(0, remainingWidth-colWidths[i])
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
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
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

// calculateOptimalWidthForColumn determines the best width for a specific column
func calculateOptimalWidthForColumn(header string, rows [][]string, colIndex int) int {
	// Start with header length
	maxWidth := len(header)

	// Analyze content in this column
	for _, row := range rows {
		if colIndex < len(row) {
			cellLength := len(row[colIndex])
			if cellLength > maxWidth {
				maxWidth = cellLength
			}
		}
	}

	// Get terminal width to determine if we can be more generous with column widths
	terminalWidth := getTerminalWidth()

	// Calculate available space per column (rough estimate)
	// Account for borders, padding, and assume this column gets 1/numColumns of space
	numColumns := len(rows[0])
	if numColumns == 0 {
		numColumns = 1
	}

	// Rough calculation: terminal width minus borders and padding, divided by number of columns
	availablePerColumn := (terminalWidth - (numColumns * 4)) / numColumns // 4 chars for borders/padding per column

	// Apply content-type specific limits, but be more generous if terminal is wide
	switch {
	case strings.Contains(strings.ToLower(header), "recipient"):
		limit := 40
		if terminalWidth > 120 {
			limit = 60 // More generous on wide terminals
		}
		if maxWidth > limit && availablePerColumn < limit {
			maxWidth = limit
		}
	case strings.Contains(strings.ToLower(header), "label"):
		limit := 25
		if terminalWidth > 120 {
			limit = 40
		}
		if maxWidth > limit && availablePerColumn < limit {
			maxWidth = limit
		}
	case strings.Contains(strings.ToLower(header), "name"):
		limit := 20
		if terminalWidth > 120 {
			limit = 30
		}
		if maxWidth > limit && availablePerColumn < limit {
			maxWidth = limit
		}
	case strings.Contains(strings.ToLower(header), "domain"):
		limit := 25
		if terminalWidth > 120 {
			limit = 35
		}
		if maxWidth > limit && availablePerColumn < limit {
			maxWidth = limit
		}
	case strings.Contains(strings.ToLower(header), "description"):
		limit := 60
		if terminalWidth > 150 {
			limit = 100
		}
		if maxWidth > limit && availablePerColumn < limit {
			maxWidth = limit
		}
	default:
		// General content should not exceed 50 characters, but be more generous on wide terminals
		limit := 50
		if terminalWidth > 120 {
			limit = 80
		}
		if maxWidth > limit && availablePerColumn < limit {
			maxWidth = limit
		}
	}

	// Ensure minimum readability
	if maxWidth < 8 {
		maxWidth = 8
	}

	return maxWidth
}

// convertToInterface converts a slice of strings to a slice of interface{}
func convertToInterface(strings []string) []interface{} {
	result := make([]interface{}, len(strings))
	for i, s := range strings {
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
			return "Yes"
		}
		return "No"
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
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %sB", float64(bytes)/float64(div), []string{"K", "M", "G", "T", "P", "E"}[exp])
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
