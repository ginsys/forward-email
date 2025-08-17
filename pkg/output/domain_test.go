package output

import (
	"testing"
	"time"

	"github.com/ginsys/forwardemail-cli/pkg/api"
)

func TestFormatDomainList(t *testing.T) {
	domains := []api.Domain{
		{
			ID:         "domain1",
			Name:       "example.com",
			IsVerified: true,
			Plan:       "free",
			Members:    []api.DomainMember{{}, {}}, // 2 members
			CreatedAt:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:         "domain2",
			Name:       "test.org",
			IsVerified: false,
			Plan:       "team",
			Members:    []api.DomainMember{{}}, // 1 member
			CreatedAt:  time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	table, err := FormatDomainList(domains, FormatTable)
	if err != nil {
		t.Fatalf("FormatDomainList failed: %v", err)
	}

	expectedHeaders := []string{"NAME", "VERIFIED", "PLAN", "ALIASES", "MEMBERS", "CREATED"}
	if len(table.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(table.Headers))
	}

	for i, header := range expectedHeaders {
		if table.Headers[i] != header {
			t.Errorf("Expected header %s, got %s", header, table.Headers[i])
		}
	}

	if len(table.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(table.Rows))
	}

	// Check first row
	row1 := table.Rows[0]
	if row1[0] != "example.com" {
		t.Errorf("Expected domain name 'example.com', got %s", row1[0])
	}
	if row1[1] != "✓" {
		t.Errorf("Expected verified symbol '✓', got %s", row1[1])
	}
	if row1[2] != "free" {
		t.Errorf("Expected plan 'free', got %s", row1[2])
	}
	if row1[4] != "2" {
		t.Errorf("Expected 2 members, got %s", row1[4])
	}
	if row1[5] != "2023-01-01" {
		t.Errorf("Expected date '2023-01-01', got %s", row1[5])
	}

	// Check second row
	row2 := table.Rows[1]
	if row2[0] != "test.org" {
		t.Errorf("Expected domain name 'test.org', got %s", row2[0])
	}
	if row2[1] != "✗" {
		t.Errorf("Expected unverified symbol '✗', got %s", row2[1])
	}
	if row2[2] != "team" {
		t.Errorf("Expected plan 'team', got %s", row2[2])
	}
}

func TestFormatDomainDetails(t *testing.T) {
	domain := &api.Domain{
		ID:                    "test-id",
		Name:                  "example.com",
		IsVerified:            true,
		IsGlobal:              false,
		Plan:                  "enhanced_protection",
		HasMXRecord:           true,
		HasTXTRecord:          true,
		HasDMARCRecord:        false,
		HasSPFRecord:          true,
		HasDKIMRecord:         true,
		MaxForwardedAddresses: 50,
		RetentionDays:         30,
		VerificationRecord:    "very-long-verification-record-that-should-be-truncated-for-display",
		Members:               []api.DomainMember{{}, {}, {}}, // 3 members
		Invitations:           []api.DomainInvitation{{}},     // 1 invitation
		CreatedAt:             time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2023, 6, 1, 15, 30, 0, 0, time.UTC),
	}

	table, err := FormatDomainDetails(domain, FormatTable)
	if err != nil {
		t.Fatalf("FormatDomainDetails failed: %v", err)
	}

	expectedHeaders := []string{"PROPERTY", "VALUE"}
	if len(table.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(table.Headers))
	}

	// Find specific rows and check values
	rowMap := make(map[string]string)
	for _, row := range table.Rows {
		rowMap[row[0]] = row[1]
	}

	tests := []struct {
		property string
		expected string
	}{
		{"ID", "test-id"},
		{"Name", "example.com"},
		{"Verified", "✓"},
		{"Plan", "enhanced_protection"},
		{"Global", "✗"},
		{"MX Record", "✓"},
		{"TXT Record", "✓"},
		{"DMARC Record", "✗"},
		{"SPF Record", "✓"},
		{"DKIM Record", "✓"},
		{"Max Forwarded Addresses", "50"},
		{"Retention Days", "30"},
		{"Members", "3"},
		{"Pending Invitations", "1"},
	}

	for _, test := range tests {
		if value, exists := rowMap[test.property]; !exists {
			t.Errorf("Property '%s' not found", test.property)
		} else if value != test.expected {
			t.Errorf("Property '%s': expected '%s', got '%s'", test.property, test.expected, value)
		}
	}

	// Check that verification record is truncated
	if verificationRecord, exists := rowMap["Verification Record"]; exists {
		if len(verificationRecord) > 50 {
			t.Errorf("Verification record should be truncated to 50 characters, got %d", len(verificationRecord))
		}
		if verificationRecord[len(verificationRecord)-3:] != "..." {
			t.Errorf("Truncated verification record should end with '...', got %s", verificationRecord)
		}
	}
}

func TestFormatDNSRecords(t *testing.T) {
	records := []api.DNSRecord{
		{
			Type:     "MX",
			Name:     "@",
			Value:    "mx1.forwardemail.net",
			Priority: 10,
			TTL:      3600,
			Required: true,
			Purpose:  "Email forwarding",
		},
		{
			Type:     "TXT",
			Name:     "@",
			Value:    "v=spf1 include:spf.forwardemail.net ~all",
			Priority: 0,
			TTL:      0,
			Required: true,
			Purpose:  "SPF record",
		},
		{
			Type:     "CNAME",
			Name:     "www",
			Value:    "example.com",
			Required: false,
			Purpose:  "WWW redirect",
		},
	}

	table, err := FormatDNSRecords(records, FormatTable)
	if err != nil {
		t.Fatalf("FormatDNSRecords failed: %v", err)
	}

	expectedHeaders := []string{"TYPE", "NAME", "VALUE", "PRIORITY", "TTL", "REQUIRED", "PURPOSE"}
	if len(table.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(table.Headers))
	}

	if len(table.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(table.Rows))
	}

	// Check MX record
	mx := table.Rows[0]
	if mx[0] != "MX" {
		t.Errorf("Expected type 'MX', got %s", mx[0])
	}
	if mx[3] != "10" {
		t.Errorf("Expected priority '10', got %s", mx[3])
	}
	if mx[4] != "3600" {
		t.Errorf("Expected TTL '3600', got %s", mx[4])
	}
	if mx[5] != "✓" {
		t.Errorf("Expected required '✓', got %s", mx[5])
	}

	// Check TXT record (no priority/TTL)
	txt := table.Rows[1]
	if txt[0] != "TXT" {
		t.Errorf("Expected type 'TXT', got %s", txt[0])
	}
	if txt[3] != "-" {
		t.Errorf("Expected priority '-', got %s", txt[3])
	}
	if txt[4] != "-" {
		t.Errorf("Expected TTL '-', got %s", txt[4])
	}

	// Check CNAME record (not required)
	cname := table.Rows[2]
	if cname[0] != "CNAME" {
		t.Errorf("Expected type 'CNAME', got %s", cname[0])
	}
	if cname[5] != "✗" {
		t.Errorf("Expected not required '✗', got %s", cname[5])
	}
}

func TestFormatDomainVerification(t *testing.T) {
	verification := &api.DomainVerification{
		IsVerified: false,
		DNSRecords: []api.DNSRecord{
			{Type: "MX", Name: "@", Value: "mx1.forwardemail.net"},
			{Type: "TXT", Name: "@", Value: "v=spf1 include:spf.forwardemail.net ~all"},
		},
		MissingRecords: []api.DNSRecord{
			{Type: "DMARC", Name: "_dmarc", Value: "v=DMARC1; p=quarantine"},
		},
		LastCheckedAt:   time.Date(2023, 6, 1, 10, 0, 0, 0, time.UTC),
		VerificationURL: "https://example.com/verify",
		Errors:          []string{"DMARC record missing", "DNS propagation incomplete"},
	}

	table, err := FormatDomainVerification(verification, FormatTable)
	if err != nil {
		t.Fatalf("FormatDomainVerification failed: %v", err)
	}

	expectedHeaders := []string{"PROPERTY", "VALUE"}
	if len(table.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(table.Headers))
	}

	// Convert to map for easier testing
	rowMap := make(map[string]string)
	for _, row := range table.Rows {
		rowMap[row[0]] = row[1]
	}

	if rowMap["Verified"] != "✗" {
		t.Errorf("Expected verified '✗', got %s", rowMap["Verified"])
	}
	if rowMap["DNS Records Found"] != "2" {
		t.Errorf("Expected DNS records found '2', got %s", rowMap["DNS Records Found"])
	}
	if rowMap["Missing Records"] != "1" {
		t.Errorf("Expected missing records '1', got %s", rowMap["Missing Records"])
	}
	if rowMap["Verification URL"] != "https://example.com/verify" {
		t.Errorf("Expected verification URL 'https://example.com/verify', got %s", rowMap["Verification URL"])
	}
	if rowMap["Errors"] != "2" {
		t.Errorf("Expected errors '2', got %s", rowMap["Errors"])
	}

	// Check that error details are included
	errorCount := 0
	for property := range rowMap {
		if property == "Error 1" || property == "Error 2" {
			errorCount++
		}
	}
	if errorCount != 2 {
		t.Errorf("Expected 2 error detail rows, got %d", errorCount)
	}
}

func TestFormatDomainQuota(t *testing.T) {
	quota := &api.DomainQuota{
		StorageUsed:      1024 * 1024 * 500,  // 500MB
		StorageLimit:     1024 * 1024 * 1024, // 1GB
		AliasesUsed:      15,
		AliasesLimit:     25,
		ForwardingUsed:   8,
		ForwardingLimit:  10,
		BandwidthUsed:    1024 * 1024 * 100, // 100MB
		BandwidthLimit:   1024 * 1024 * 200, // 200MB
		EmailsSentToday:  75,
		EmailsLimitDaily: 100,
	}

	table, err := FormatDomainQuota(quota, FormatTable)
	if err != nil {
		t.Fatalf("FormatDomainQuota failed: %v", err)
	}

	expectedHeaders := []string{"RESOURCE", "USED", "LIMIT", "PERCENTAGE"}
	if len(table.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(table.Headers))
	}

	if len(table.Rows) != 5 {
		t.Errorf("Expected 5 rows, got %d", len(table.Rows))
	}

	// Check storage row
	storage := table.Rows[0]
	if storage[0] != "Storage" {
		t.Errorf("Expected resource 'Storage', got %s", storage[0])
	}
	if storage[1] != "500.0 MB" {
		t.Errorf("Expected used '500.0 MB', got %s", storage[1])
	}
	if storage[2] != "1.0 GB" {
		t.Errorf("Expected limit '1.0 GB', got %s", storage[2])
	}
	if storage[3] != "48.8%" {
		t.Errorf("Expected percentage '48.8%%', got %s", storage[3])
	}

	// Check aliases row
	aliases := table.Rows[1]
	if aliases[0] != "Aliases" {
		t.Errorf("Expected resource 'Aliases', got %s", aliases[0])
	}
	if aliases[1] != "15" {
		t.Errorf("Expected used '15', got %s", aliases[1])
	}
	if aliases[3] != "60.0%" {
		t.Errorf("Expected percentage '60.0%%', got %s", aliases[3])
	}

	// Check daily emails row
	emails := table.Rows[4]
	if emails[0] != "Daily Emails" {
		t.Errorf("Expected resource 'Daily Emails', got %s", emails[0])
	}
	if emails[3] != "75.0%" {
		t.Errorf("Expected percentage '75.0%%', got %s", emails[3])
	}
}

func TestFormatDomainStats(t *testing.T) {
	stats := &api.DomainStats{
		TotalAliases:   20,
		ActiveAliases:  18,
		TotalMembers:   5,
		EmailsSent:     1500,
		EmailsReceived: 2000,
		LastActivityAt: time.Date(2023, 6, 1, 14, 30, 0, 0, time.UTC),
		CreatedAt:      time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	table, err := FormatDomainStats(stats, FormatTable)
	if err != nil {
		t.Fatalf("FormatDomainStats failed: %v", err)
	}

	expectedHeaders := []string{"METRIC", "VALUE"}
	if len(table.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(table.Headers))
	}

	// Convert to map for easier testing
	rowMap := make(map[string]string)
	for _, row := range table.Rows {
		rowMap[row[0]] = row[1]
	}

	tests := []struct {
		metric   string
		expected string
	}{
		{"Total Aliases", "20"},
		{"Active Aliases", "18"},
		{"Total Members", "5"},
		{"Emails Sent", "1500"},
		{"Emails Received", "2000"},
	}

	for _, test := range tests {
		if value, exists := rowMap[test.metric]; !exists {
			t.Errorf("Metric '%s' not found", test.metric)
		} else if value != test.expected {
			t.Errorf("Metric '%s': expected '%s', got '%s'", test.metric, test.expected, value)
		}
	}
}

func TestFormatDomainMembers(t *testing.T) {
	members := []api.DomainMember{
		{
			ID:    "member1",
			Group: "admin",
			User: api.User{
				ID:          "user1",
				Email:       "admin@example.com",
				DisplayName: "Admin User",
				GivenName:   "Admin",
				FamilyName:  "User",
			},
			JoinedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:    "member2",
			Group: "user",
			User: api.User{
				ID:         "user2",
				Email:      "user@example.com",
				GivenName:  "Regular",
				FamilyName: "User",
			},
			JoinedAt: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:    "member3",
			Group: "user",
			User: api.User{
				ID:    "user3",
				Email: "simple@example.com",
			},
			JoinedAt: time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	table, err := FormatDomainMembers(members, FormatTable)
	if err != nil {
		t.Fatalf("FormatDomainMembers failed: %v", err)
	}

	expectedHeaders := []string{"ID", "EMAIL", "NAME", "GROUP", "JOINED"}
	if len(table.Headers) != len(expectedHeaders) {
		t.Errorf("Expected %d headers, got %d", len(expectedHeaders), len(table.Headers))
	}

	if len(table.Rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(table.Rows))
	}

	// Check first member (with display name)
	member1 := table.Rows[0]
	if member1[0] != "member1" {
		t.Errorf("Expected ID 'member1', got %s", member1[0])
	}
	if member1[1] != "admin@example.com" {
		t.Errorf("Expected email 'admin@example.com', got %s", member1[1])
	}
	if member1[2] != "Admin User" {
		t.Errorf("Expected name 'Admin User', got %s", member1[2])
	}
	if member1[3] != "admin" {
		t.Errorf("Expected group 'admin', got %s", member1[3])
	}
	if member1[4] != "2023-01-01" {
		t.Errorf("Expected joined '2023-01-01', got %s", member1[4])
	}

	// Check second member (with given/family name)
	member2 := table.Rows[1]
	if member2[2] != "Regular User" {
		t.Errorf("Expected name 'Regular User', got %s", member2[2])
	}

	// Check third member (no name)
	member3 := table.Rows[2]
	if member3[2] != "-" {
		t.Errorf("Expected name '-', got %s", member3[2])
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, "-"},
		{"", "-"},
		{"hello", "hello"},
		{true, "✓"},
		{false, "✗"},
		{42, "42"},
		{int64(100), "100"},
		{3.14, "3.14"},
		{float32(2.5), "2.50"},
	}

	for _, test := range tests {
		result := FormatValue(test.input)
		if result != test.expected {
			t.Errorf("FormatValue(%v): expected '%s', got '%s'", test.input, test.expected, result)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly ten!", 12, "exactly ten!"},
		{"this is a very long string", 10, "this is..."},
		{"short", 3, "sho"},
		{"ab", 2, "ab"},
		{"a", 1, "a"},
	}

	for _, test := range tests {
		result := TruncateString(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("TruncateString('%s', %d): expected '%s', got '%s'",
				test.input, test.maxLen, test.expected, result)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, test := range tests {
		result := FormatBytes(test.input)
		if result != test.expected {
			t.Errorf("FormatBytes(%d): expected '%s', got '%s'",
				test.input, test.expected, result)
		}
	}
}

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		value    int64
		total    int64
		expected string
	}{
		{0, 100, "0.0%"},
		{25, 100, "25.0%"},
		{50, 100, "50.0%"},
		{75, 100, "75.0%"},
		{100, 100, "100.0%"},
		{1, 3, "33.3%"},
		{2, 3, "66.7%"},
		{5, 0, "0%"}, // Division by zero case
	}

	for _, test := range tests {
		result := FormatPercentage(test.value, test.total)
		if result != test.expected {
			t.Errorf("FormatPercentage(%d, %d): expected '%s', got '%s'",
				test.value, test.total, test.expected, result)
		}
	}
}
