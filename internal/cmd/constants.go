package cmd

// CLI command constants used throughout the application.
// These provide consistent defaults and string literals for command processing.
const (
	yesStr           = "yes"                                                                                     // Standard affirmative response for user prompts
	defaultProfile   = "default"                                                                                 // Default profile name for first-time users
	outputTable      = "table"                                                                                   // Default output format for CLI responses
	sortAsc          = "asc"                                                                                     // Default sort order (ascending)
	sortFieldName    = "name"                                                                                    // Default sort field for list operations
	sourceOSKeyring  = "OS keyring"                                                                              // Display label for OS keyring credential source
	sourceConfigFile = "config file"                                                                             // Display label for config file credential source
	validColumnsList = "name, domain, recipients, enabled, imap, labels, created, updated, description, pgp, id" // Valid column names for table output
)
