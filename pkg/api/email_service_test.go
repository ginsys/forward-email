package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ginsys/forward-email/pkg/auth"
)

func newTestClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	u, _ := url.Parse(srv.URL)
	c := &Client{HTTPClient: srv.Client(), BaseURL: u, Auth: auth.MockProvider("test-key"), UserAgent: "test"}
	c.Emails = &EmailService{client: c}
	return c
}

func TestEmail_SendEmail_Validation(t *testing.T) {
	c := newTestClient(t, http.NotFoundHandler())
	_, err := c.Emails.SendEmail(context.Background(), &SendEmailRequest{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestEmail_SendEmail_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/emails" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(SendEmailResponse{ID: "123", Status: "queued"})
	})
	c := newTestClient(t, handler)

	resp, err := c.Emails.SendEmail(context.Background(), &SendEmailRequest{
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Subject: "hi",
		Text:    "hello",
	})
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if resp == nil || resp.ID == "" {
		t.Fatalf("expected response with ID, got %+v", resp)
	}
}

func TestEmail_ListEmails(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/emails" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]Email{{ID: "1"}, {ID: "2"}})
	})
	c := newTestClient(t, handler)

	out, err := c.Emails.ListEmails(context.Background(), &ListEmailsOptions{Page: 1, Limit: 50})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if out.TotalCount != 2 || out.TotalPages != 1 {
		t.Fatalf("unexpected pagination: %+v", out)
	}
}

func TestEmail_Get_Delete(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/emails/abc", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(Email{ID: "abc"})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	c := newTestClient(t, mux)

	e, err := c.Emails.GetEmail(context.Background(), "abc")
	if err != nil || e == nil || e.ID != "abc" {
		t.Fatalf("get failed: %v, e=%+v", err, e)
	}
	if err := c.Emails.DeleteEmail(context.Background(), "abc"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}
