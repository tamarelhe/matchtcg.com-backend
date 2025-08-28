package service

import (
	"strings"
	"testing"

	"github.com/matchtcg/backend/internal/domain"
)

func TestNotificationTemplateManager(t *testing.T) {
	baseURL := "https://matchtcg.com"
	manager := NewNotificationTemplateManager(baseURL)

	t.Run("GetTemplate", func(t *testing.T) {
		// Test getting existing template
		template, err := manager.GetTemplate(domain.NotificationTypeEventRSVP)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if template == nil {
			t.Fatal("Expected template to exist")
		}

		if template.Subject == "" {
			t.Error("Expected template to have subject")
		}

		if template.HTMLBody == "" {
			t.Error("Expected template to have HTML body")
		}

		if template.TextBody == "" {
			t.Error("Expected template to have text body")
		}

		// Test getting non-existent template
		_, err = manager.GetTemplate("invalid_type")
		if err == nil {
			t.Error("Expected error for invalid template type")
		}
	})

	t.Run("RenderEventRSVPTemplate", func(t *testing.T) {
		data := map[string]interface{}{
			"UserName":         "John Doe",
			"EventTitle":       "Friday Night Magic",
			"EventDate":        "2024-01-05",
			"EventTime":        "19:00",
			"VenueName":        "Local Game Store",
			"VenueAddress":     "123 Main St, Lisbon",
			"EventDescription": "Weekly Standard tournament",
			"RSVPStatus":       "Going",
			"EventID":          "123e4567-e89b-12d3-a456-426614174000",
		}

		subject, htmlBody, textBody, err := manager.RenderTemplate(domain.NotificationTypeEventRSVP, data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check subject
		expectedSubject := "RSVP Confirmation: Friday Night Magic"
		if subject != expectedSubject {
			t.Errorf("Expected subject: %s, got %s", expectedSubject, subject)
		}

		// Check HTML body contains expected content
		expectedHTMLParts := []string{
			"John Doe",
			"Friday Night Magic",
			"2024-01-05",
			"19:00",
			"Local Game Store",
			"123 Main St, Lisbon",
			"Weekly Standard tournament",
			"Going",
			baseURL + "/events/123e4567-e89b-12d3-a456-426614174000",
		}

		for _, part := range expectedHTMLParts {
			if !strings.Contains(htmlBody, part) {
				t.Errorf("Expected HTML body to contain '%s', but it didn't", part)
			}
		}

		// Check text body contains expected content
		for _, part := range expectedHTMLParts {
			if !strings.Contains(textBody, part) {
				t.Errorf("Expected text body to contain '%s', but it didn't", part)
			}
		}

		// Check base URL is included
		if !strings.Contains(htmlBody, baseURL) {
			t.Error("Expected HTML body to contain base URL")
		}

		if !strings.Contains(textBody, baseURL) {
			t.Error("Expected text body to contain base URL")
		}
	})

	t.Run("RenderEventUpdateTemplate", func(t *testing.T) {
		data := map[string]interface{}{
			"UserName":      "Jane Smith",
			"EventTitle":    "Commander Tournament",
			"UpdateMessage": "The start time has been changed from 18:00 to 19:00",
			"EventDate":     "2024-01-10",
			"EventTime":     "19:00",
			"VenueName":     "Game Hub",
			"VenueAddress":  "456 Oak Ave, Porto",
			"EventID":       "456e7890-e89b-12d3-a456-426614174001",
		}

		subject, htmlBody, textBody, err := manager.RenderTemplate(domain.NotificationTypeEventUpdate, data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedSubject := "Event Updated: Commander Tournament"
		if subject != expectedSubject {
			t.Errorf("Expected subject: %s, got %s", expectedSubject, subject)
		}

		expectedParts := []string{
			"Jane Smith",
			"Commander Tournament",
			"The start time has been changed from 18:00 to 19:00",
			"2024-01-10",
			"19:00",
			"Game Hub",
		}

		for _, part := range expectedParts {
			if !strings.Contains(htmlBody, part) {
				t.Errorf("Expected HTML body to contain '%s', but it didn't", part)
			}
			if !strings.Contains(textBody, part) {
				t.Errorf("Expected text body to contain '%s', but it didn't", part)
			}
		}
	})

	t.Run("RenderEventReminderTemplate", func(t *testing.T) {
		data := map[string]interface{}{
			"UserName":       "Bob Wilson",
			"EventTitle":     "Draft Night",
			"TimeUntilEvent": "tomorrow",
			"EventDate":      "2024-01-15",
			"EventTime":      "20:00",
			"VenueName":      "Card Castle",
			"VenueAddress":   "789 Pine St, Braga",
			"EventID":        "789e0123-e89b-12d3-a456-426614174002",
		}

		subject, htmlBody, textBody, err := manager.RenderTemplate(domain.NotificationTypeEventReminder, data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedSubject := "Reminder: Draft Night is coming up!"
		if subject != expectedSubject {
			t.Errorf("Expected subject: %s, got %s", expectedSubject, subject)
		}

		expectedParts := []string{
			"Bob Wilson",
			"Draft Night",
			"tomorrow",
			"2024-01-15",
			"20:00",
			"Card Castle",
		}

		for _, part := range expectedParts {
			if !strings.Contains(htmlBody, part) {
				t.Errorf("Expected HTML body to contain '%s', but it didn't", part)
			}
			if !strings.Contains(textBody, part) {
				t.Errorf("Expected text body to contain '%s', but it didn't", part)
			}
		}
	})

	t.Run("RenderGroupInviteTemplate", func(t *testing.T) {
		data := map[string]interface{}{
			"UserName":         "Alice Cooper",
			"GroupName":        "Lisbon MTG Players",
			"GroupDescription": "A group for Magic players in Lisbon",
			"InviterName":      "Charlie Brown",
			"Role":             "Member",
			"GroupID":          "abc12345-e89b-12d3-a456-426614174003",
			"InviteToken":      "invite-token-123",
		}

		subject, htmlBody, textBody, err := manager.RenderTemplate(domain.NotificationTypeGroupInvite, data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedSubject := "You've been invited to join Lisbon MTG Players"
		if subject != expectedSubject {
			t.Errorf("Expected subject: %s, got %s", expectedSubject, subject)
		}

		expectedParts := []string{
			"Alice Cooper",
			"Lisbon MTG Players",
			"A group for Magic players in Lisbon",
			"Charlie Brown",
			"Member",
			baseURL + "/groups/abc12345-e89b-12d3-a456-426614174003/accept-invite?token=invite-token-123",
		}

		for _, part := range expectedParts {
			if !strings.Contains(htmlBody, part) {
				t.Errorf("Expected HTML body to contain '%s', but it didn't", part)
			}
			if !strings.Contains(textBody, part) {
				t.Errorf("Expected text body to contain '%s', but it didn't", part)
			}
		}
	})

	t.Run("RenderGroupEventTemplate", func(t *testing.T) {
		data := map[string]interface{}{
			"UserName":     "David Lee",
			"GroupName":    "Porto TCG Community",
			"EventTitle":   "Sealed Deck Tournament",
			"EventDate":    "2024-01-20",
			"EventTime":    "14:00",
			"VenueName":    "Gaming Center",
			"VenueAddress": "321 Cedar Rd, Porto",
			"HostName":     "Event Host",
			"EventID":      "def45678-e89b-12d3-a456-426614174004",
		}

		subject, htmlBody, textBody, err := manager.RenderTemplate(domain.NotificationTypeGroupEvent, data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expectedSubject := "New Event in Porto TCG Community: Sealed Deck Tournament"
		if subject != expectedSubject {
			t.Errorf("Expected subject: %s, got %s", expectedSubject, subject)
		}

		expectedParts := []string{
			"David Lee",
			"Porto TCG Community",
			"Sealed Deck Tournament",
			"2024-01-20",
			"14:00",
			"Gaming Center",
			"Event Host",
		}

		for _, part := range expectedParts {
			if !strings.Contains(htmlBody, part) {
				t.Errorf("Expected HTML body to contain '%s', but it didn't", part)
			}
			if !strings.Contains(textBody, part) {
				t.Errorf("Expected text body to contain '%s', but it didn't", part)
			}
		}
	})

	t.Run("RenderWithMissingData", func(t *testing.T) {
		// Test with minimal data to ensure templates handle missing fields gracefully
		data := map[string]interface{}{
			"UserName":   "Test User",
			"EventTitle": "Test Event",
			"EventID":    "test-id",
		}

		subject, htmlBody, textBody, err := manager.RenderTemplate(domain.NotificationTypeEventRSVP, data)
		if err != nil {
			t.Fatalf("Expected no error with minimal data, got %v", err)
		}

		// Should still contain the basic required fields
		if !strings.Contains(subject, "Test Event") {
			t.Error("Expected subject to contain event title")
		}

		if !strings.Contains(htmlBody, "Test User") {
			t.Error("Expected HTML body to contain user name")
		}

		if !strings.Contains(textBody, "Test User") {
			t.Error("Expected text body to contain user name")
		}
	})

	t.Run("AllNotificationTypesHaveTemplates", func(t *testing.T) {
		notificationTypes := []domain.NotificationType{
			domain.NotificationTypeEventRSVP,
			domain.NotificationTypeEventUpdate,
			domain.NotificationTypeEventReminder,
			domain.NotificationTypeGroupInvite,
			domain.NotificationTypeGroupEvent,
		}

		for _, notType := range notificationTypes {
			template, err := manager.GetTemplate(notType)
			if err != nil {
				t.Errorf("Expected template to exist for notification type %s, got error: %v", notType, err)
			}

			if template == nil {
				t.Errorf("Expected template to exist for notification type %s", notType)
			}

			if template.Subject == "" {
				t.Errorf("Expected template subject to be non-empty for notification type %s", notType)
			}

			if template.HTMLBody == "" {
				t.Errorf("Expected template HTML body to be non-empty for notification type %s", notType)
			}

			if template.TextBody == "" {
				t.Errorf("Expected template text body to be non-empty for notification type %s", notType)
			}
		}
	})
}
