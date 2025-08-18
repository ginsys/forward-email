package output

import (
	"strings"
	"testing"
	"time"

    "github.com/ginsys/forward-email/pkg/api"
)

func TestFormatAliasList(t *testing.T) {
	// Create test aliases
	aliases := []api.Alias{
		{
			ID:          "alias1",
			DomainID:    "example.com",
			Name:        "sales",
			IsEnabled:   true,
			Recipients:  []string{"sales@company.com", "backup@company.com"},
			Labels:      []string{"business", "important"},
			Description: "Sales inquiries",
			HasIMAP:     false,
			HasPGP:      false,
			HasPassword: false,
			CreatedAt:   time.Date(2023, 1, 15, 10, 30, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2023, 1, 20, 14, 45, 0, 0, time.UTC),
		},
		{
			ID:          "alias2",
			DomainID:    "example.com",
			Name:        "support",
			IsEnabled:   false,
			Recipients:  []string{"support@company.com"},
			Labels:      []string{"customer-service"},
			Description: "Customer support",
			HasIMAP:     true,
			HasPGP:      true,
			HasPassword: true,
			CreatedAt:   time.Date(2023, 1, 10, 9, 15, 0, 0, time.UTC),
			UpdatedAt:   time.Date(2023, 1, 18, 11, 20, 0, 0, time.UTC),
		},
		{
			ID:          "alias3",
			DomainID:    "test.com",
			Name:        "very-long-alias-name-that-should-be-truncated-in-table-view",
			IsEnabled:   true,
			Recipients:  []string{"user1@company.com", "user2@company.com", "user3@company.com", "user4@company.com", "user5@company.com"},
			Labels:      []string{"team", "development", "testing", "quality-assurance", "operations", "management"},
			Description: "Very long description that should be handled properly",
			HasIMAP:     false,
			HasPGP:      false,
			HasPassword: false,
			CreatedAt:   time.Time{}, // Zero time
			UpdatedAt:   time.Time{}, // Zero time
		},
	}

	tests := []struct {
		name            string
		format          Format
		expectError     bool
		expectedRows    int
		expectedFields  []string
		checkTruncation bool
	}{
		{
			name:            "table format",
			format:          FormatTable,
			expectError:     false,
			expectedRows:    3,
			expectedFields:  []string{"sales", "support", "example.com", "business", "customer-service"},
			checkTruncation: true,
		},
		{
			name:            "CSV format",
			format:          FormatCSV,
			expectError:     false,
			expectedRows:    3,
			expectedFields:  []string{"sales", "support", "example.com", "business", "customer-service"},
			checkTruncation: false,
		},
		{
			name:        "JSON format should fail",
			format:      FormatJSON,
			expectError: true,
		},
		{
			name:        "YAML format should fail",
			format:      FormatYAML,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatAliasList(aliases, tt.format, "example.com")

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result == nil {
					t.Fatal("Expected result but got nil")
				}

				// Check number of rows (excluding header)
				if len(result.Rows) != tt.expectedRows {
					t.Errorf("Expected %d rows, got %d", tt.expectedRows, len(result.Rows))
				}

				// Check headers
				expectedHeaders := []string{"NAME", "DOMAIN", "RECIPIENTS", "ENABLED", "IMAP", "LABELS", "CREATED"}
				if len(result.Headers) != len(expectedHeaders) {
					t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(result.Headers))
				}
				for i, expected := range expectedHeaders {
					if i < len(result.Headers) && result.Headers[i] != expected {
						t.Errorf("Expected header %d to be '%s', got '%s'", i, expected, result.Headers[i])
					}
				}

				// Check that all expected fields are present
				allData := strings.Join(result.Headers, " ")
				for _, row := range result.Rows {
					allData += " " + strings.Join(row, " ")
				}

				for _, field := range tt.expectedFields {
					if !strings.Contains(allData, field) {
						t.Errorf("Expected field '%s' not found in output", field)
					}
				}

				// Check enabled/disabled formatting
				if !strings.Contains(allData, "Yes") || !strings.Contains(allData, "No") {
					t.Error("Expected boolean values to be formatted as Yes/No")
				}

				// Check zero date handling
				if strings.Contains(allData, "0001-01-01") {
					t.Error("Zero dates should be displayed as '-', not '0001-01-01'")
				}
				if !strings.Contains(allData, "-") {
					t.Error("Expected zero dates to be displayed as '-'")
				}

				// Test that very long content will be handled by formatter's text wrapping
				if tt.checkTruncation {
					// The raw table data may contain long content - text wrapping happens at format time
					// Just ensure we have the long test content that will trigger wrapping
					hasLongContent := false
					for _, row := range result.Rows {
						for _, cell := range row {
							if len(cell) > 60 { // Our test data has very long content
								hasLongContent = true
							}
						}
					}
					if !hasLongContent {
						t.Error("Expected to have long content in test data that will trigger text wrapping")
					}

					// Test actual formatting with text wrapping by running it through the formatter
					var buf strings.Builder
					formatter := NewFormatter(tt.format, &buf)
					err := formatter.Format(result)
					if err != nil {
						t.Errorf("Failed to format table: %v", err)
					}

					// Check that the formatted output contains wrapped content with continuation indicators
					formattedOutput := buf.String()
					if !strings.Contains(formattedOutput, "┗") {
						t.Error("Expected formatted output to contain wrapped content with continuation indicators (┗)")
					}
				} else {
					// For CSV format, full content should be preserved
					foundFullRecipients := false
					for _, row := range result.Rows {
						if len(row) > 2 { // Check recipients column
							recipients := row[2]
							if strings.Contains(recipients, "user1@company.com") &&
								strings.Contains(recipients, "user5@company.com") {
								foundFullRecipients = true
								break
							}
						}
					}
					if !foundFullRecipients {
						t.Error("CSV format should preserve full content without truncation")
					}
				}
			}
		})
	}
}

func TestFormatAliasDetails(t *testing.T) {
	// Create test alias with comprehensive data
	alias := &api.Alias{
		ID:          "detail-alias-id",
		DomainID:    "example.com",
		Name:        "detailed-alias",
		IsEnabled:   true,
		Recipients:  []string{"user1@company.com", "user2@company.com", "user3@company.com"},
		Labels:      []string{"important", "team", "production"},
		Description: "Detailed test alias with all fields populated",
		HasIMAP:     true,
		HasPGP:      true,
		HasPassword: true,
		PublicKey:   "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: GnuPG v2\n\nmQENBF...", // Truncated for test
		Quota: &api.AliasQuota{
			StorageUsed:  1024 * 1024 * 75,  // 75MB
			StorageLimit: 1024 * 1024 * 500, // 500MB
			EmailsSent:   25,
			EmailsLimit:  100,
		},
		Vacation: &api.VacationResponder{
			IsEnabled: true,
			Subject:   "Out of office",
			Message:   "I am currently out of office.",
			StartDate: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		CreatedAt: time.Date(2023, 1, 10, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2023, 1, 20, 15, 45, 0, 0, time.UTC),
	}

	tests := []struct {
		name         string
		format       Format
		expectError  bool
		expectedKeys []string
	}{
		{
			name:        "table format",
			format:      FormatTable,
			expectError: false,
			expectedKeys: []string{"ID", "Name", "Domain ID", "Enabled", "Recipients", "Labels",
				"Description", "IMAP Enabled", "PGP Enabled", "Has Password", "Storage Used",
				"Storage Limit", "Emails Sent Today", "Vacation Enabled"},
		},
		{
			name:         "CSV format",
			format:       FormatCSV,
			expectError:  false,
			expectedKeys: []string{"ID", "Name", "Domain ID", "Enabled"},
		},
		{
			name:        "JSON format should fail",
			format:      FormatJSON,
			expectError: true,
		},
		{
			name:        "YAML format should fail",
			format:      FormatYAML,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatAliasDetails(alias, tt.format)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result == nil {
					t.Fatal("Expected result but got nil")
				}

				// Check headers
				expectedHeaders := []string{"PROPERTY", "VALUE"}
				if len(result.Headers) != len(expectedHeaders) {
					t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(result.Headers))
				}

				// Convert all data to string for searching
				allData := strings.Join(result.Headers, " ")
				for _, row := range result.Rows {
					allData += " " + strings.Join(row, " ")
				}

				// Check that expected keys are present
				for _, key := range tt.expectedKeys {
					if !strings.Contains(allData, key) {
						t.Errorf("Expected key '%s' not found in output", key)
					}
				}

				// Check specific values
				if !strings.Contains(allData, "detailed-alias") {
					t.Error("Expected alias name 'detailed-alias' not found")
				}
				if !strings.Contains(allData, "example.com") {
					t.Error("Expected domain 'example.com' not found")
				}
				if !strings.Contains(allData, "user1@company.com") {
					t.Error("Expected recipient 'user1@company.com' not found")
				}
				if !strings.Contains(allData, "important") {
					t.Error("Expected label 'important' not found")
				}

				// Check boolean formatting
				if !strings.Contains(allData, "Yes") {
					t.Error("Expected boolean values to be formatted as 'Yes'")
				}

				// Check quota formatting (should have MB formatting)
				if !strings.Contains(allData, "MB") {
					t.Error("Expected storage quota to be formatted with MB units")
				}

				// Check vacation responder
				if !strings.Contains(allData, "Out of office") {
					t.Error("Expected vacation subject 'Out of office' not found")
				}

				// Check that PGP key is truncated if too long
				foundLongKey := false
				for _, row := range result.Rows {
					if len(row) > 1 && strings.Contains(row[0], "PGP Public Key") {
						if len(row[1]) > 150 { // Should be truncated
							foundLongKey = true
						}
					}
				}
				if foundLongKey {
					t.Error("PGP public key should be truncated for display")
				}
			}
		})
	}
}

func TestFormatAliasDetails_MinimalData(t *testing.T) {
	// Test with minimal alias data (no optional fields)
	alias := &api.Alias{
		ID:          "minimal-alias",
		DomainID:    "test.com",
		Name:        "minimal",
		IsEnabled:   false,
		Recipients:  []string{"minimal@test.com"},
		HasIMAP:     false,
		HasPGP:      false,
		HasPassword: false,
		CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	result, err := FormatAliasDetails(alias, FormatTable)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Convert to string for easy searching
	allData := ""
	for _, row := range result.Rows {
		allData += strings.Join(row, " ") + " "
	}

	// Check basic fields are present
	if !strings.Contains(allData, "minimal-alias") {
		t.Error("Expected alias ID 'minimal-alias' not found")
	}
	if !strings.Contains(allData, "minimal") {
		t.Error("Expected alias name 'minimal' not found")
	}
	if !strings.Contains(allData, "test.com") {
		t.Error("Expected domain 'test.com' not found")
	}
	if !strings.Contains(allData, "No") {
		t.Error("Expected 'No' for disabled fields")
	}

	// Check that optional fields are not present when empty
	if strings.Contains(allData, "Labels") {
		t.Error("Labels should not be shown when empty")
	}
	if strings.Contains(allData, "Description") {
		t.Error("Description should not be shown when empty")
	}
	if strings.Contains(allData, "PGP Public Key") {
		t.Error("PGP Public Key should not be shown when empty")
	}
	if strings.Contains(allData, "Vacation") {
		t.Error("Vacation should not be shown when not configured")
	}
	if strings.Contains(allData, "Storage") {
		t.Error("Storage quota should not be shown when not present")
	}
}

func TestFormatAliasQuota(t *testing.T) {
	quota := &api.AliasQuota{
		StorageUsed:  1024 * 1024 * 150,  // 150MB
		StorageLimit: 1024 * 1024 * 1024, // 1GB
		EmailsSent:   45,
		EmailsLimit:  100,
	}

	tests := []struct {
		name        string
		format      Format
		expectError bool
	}{
		{
			name:        "table format",
			format:      FormatTable,
			expectError: false,
		},
		{
			name:        "CSV format",
			format:      FormatCSV,
			expectError: false,
		},
		{
			name:        "JSON format should fail",
			format:      FormatJSON,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatAliasQuota(quota, tt.format)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result == nil {
					t.Fatal("Expected result but got nil")
				}

				// Check headers
				expectedHeaders := []string{"METRIC", "USED", "LIMIT", "PERCENTAGE"}
				if len(result.Headers) != len(expectedHeaders) {
					t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(result.Headers))
				}

				// Should have 2 rows: Storage and Emails
				if len(result.Rows) != 2 {
					t.Errorf("Expected 2 rows, got %d", len(result.Rows))
				}

				// Convert to string for searching
				allData := ""
				for _, row := range result.Rows {
					allData += strings.Join(row, " ") + " "
				}

				// Check for storage formatting
				if !strings.Contains(allData, "Storage") {
					t.Error("Expected 'Storage' metric not found")
				}
				if !strings.Contains(allData, "150.0 MB") {
					t.Error("Expected storage used '150.0 MB' not found")
				}
				if !strings.Contains(allData, "1.0 GB") {
					t.Error("Expected storage limit '1.0 GB' not found")
				}

				// Check for email formatting
				if !strings.Contains(allData, "Emails") {
					t.Error("Expected 'Emails' metric not found")
				}
				if !strings.Contains(allData, "45") {
					t.Error("Expected emails sent '45' not found")
				}
				if !strings.Contains(allData, "100") {
					t.Error("Expected emails limit '100' not found")
				}

				// Check for percentage formatting
				if !strings.Contains(allData, "%") {
					t.Error("Expected percentage values not found")
				}
			}
		})
	}
}

func TestFormatAliasStats(t *testing.T) {
	stats := &api.AliasStats{
		EmailsReceived: 1250,
		EmailsSent:     875,
		StorageUsed:    1024 * 1024 * 256, // 256MB
		LastActivity:   time.Date(2023, 6, 15, 14, 30, 0, 0, time.UTC),
		RecentSenders:  []string{"client1@company.com", "client2@company.com", "partner@external.com"},
	}

	tests := []struct {
		name        string
		format      Format
		expectError bool
	}{
		{
			name:        "table format",
			format:      FormatTable,
			expectError: false,
		},
		{
			name:        "CSV format",
			format:      FormatCSV,
			expectError: false,
		},
		{
			name:        "JSON format should fail",
			format:      FormatJSON,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatAliasStats(stats, tt.format)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result == nil {
					t.Fatal("Expected result but got nil")
				}

				// Check headers
				expectedHeaders := []string{"STATISTIC", "VALUE"}
				if len(result.Headers) != len(expectedHeaders) {
					t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(result.Headers))
				}

				// Convert to string for searching
				allData := ""
				for _, row := range result.Rows {
					allData += strings.Join(row, " ") + " "
				}

				// Check for expected statistics
				if !strings.Contains(allData, "Emails Received") {
					t.Error("Expected 'Emails Received' statistic not found")
				}
				if !strings.Contains(allData, "1250") {
					t.Error("Expected emails received count '1250' not found")
				}

				if !strings.Contains(allData, "Emails Sent") {
					t.Error("Expected 'Emails Sent' statistic not found")
				}
				if !strings.Contains(allData, "875") {
					t.Error("Expected emails sent count '875' not found")
				}

				if !strings.Contains(allData, "Storage Used") {
					t.Error("Expected 'Storage Used' statistic not found")
				}
				if !strings.Contains(allData, "256.0 MB") {
					t.Error("Expected storage used '256.0 MB' not found")
				}

				if !strings.Contains(allData, "Last Activity") {
					t.Error("Expected 'Last Activity' statistic not found")
				}
				if !strings.Contains(allData, "2023-06-15") {
					t.Error("Expected last activity date not found")
				}

				if !strings.Contains(allData, "Recent Senders") {
					t.Error("Expected 'Recent Senders' statistic not found")
				}
				if !strings.Contains(allData, "client1@company.com") {
					t.Error("Expected recent sender 'client1@company.com' not found")
				}
			}
		})
	}
}

func TestFormatAliasStats_ZeroTime(t *testing.T) {
	// Test with zero time (no last activity)
	stats := &api.AliasStats{
		EmailsReceived: 0,
		EmailsSent:     0,
		StorageUsed:    0,
		LastActivity:   time.Time{}, // Zero time
		RecentSenders:  []string{},
	}

	result, err := FormatAliasStats(stats, FormatTable)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Convert to string for searching
	allData := ""
	for _, row := range result.Rows {
		allData += strings.Join(row, " ") + " "
	}

	// Should still have basic statistics
	if !strings.Contains(allData, "Emails Received") {
		t.Error("Expected 'Emails Received' statistic not found")
	}
	if !strings.Contains(allData, "0") {
		t.Error("Expected zero values not found")
	}

	// Should not have Last Activity when time is zero
	if strings.Contains(allData, "Last Activity") {
		t.Error("Last Activity should not be shown when time is zero")
	}

	// Should not have Recent Senders when empty
	if strings.Contains(allData, "Recent Senders") {
		t.Error("Recent Senders should not be shown when empty")
	}
}

func TestFormatAliasRecipients(t *testing.T) {
	recipients := []string{
		"user@example.com",
		"webhook://hooks.slack.com/services/T123/B456/xyz",
		"192.168.1.100",
		"mail.company.com",
		"support@company.co.uk",
	}

	tests := []struct {
		name        string
		format      Format
		expectError bool
	}{
		{
			name:        "table format",
			format:      FormatTable,
			expectError: false,
		},
		{
			name:        "CSV format",
			format:      FormatCSV,
			expectError: false,
		},
		{
			name:        "JSON format should fail",
			format:      FormatJSON,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatAliasRecipients(recipients, tt.format)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				if result == nil {
					t.Fatal("Expected result but got nil")
				}

				// Check headers
				expectedHeaders := []string{"RECIPIENT", "TYPE"}
				if len(result.Headers) != len(expectedHeaders) {
					t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(result.Headers))
				}

				// Should have 5 rows for 5 recipients
				if len(result.Rows) != 5 {
					t.Errorf("Expected 5 rows, got %d", len(result.Rows))
				}

				// Convert to string for searching
				allData := ""
				for _, row := range result.Rows {
					allData += strings.Join(row, " ") + " "
				}

				// Check that all recipients are present
				for _, recipient := range recipients {
					if !strings.Contains(allData, recipient) {
						t.Errorf("Expected recipient '%s' not found", recipient)
					}
				}

				// Check type detection
				if !strings.Contains(allData, "Email") {
					t.Error("Expected 'Email' type not found")
				}
				if !strings.Contains(allData, "Webhook") {
					t.Error("Expected 'Webhook' type not found")
				}
				if !strings.Contains(allData, "FQDN/IP") {
					t.Error("Expected 'FQDN/IP' type not found")
				}
			}
		})
	}
}
