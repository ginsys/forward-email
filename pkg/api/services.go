package api

// AccountService handles account operations
type AccountService struct {
	client *Client
}

// DomainService handles domain operations
type DomainService struct {
	client *Client
}

// AliasService handles alias operations
type AliasService struct {
	client *Client
}

// EmailService handles email operations
type EmailService struct {
	client *Client
}

// LogService handles log operations
type LogService struct {
	client *Client
}

// CryptoService handles encryption operations
type CryptoService struct {
	client *Client
}
