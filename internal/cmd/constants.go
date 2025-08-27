package cmd

// CLI command constants used throughout the application.
// These provide consistent defaults and string literals for command processing.
const (
	// Standard affirmative response for user prompts
	yesStr = "yes"
	// Default profile name for first-time users
	defaultProfile = "default"
	// Default output format for CLI responses
	outputTable = "table"
	// Default sort order (ascending)
	sortAsc = "asc"
	// Default sort field for list operations
	sortFieldName = "name"
	// Display label for OS keyring credential source
	sourceOSKeyring = "OS keyring"
	// Display label for config file credential source
	sourceConfigFile = "config file"
	// Valid column names for table output
	validColumnsList = "name, domain, recipients, enabled, imap, labels, created, updated, description, pgp, id"
)
