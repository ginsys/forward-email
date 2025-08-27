package cmd

import (
	"testing"

	"github.com/ginsys/forward-email/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateEmailRequest_SecurityFix(t *testing.T) {
	tests := []struct {
		name        string
		req         *api.SendEmailRequest
		wantErr     bool
		errContains string
	}{
		{
			name: "valid email addresses should pass",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr: false,
		},
		{
			name: "valid email with display name should pass",
			req: &api.SendEmailRequest{
				From:    "John Doe <john@example.com>",
				To:      []string{"Jane Smith <jane@example.com>"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr: false,
		},
		{
			name: "multiple valid recipients should pass",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient1@example.com", "recipient2@example.com"},
				CC:      []string{"cc@example.com"},
				BCC:     []string{"bcc@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr: false,
		},
		{
			name: "invalid from email - no @ symbol",
			req: &api.SendEmailRequest{
				From:    "invalid-email",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "invalid from email - multiple @ symbols (vulnerability test)",
			req: &api.SendEmailRequest{
				From:    "@@@@",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "invalid from email - just @ symbol (vulnerability test)",
			req: &api.SendEmailRequest{
				From:    "@",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "invalid to email - no domain",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "invalid CC email - missing @",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				CC:      []string{"invalid-cc-email"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "invalid BCC email - malformed",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				BCC:     []string{"@invalid.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "security test - potential injection attempt",
			req: &api.SendEmailRequest{
				From:    "'; DROP TABLE users; --@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "security test - XSS attempt",
			req: &api.SendEmailRequest{
				From:    "<script>alert('xss')</script>@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "invalid email address",
		},
		{
			name: "empty from address",
			req: &api.SendEmailRequest{
				From:    "",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "from address is required",
		},
		{
			name: "empty to addresses",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "at least one recipient is required",
		},
		{
			name: "empty subject",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "",
				Text:    "Test message",
			},
			wantErr:     true,
			errContains: "subject is required",
		},
		{
			name: "empty text and HTML content",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com"},
				Subject: "Test Email",
				Text:    "",
				HTML:    "",
			},
			wantErr:     true,
			errContains: "either text or HTML content is required",
		},
		{
			name: "empty strings in address list should be ignored",
			req: &api.SendEmailRequest{
				From:    "sender@example.com",
				To:      []string{"recipient@example.com", ""},
				CC:      []string{"", "cc@example.com", ""},
				Subject: "Test Email",
				Text:    "Test message",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmailRequest(tt.req)

			if tt.wantErr {
				require.Error(t, err, "Expected validation to fail for test case: %s", tt.name)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains,
						"Error message should contain expected text for test case: %s", tt.name)
				}
			} else {
				assert.NoError(t, err, "Expected validation to pass for test case: %s", tt.name)
			}
		})
	}
}

func TestValidateEmailRequest_EmptyAddressHandling(t *testing.T) {
	// Test that empty strings in address arrays don't cause issues
	req := &api.SendEmailRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com", "", "another@example.com"},
		CC:      []string{"", "cc@example.com"},
		BCC:     []string{"bcc@example.com", ""},
		Subject: "Test Email",
		Text:    "Test message",
	}

	err := validateEmailRequest(req)
	assert.NoError(t, err, "Empty strings in address arrays should be ignored")
}

// TestValidateEmailRequest_RFC5322Compliance tests compliance with email address format standards
func TestValidateEmailRequest_RFC5322Compliance(t *testing.T) {
	validEmails := []string{
		"user@domain.com",
		"user.name@domain.com",
		"user+tag@domain.com",
		"user@subdomain.domain.com",
		"123@domain.com",
		"user@domain-name.com",
		"\"quoted user\"@domain.com",
		"John Doe <john@domain.com>",
		"user@domain.co.uk",
		"user@127.0.0.1",
		"user@domain", // Valid according to RFC for internal networks
	}

	invalidEmails := []string{
		"@",
		"@@",
		"@domain.com",
		"user@",
		"user@@domain.com",
		"user..name@domain.com",
		".user@domain.com",
		"user.@domain.com",
		"user name@domain.com",
		"user@domain.com.",
		"user@.domain.com",
		"user@domain..com",
		"@@@",
		"user@domain@com",
		"<>@domain.com",
		"user@[invalid ip]",
	}

	// Test valid emails
	for _, email := range validEmails {
		t.Run("valid_"+email, func(t *testing.T) {
			req := &api.SendEmailRequest{
				From:    email,
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Text:    "Test",
			}
			err := validateEmailRequest(req)
			assert.NoError(t, err, "Email should be valid: %s", email)
		})
	}

	// Test invalid emails
	for _, email := range invalidEmails {
		t.Run("invalid_"+email, func(t *testing.T) {
			req := &api.SendEmailRequest{
				From:    email,
				To:      []string{"recipient@example.com"},
				Subject: "Test",
				Text:    "Test",
			}
			err := validateEmailRequest(req)
			assert.Error(t, err, "Email should be invalid: %s", email)
			assert.Contains(t, err.Error(), "invalid email address",
				"Error should mention invalid email address for: %s", email)
		})
	}

	// Test empty email separately since it triggers different validation
	t.Run("empty_email_from_field", func(t *testing.T) {
		req := &api.SendEmailRequest{
			From:    "",
			To:      []string{"recipient@example.com"},
			Subject: "Test",
			Text:    "Test",
		}
		err := validateEmailRequest(req)
		assert.Error(t, err, "Empty from should be invalid")
		assert.Contains(t, err.Error(), "from address is required",
			"Error should mention from address is required")
	})
}
