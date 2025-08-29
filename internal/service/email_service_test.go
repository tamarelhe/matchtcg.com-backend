package service

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// MockEmailProvider implements EmailProvider for testing
type MockEmailProvider struct {
	mu         sync.RWMutex
	SentEmails []MockEmail
}

// MockEmail represents a sent email for testing
type MockEmail struct {
	To       []string
	Subject  string
	HTMLBody string
	TextBody string
}

// NewMockEmailProvider creates a new mock email provider for testing
func NewMockEmailProvider() *MockEmailProvider {
	return &MockEmailProvider{
		SentEmails: make([]MockEmail, 0),
	}
}

// SendEmail sends a plain text email (mock implementation)
func (m *MockEmailProvider) SendEmail(ctx context.Context, to []string, subject, body string) error {
	return m.SendHTMLEmail(ctx, to, subject, "", body)
}

// SendHTMLEmail sends an HTML email with text fallback (mock implementation)
func (m *MockEmailProvider) SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SentEmails = append(m.SentEmails, MockEmail{
		To:       to,
		Subject:  subject,
		HTMLBody: htmlBody,
		TextBody: textBody,
	})
	return nil
}

// Reset clears all sent emails (for testing)
func (m *MockEmailProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.SentEmails = make([]MockEmail, 0)
}

// GetLastEmail returns the last sent email (for testing)
func (m *MockEmailProvider) GetLastEmail() *MockEmail {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.SentEmails) == 0 {
		return nil
	}
	return &m.SentEmails[len(m.SentEmails)-1]
}

// GetEmailCount returns the number of sent emails (for testing)
func (m *MockEmailProvider) GetEmailCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.SentEmails)
}

func TestMockEmailProvider(t *testing.T) {
	provider := NewMockEmailProvider()
	service := NewEmailService(provider, "test@matchtcg.com", "MatchTCG")

	ctx := context.Background()

	t.Run("SendEmail", func(t *testing.T) {
		to := []string{"user@example.com"}
		subject := "Test Subject"
		body := "Test body content"

		err := service.SendEmail(ctx, to, subject, body)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if provider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email sent, got %d", provider.GetEmailCount())
		}

		lastEmail := provider.GetLastEmail()
		if lastEmail == nil {
			t.Fatal("Expected last email to exist")
		}

		if len(lastEmail.To) != 1 || lastEmail.To[0] != "user@example.com" {
			t.Errorf("Expected to: [user@example.com], got %v", lastEmail.To)
		}

		if lastEmail.Subject != subject {
			t.Errorf("Expected subject: %s, got %s", subject, lastEmail.Subject)
		}

		if lastEmail.TextBody != body {
			t.Errorf("Expected text body: %s, got %s", body, lastEmail.TextBody)
		}
	})

	t.Run("SendHTMLEmail", func(t *testing.T) {
		provider.Reset()

		to := []string{"user1@example.com", "user2@example.com"}
		subject := "HTML Test Subject"
		htmlBody := "<h1>HTML Content</h1>"
		textBody := "Text Content"

		err := service.SendHTMLEmail(ctx, to, subject, htmlBody, textBody)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if provider.GetEmailCount() != 1 {
			t.Fatalf("Expected 1 email sent, got %d", provider.GetEmailCount())
		}

		lastEmail := provider.GetLastEmail()
		if len(lastEmail.To) != 2 {
			t.Errorf("Expected 2 recipients, got %d", len(lastEmail.To))
		}

		if lastEmail.HTMLBody != htmlBody {
			t.Errorf("Expected HTML body: %s, got %s", htmlBody, lastEmail.HTMLBody)
		}

		if lastEmail.TextBody != textBody {
			t.Errorf("Expected text body: %s, got %s", textBody, lastEmail.TextBody)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		provider.Reset()
		if provider.GetEmailCount() != 0 {
			t.Errorf("Expected 0 emails after reset, got %d", provider.GetEmailCount())
		}

		if provider.GetLastEmail() != nil {
			t.Error("Expected no last email after reset")
		}
	})
}

func TestSMTPProvider_buildMessage(t *testing.T) {
	provider := NewSMTPProvider("smtp.example.com", "587", "user", "pass", "from@example.com", "Test Sender")

	t.Run("PlainTextMessage", func(t *testing.T) {
		to := []string{"user@example.com"}
		subject := "Test Subject"
		textBody := "Plain text content"

		msg := provider.buildMessage(to, subject, "", textBody)

		expectedParts := []string{
			"From: Test Sender <from@example.com>",
			"To: user@example.com",
			"Subject: Test Subject",
			"MIME-Version: 1.0",
			"Content-Type: text/plain; charset=UTF-8",
			"Plain text content",
		}

		for _, part := range expectedParts {
			if !strings.Contains(msg, part) {
				t.Errorf("Expected message to contain '%s', but it didn't. Message: %s", part, msg)
			}
		}
	})

	t.Run("HTMLMessage", func(t *testing.T) {
		to := []string{"user@example.com"}
		subject := "HTML Test"
		htmlBody := "<h1>HTML Content</h1>"
		textBody := "Text fallback"

		msg := provider.buildMessage(to, subject, htmlBody, textBody)

		expectedParts := []string{
			"From: Test Sender <from@example.com>",
			"To: user@example.com",
			"Subject: HTML Test",
			"Content-Type: multipart/alternative",
			"Content-Type: text/plain; charset=UTF-8",
			"Content-Type: text/html; charset=UTF-8",
			"<h1>HTML Content</h1>",
			"Text fallback",
		}

		for _, part := range expectedParts {
			if !strings.Contains(msg, part) {
				t.Errorf("Expected message to contain '%s', but it didn't. Message: %s", part, msg)
			}
		}
	})

	t.Run("MultipleRecipients", func(t *testing.T) {
		to := []string{"user1@example.com", "user2@example.com", "user3@example.com"}
		subject := "Multiple Recipients"
		textBody := "Content for all"

		msg := provider.buildMessage(to, subject, "", textBody)

		expectedTo := "To: user1@example.com, user2@example.com, user3@example.com"
		if !strings.Contains(msg, expectedTo) {
			t.Errorf("Expected message to contain '%s', but it didn't. Message: %s", expectedTo, msg)
		}
	})
}
