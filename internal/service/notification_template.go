package service

import (
	"bytes"
	"fmt"
	"html/template"
	text_template "text/template"

	"github.com/matchtcg/backend/internal/domain"
)

// NotificationTemplate represents an email template
type NotificationTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
	htmlTmpl *template.Template
	textTmpl *text_template.Template
}

// NotificationTemplateManager manages email templates for different notification types
type NotificationTemplateManager struct {
	templates map[domain.NotificationType]*NotificationTemplate
	baseURL   string
}

// NewNotificationTemplateManager creates a new template manager
func NewNotificationTemplateManager(baseURL string) *NotificationTemplateManager {
	manager := &NotificationTemplateManager{
		templates: make(map[domain.NotificationType]*NotificationTemplate),
		baseURL:   baseURL,
	}

	// Initialize default templates
	manager.initializeTemplates()
	return manager
}

// GetTemplate returns the template for a specific notification type
func (m *NotificationTemplateManager) GetTemplate(notificationType domain.NotificationType) (*NotificationTemplate, error) {
	template, exists := m.templates[notificationType]
	if !exists {
		return nil, fmt.Errorf("template not found for notification type: %s", notificationType)
	}
	return template, nil
}

// RenderTemplate renders a template with the provided data
func (m *NotificationTemplateManager) RenderTemplate(notificationType domain.NotificationType, data map[string]interface{}) (subject, htmlBody, textBody string, err error) {
	tmpl, err := m.GetTemplate(notificationType)
	if err != nil {
		return "", "", "", err
	}

	// Add base URL to template data
	data["BaseURL"] = m.baseURL

	// Render subject
	subject, err = m.renderString(tmpl.Subject, data)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to render subject: %w", err)
	}

	// Render HTML body
	if tmpl.htmlTmpl != nil {
		var htmlBuf bytes.Buffer
		if err := tmpl.htmlTmpl.Execute(&htmlBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render HTML body: %w", err)
		}
		htmlBody = htmlBuf.String()
	}

	// Render text body
	if tmpl.textTmpl != nil {
		var textBuf bytes.Buffer
		if err := tmpl.textTmpl.Execute(&textBuf, data); err != nil {
			return "", "", "", fmt.Errorf("failed to render text body: %w", err)
		}
		textBody = textBuf.String()
	}

	return subject, htmlBody, textBody, nil
}

// renderString renders a simple string template
func (m *NotificationTemplateManager) renderString(tmplStr string, data map[string]interface{}) (string, error) {
	tmpl, err := text_template.New("subject").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// initializeTemplates sets up default templates for all notification types
func (m *NotificationTemplateManager) initializeTemplates() {
	// Event RSVP Confirmation Template
	m.templates[domain.NotificationTypeEventRSVP] = &NotificationTemplate{
		Subject:  "RSVP Confirmation: {{.EventTitle}}",
		HTMLBody: eventRSVPHTMLTemplate,
		TextBody: eventRSVPTextTemplate,
	}

	// Event Update Template
	m.templates[domain.NotificationTypeEventUpdate] = &NotificationTemplate{
		Subject:  "Event Updated: {{.EventTitle}}",
		HTMLBody: eventUpdateHTMLTemplate,
		TextBody: eventUpdateTextTemplate,
	}

	// Event Reminder Template
	m.templates[domain.NotificationTypeEventReminder] = &NotificationTemplate{
		Subject:  "Reminder: {{.EventTitle}} is coming up!",
		HTMLBody: eventReminderHTMLTemplate,
		TextBody: eventReminderTextTemplate,
	}

	// Group Invite Template
	m.templates[domain.NotificationTypeGroupInvite] = &NotificationTemplate{
		Subject:  "You've been invited to join {{.GroupName}}",
		HTMLBody: groupInviteHTMLTemplate,
		TextBody: groupInviteTextTemplate,
	}

	// Group Event Template
	m.templates[domain.NotificationTypeGroupEvent] = &NotificationTemplate{
		Subject:  "New Event in {{.GroupName}}: {{.EventTitle}}",
		HTMLBody: groupEventHTMLTemplate,
		TextBody: groupEventTextTemplate,
	}

	// Compile templates
	for _, tmpl := range m.templates {
		if tmpl.HTMLBody != "" {
			htmlTmpl, err := template.New("html").Parse(tmpl.HTMLBody)
			if err == nil {
				tmpl.htmlTmpl = htmlTmpl
			}
		}
		if tmpl.TextBody != "" {
			textTmpl, err := text_template.New("text").Parse(tmpl.TextBody)
			if err == nil {
				tmpl.textTmpl = textTmpl
			}
		}
	}
}

// Template constants
const eventRSVPHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>RSVP Confirmation</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c3e50;">RSVP Confirmation</h1>
        
        <p>Hi {{.UserName}},</p>
        
        <p>Your RSVP for <strong>{{.EventTitle}}</strong> has been confirmed!</p>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0;">Event Details</h3>
            <p><strong>Event:</strong> {{.EventTitle}}</p>
            <p><strong>Date:</strong> {{.EventDate}}</p>
            <p><strong>Time:</strong> {{.EventTime}}</p>
            <p><strong>Location:</strong> {{.VenueName}}<br>{{.VenueAddress}}</p>
            {{if .EventDescription}}<p><strong>Description:</strong> {{.EventDescription}}</p>{{end}}
            <p><strong>Your Status:</strong> {{.RSVPStatus}}</p>
        </div>
        
        <p><a href="{{.BaseURL}}/events/{{.EventID}}" style="background-color: #3498db; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Event</a></p>
        
        <p>See you there!</p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
            you can update your preferences in your account settings.
        </p>
    </div>
</body>
</html>
`

const eventRSVPTextTemplate = `
RSVP Confirmation

Hi {{.UserName}},

Your RSVP for {{.EventTitle}} has been confirmed!

Event Details:
- Event: {{.EventTitle}}
- Date: {{.EventDate}}
- Time: {{.EventTime}}
- Location: {{.VenueName}}, {{.VenueAddress}}
{{if .EventDescription}}- Description: {{.EventDescription}}{{end}}
- Your Status: {{.RSVPStatus}}

View Event: {{.BaseURL}}/events/{{.EventID}}

See you there!

---
This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
you can update your preferences in your account settings.
`

const eventUpdateHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Event Updated</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #e74c3c;">Event Updated</h1>
        
        <p>Hi {{.UserName}},</p>
        
        <p>The event <strong>{{.EventTitle}}</strong> that you're attending has been updated.</p>
        
        <div style="background-color: #fff3cd; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #ffc107;">
            <h3 style="margin-top: 0;">What Changed</h3>
            <p>{{.UpdateMessage}}</p>
        </div>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0;">Current Event Details</h3>
            <p><strong>Event:</strong> {{.EventTitle}}</p>
            <p><strong>Date:</strong> {{.EventDate}}</p>
            <p><strong>Time:</strong> {{.EventTime}}</p>
            <p><strong>Location:</strong> {{.VenueName}}<br>{{.VenueAddress}}</p>
            {{if .EventDescription}}<p><strong>Description:</strong> {{.EventDescription}}</p>{{end}}
        </div>
        
        <p><a href="{{.BaseURL}}/events/{{.EventID}}" style="background-color: #3498db; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Updated Event</a></p>
        
        <p>Thanks for staying updated!</p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
            you can update your preferences in your account settings.
        </p>
    </div>
</body>
</html>
`

const eventUpdateTextTemplate = `
Event Updated

Hi {{.UserName}},

The event {{.EventTitle}} that you're attending has been updated.

What Changed:
{{.UpdateMessage}}

Current Event Details:
- Event: {{.EventTitle}}
- Date: {{.EventDate}}
- Time: {{.EventTime}}
- Location: {{.VenueName}}, {{.VenueAddress}}
{{if .EventDescription}}- Description: {{.EventDescription}}{{end}}

View Updated Event: {{.BaseURL}}/events/{{.EventID}}

Thanks for staying updated!

---
This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
you can update your preferences in your account settings.
`

const eventReminderHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Event Reminder</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #27ae60;">Event Reminder</h1>
        
        <p>Hi {{.UserName}},</p>
        
        <p>Don't forget! <strong>{{.EventTitle}}</strong> is coming up {{.TimeUntilEvent}}.</p>
        
        <div style="background-color: #d4edda; padding: 20px; border-radius: 5px; margin: 20px 0; border-left: 4px solid #28a745;">
            <h3 style="margin-top: 0;">Event Details</h3>
            <p><strong>Event:</strong> {{.EventTitle}}</p>
            <p><strong>Date:</strong> {{.EventDate}}</p>
            <p><strong>Time:</strong> {{.EventTime}}</p>
            <p><strong>Location:</strong> {{.VenueName}}<br>{{.VenueAddress}}</p>
            {{if .EventDescription}}<p><strong>Description:</strong> {{.EventDescription}}</p>{{end}}
        </div>
        
        <p><a href="{{.BaseURL}}/events/{{.EventID}}" style="background-color: #28a745; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Event</a></p>
        
        <p>We look forward to seeing you there!</p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
            you can update your preferences in your account settings.
        </p>
    </div>
</body>
</html>
`

const eventReminderTextTemplate = `
Event Reminder

Hi {{.UserName}},

Don't forget! {{.EventTitle}} is coming up {{.TimeUntilEvent}}.

Event Details:
- Event: {{.EventTitle}}
- Date: {{.EventDate}}
- Time: {{.EventTime}}
- Location: {{.VenueName}}, {{.VenueAddress}}
{{if .EventDescription}}- Description: {{.EventDescription}}{{end}}

View Event: {{.BaseURL}}/events/{{.EventID}}

We look forward to seeing you there!

---
This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
you can update your preferences in your account settings.
`

const groupInviteHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Group Invitation</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #8e44ad;">Group Invitation</h1>
        
        <p>Hi {{.UserName}},</p>
        
        <p>You've been invited to join the group <strong>{{.GroupName}}</strong>!</p>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0;">Group Details</h3>
            <p><strong>Group:</strong> {{.GroupName}}</p>
            {{if .GroupDescription}}<p><strong>Description:</strong> {{.GroupDescription}}</p>{{end}}
            <p><strong>Invited by:</strong> {{.InviterName}}</p>
            <p><strong>Role:</strong> {{.Role}}</p>
        </div>
        
        <p><a href="{{.BaseURL}}/groups/{{.GroupID}}/accept-invite?token={{.InviteToken}}" style="background-color: #8e44ad; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">Accept Invitation</a></p>
        
        <p>Join the group to participate in private events and connect with other members!</p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
            you can update your preferences in your account settings.
        </p>
    </div>
</body>
</html>
`

const groupInviteTextTemplate = `
Group Invitation

Hi {{.UserName}},

You've been invited to join the group {{.GroupName}}!

Group Details:
- Group: {{.GroupName}}
{{if .GroupDescription}}- Description: {{.GroupDescription}}{{end}}
- Invited by: {{.InviterName}}
- Role: {{.Role}}

Accept Invitation: {{.BaseURL}}/groups/{{.GroupID}}/accept-invite?token={{.InviteToken}}

Join the group to participate in private events and connect with other members!

---
This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
you can update your preferences in your account settings.
`

const groupEventHTMLTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>New Group Event</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #16a085;">New Group Event</h1>
        
        <p>Hi {{.UserName}},</p>
        
        <p>A new event has been created in your group <strong>{{.GroupName}}</strong>!</p>
        
        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0;">
            <h3 style="margin-top: 0;">Event Details</h3>
            <p><strong>Event:</strong> {{.EventTitle}}</p>
            <p><strong>Date:</strong> {{.EventDate}}</p>
            <p><strong>Time:</strong> {{.EventTime}}</p>
            <p><strong>Location:</strong> {{.VenueName}}<br>{{.VenueAddress}}</p>
            {{if .EventDescription}}<p><strong>Description:</strong> {{.EventDescription}}</p>{{end}}
            <p><strong>Hosted by:</strong> {{.HostName}}</p>
        </div>
        
        <p><a href="{{.BaseURL}}/events/{{.EventID}}" style="background-color: #16a085; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Event & RSVP</a></p>
        
        <p>Don't miss out on this group event!</p>
        
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #eee;">
        <p style="font-size: 12px; color: #666;">
            This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
            you can update your preferences in your account settings.
        </p>
    </div>
</body>
</html>
`

const groupEventTextTemplate = `
New Group Event

Hi {{.UserName}},

A new event has been created in your group {{.GroupName}}!

Event Details:
- Event: {{.EventTitle}}
- Date: {{.EventDate}}
- Time: {{.EventTime}}
- Location: {{.VenueName}}, {{.VenueAddress}}
{{if .EventDescription}}- Description: {{.EventDescription}}{{end}}
- Hosted by: {{.HostName}}

View Event & RSVP: {{.BaseURL}}/events/{{.EventID}}

Don't miss out on this group event!

---
This email was sent by MatchTCG. If you no longer wish to receive these notifications, 
you can update your preferences in your account settings.
`
