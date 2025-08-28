package service

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailProvider defines the interface for email service providers
type EmailProvider interface {
	SendEmail(ctx context.Context, to []string, subject, body string) error
	SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error
}

// EmailService provides email sending capabilities with provider abstraction
type EmailService struct {
	provider EmailProvider
	fromAddr string
	fromName string
}

// NewEmailService creates a new email service with the specified provider
func NewEmailService(provider EmailProvider, fromAddr, fromName string) *EmailService {
	return &EmailService{
		provider: provider,
		fromAddr: fromAddr,
		fromName: fromName,
	}
}

// SendEmail sends a plain text email
func (s *EmailService) SendEmail(ctx context.Context, to []string, subject, body string) error {
	return s.provider.SendEmail(ctx, to, subject, body)
}

// SendHTMLEmail sends an HTML email with text fallback
func (s *EmailService) SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error {
	return s.provider.SendHTMLEmail(ctx, to, subject, htmlBody, textBody)
}

// SMTPProvider implements EmailProvider using SMTP
type SMTPProvider struct {
	host     string
	port     string
	username string
	password string
	fromAddr string
	fromName string
}

// NewSMTPProvider creates a new SMTP email provider
func NewSMTPProvider(host, port, username, password, fromAddr, fromName string) *SMTPProvider {
	return &SMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		fromAddr: fromAddr,
		fromName: fromName,
	}
}

// SendEmail sends a plain text email via SMTP
func (p *SMTPProvider) SendEmail(ctx context.Context, to []string, subject, body string) error {
	return p.SendHTMLEmail(ctx, to, subject, "", body)
}

// SendHTMLEmail sends an HTML email with text fallback via SMTP
func (p *SMTPProvider) SendHTMLEmail(ctx context.Context, to []string, subject, htmlBody, textBody string) error {
	// Create authentication
	auth := smtp.PlainAuth("", p.username, p.password, p.host)

	// Build message
	msg := p.buildMessage(to, subject, htmlBody, textBody)

	// Send email
	addr := fmt.Sprintf("%s:%s", p.host, p.port)
	err := smtp.SendMail(addr, auth, p.fromAddr, to, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildMessage constructs the email message with proper headers
func (p *SMTPProvider) buildMessage(to []string, subject, htmlBody, textBody string) string {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", p.fromName, p.fromAddr))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" {
		// Multipart message with HTML and text
		boundary := "boundary-matchtcg-email"
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
		msg.WriteString("\r\n")

		// Text part
		if textBody != "" {
			msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
			msg.WriteString("\r\n")
			msg.WriteString(textBody)
			msg.WriteString("\r\n")
		}

		// HTML part
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(htmlBody)
		msg.WriteString("\r\n")

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Plain text only
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		msg.WriteString("\r\n")
		msg.WriteString(textBody)
	}

	return msg.String()
}
