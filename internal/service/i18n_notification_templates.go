package service

import (
	"context"
	"fmt"
	"time"
)

// NotificationTemplateType represents different types of notification templates
type NotificationTemplateType string

const (
	TemplateEventRSVPConfirmation NotificationTemplateType = "event_rsvp_confirmation"
	TemplateEventUpdate           NotificationTemplateType = "event_update"
	TemplateEventReminder         NotificationTemplateType = "event_reminder"
	TemplateEventCancellation     NotificationTemplateType = "event_cancellation"
	TemplateWaitlistPromotion     NotificationTemplateType = "waitlist_promotion"
	TemplateGroupInvitation       NotificationTemplateType = "group_invitation"
	TemplateNewGroupEvent         NotificationTemplateType = "new_group_event"
	TemplateWelcome               NotificationTemplateType = "welcome"
	TemplatePasswordReset         NotificationTemplateType = "password_reset"
)

// LocalizedNotificationTemplate represents a localized notification template
type LocalizedNotificationTemplate struct {
	Subject string
	Body    string
}

// NotificationTemplateData contains data for template rendering
type NotificationTemplateData struct {
	UserName   string
	EventTitle string
	EventDate  string
	EventTime  string
	EventVenue string
	GroupName  string
	HostName   string
	UpdatedBy  string
	Changes    string
	ActionURL  string
	AppName    string
}

// GetNotificationTemplate returns a localized notification template
func (s *I18nService) GetNotificationTemplate(ctx context.Context, locale SupportedLocale, templateType NotificationTemplateType, data NotificationTemplateData) LocalizedNotificationTemplate {
	switch locale {
	case LocalePortuguese:
		return s.getPortugueseTemplate(templateType, data)
	case LocaleEnglish:
		return s.getEnglishTemplate(templateType, data)
	default:
		return s.getPortugueseTemplate(templateType, data)
	}
}

// getPortugueseTemplate returns Portuguese notification templates
func (s *I18nService) getPortugueseTemplate(templateType NotificationTemplateType, data NotificationTemplateData) LocalizedNotificationTemplate {
	switch templateType {
	case TemplateEventRSVPConfirmation:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Confirmação de presença - %s", data.EventTitle),
			Body: fmt.Sprintf(`Olá %s,

Sua presença foi confirmada para o evento:

📅 Evento: %s
🗓️ Data: %s às %s
📍 Local: %s
👤 Organizador: %s

Você pode visualizar mais detalhes do evento em: %s

Nos vemos lá!

Equipe %s`,
				data.UserName, data.EventTitle, data.EventDate, data.EventTime,
				data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateEventUpdate:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Evento atualizado - %s", data.EventTitle),
			Body: fmt.Sprintf(`Olá %s,

O evento "%s" foi atualizado por %s.

Alterações realizadas:
%s

📅 Evento: %s
🗓️ Data: %s às %s
📍 Local: %s

Visualize as alterações completas em: %s

Equipe %s`,
				data.UserName, data.EventTitle, data.UpdatedBy, data.Changes,
				data.EventTitle, data.EventDate, data.EventTime, data.EventVenue,
				data.ActionURL, data.AppName),
		}

	case TemplateEventReminder:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Lembrete - %s começa em breve", data.EventTitle),
			Body: fmt.Sprintf(`Olá %s,

Este é um lembrete de que o evento "%s" começa em breve!

📅 Evento: %s
🗓️ Data: %s às %s
📍 Local: %s
👤 Organizador: %s

Não se esqueça de levar seus cards e chegar com antecedência.

Ver detalhes do evento: %s

Bom jogo!
Equipe %s`,
				data.UserName, data.EventTitle, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateEventCancellation:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Evento cancelado - %s", data.EventTitle),
			Body: fmt.Sprintf(`Olá %s,

Infelizmente, o evento "%s" foi cancelado.

📅 Evento: %s
🗓️ Data original: %s às %s
📍 Local: %s
👤 Organizador: %s

Pedimos desculpas pelo inconveniente. Fique atento a novos eventos na sua região!

Explorar outros eventos: %s

Equipe %s`,
				data.UserName, data.EventTitle, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateWaitlistPromotion:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Vaga disponível - %s", data.EventTitle),
			Body: fmt.Sprintf(`Olá %s,

Boa notícia! Uma vaga foi liberada no evento "%s" e você foi promovido da lista de espera.

📅 Evento: %s
🗓️ Data: %s às %s
📍 Local: %s
👤 Organizador: %s

Sua presença está confirmada. Nos vemos lá!

Ver detalhes do evento: %s

Equipe %s`,
				data.UserName, data.EventTitle, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateGroupInvitation:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Convite para o grupo %s", data.GroupName),
			Body: fmt.Sprintf(`Olá %s,

Você foi convidado para participar do grupo "%s"!

Os grupos permitem que você organize eventos privados e se conecte com outros jogadores da sua comunidade.

Aceitar convite: %s

Equipe %s`,
				data.UserName, data.GroupName, data.ActionURL, data.AppName),
		}

	case TemplateNewGroupEvent:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Novo evento no grupo %s", data.GroupName),
			Body: fmt.Sprintf(`Olá %s,

Um novo evento foi criado no grupo "%s":

📅 Evento: %s
🗓️ Data: %s às %s
📍 Local: %s
👤 Organizador: %s

Ver evento e confirmar presença: %s

Equipe %s`,
				data.UserName, data.GroupName, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateWelcome:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Bem-vindo ao %s!", data.AppName),
			Body: fmt.Sprintf(`Olá %s,

Bem-vindo ao %s! Estamos muito felizes em tê-lo conosco.

Com o %s você pode:
• Encontrar eventos de TCG na sua região
• Criar e organizar seus próprios eventos
• Formar grupos com outros jogadores
• Receber lembretes dos seus eventos

Comece explorando eventos na sua cidade: %s

Bom jogo!
Equipe %s`,
				data.UserName, data.AppName, data.AppName, data.ActionURL, data.AppName),
		}

	case TemplatePasswordReset:
		return LocalizedNotificationTemplate{
			Subject: "Recuperação de senha",
			Body: fmt.Sprintf(`Olá %s,

Você solicitou a recuperação da sua senha no %s.

Para criar uma nova senha, clique no link abaixo:
%s

Se você não solicitou esta recuperação, ignore este email.

Este link expira em 1 hora por motivos de segurança.

Equipe %s`,
				data.UserName, data.AppName, data.ActionURL, data.AppName),
		}

	default:
		return LocalizedNotificationTemplate{
			Subject: "Notificação",
			Body:    fmt.Sprintf("Olá %s,\n\nVocê tem uma nova notificação.\n\nEquipe %s", data.UserName, data.AppName),
		}
	}
}

// getEnglishTemplate returns English notification templates
func (s *I18nService) getEnglishTemplate(templateType NotificationTemplateType, data NotificationTemplateData) LocalizedNotificationTemplate {
	switch templateType {
	case TemplateEventRSVPConfirmation:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("RSVP Confirmation - %s", data.EventTitle),
			Body: fmt.Sprintf(`Hi %s,

Your attendance has been confirmed for the event:

📅 Event: %s
🗓️ Date: %s at %s
📍 Location: %s
👤 Host: %s

You can view more event details at: %s

See you there!

%s Team`,
				data.UserName, data.EventTitle, data.EventDate, data.EventTime,
				data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateEventUpdate:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Event Updated - %s", data.EventTitle),
			Body: fmt.Sprintf(`Hi %s,

The event "%s" has been updated by %s.

Changes made:
%s

📅 Event: %s
🗓️ Date: %s at %s
📍 Location: %s

View complete changes at: %s

%s Team`,
				data.UserName, data.EventTitle, data.UpdatedBy, data.Changes,
				data.EventTitle, data.EventDate, data.EventTime, data.EventVenue,
				data.ActionURL, data.AppName),
		}

	case TemplateEventReminder:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Reminder - %s starts soon", data.EventTitle),
			Body: fmt.Sprintf(`Hi %s,

This is a reminder that the event "%s" starts soon!

📅 Event: %s
🗓️ Date: %s at %s
📍 Location: %s
👤 Host: %s

Don't forget to bring your cards and arrive early.

View event details: %s

Good luck!
%s Team`,
				data.UserName, data.EventTitle, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateEventCancellation:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Event Cancelled - %s", data.EventTitle),
			Body: fmt.Sprintf(`Hi %s,

Unfortunately, the event "%s" has been cancelled.

📅 Event: %s
🗓️ Original date: %s at %s
📍 Location: %s
👤 Host: %s

We apologize for the inconvenience. Stay tuned for new events in your area!

Explore other events: %s

%s Team`,
				data.UserName, data.EventTitle, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateWaitlistPromotion:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Spot Available - %s", data.EventTitle),
			Body: fmt.Sprintf(`Hi %s,

Good news! A spot has opened up for the event "%s" and you've been promoted from the waitlist.

📅 Event: %s
🗓️ Date: %s at %s
📍 Location: %s
👤 Host: %s

Your attendance is now confirmed. See you there!

View event details: %s

%s Team`,
				data.UserName, data.EventTitle, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateGroupInvitation:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Invitation to group %s", data.GroupName),
			Body: fmt.Sprintf(`Hi %s,

You've been invited to join the group "%s"!

Groups allow you to organize private events and connect with other players in your community.

Accept invitation: %s

%s Team`,
				data.UserName, data.GroupName, data.ActionURL, data.AppName),
		}

	case TemplateNewGroupEvent:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("New event in group %s", data.GroupName),
			Body: fmt.Sprintf(`Hi %s,

A new event has been created in the group "%s":

📅 Event: %s
🗓️ Date: %s at %s
📍 Location: %s
👤 Host: %s

View event and RSVP: %s

%s Team`,
				data.UserName, data.GroupName, data.EventTitle, data.EventDate,
				data.EventTime, data.EventVenue, data.HostName, data.ActionURL, data.AppName),
		}

	case TemplateWelcome:
		return LocalizedNotificationTemplate{
			Subject: fmt.Sprintf("Welcome to %s!", data.AppName),
			Body: fmt.Sprintf(`Hi %s,

Welcome to %s! We're excited to have you with us.

With %s you can:
• Find TCG events in your area
• Create and organize your own events
• Form groups with other players
• Get reminders for your events

Start by exploring events in your city: %s

Good luck!
%s Team`,
				data.UserName, data.AppName, data.AppName, data.ActionURL, data.AppName),
		}

	case TemplatePasswordReset:
		return LocalizedNotificationTemplate{
			Subject: "Password Reset",
			Body: fmt.Sprintf(`Hi %s,

You requested a password reset for your %s account.

To create a new password, click the link below:
%s

If you didn't request this reset, please ignore this email.

This link expires in 1 hour for security reasons.

%s Team`,
				data.UserName, data.AppName, data.ActionURL, data.AppName),
		}

	default:
		return LocalizedNotificationTemplate{
			Subject: "Notification",
			Body:    fmt.Sprintf("Hi %s,\n\nYou have a new notification.\n\n%s Team", data.UserName, data.AppName),
		}
	}
}

// FormatRelativeTime formats a relative time string based on locale
func (s *I18nService) FormatRelativeTime(ctx context.Context, locale SupportedLocale, t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 0 {
		// Future time
		diff = -diff
		switch locale {
		case LocalePortuguese:
			if diff < time.Hour {
				minutes := int(diff.Minutes())
				return fmt.Sprintf("em %d minutos", minutes)
			} else if diff < 24*time.Hour {
				hours := int(diff.Hours())
				return fmt.Sprintf("em %d horas", hours)
			} else {
				days := int(diff.Hours() / 24)
				return fmt.Sprintf("em %d dias", days)
			}
		default:
			if diff < time.Hour {
				minutes := int(diff.Minutes())
				return fmt.Sprintf("in %d minutes", minutes)
			} else if diff < 24*time.Hour {
				hours := int(diff.Hours())
				return fmt.Sprintf("in %d hours", hours)
			} else {
				days := int(diff.Hours() / 24)
				return fmt.Sprintf("in %d days", days)
			}
		}
	} else {
		// Past time
		switch locale {
		case LocalePortuguese:
			if diff < time.Hour {
				minutes := int(diff.Minutes())
				return fmt.Sprintf("há %d minutos", minutes)
			} else if diff < 24*time.Hour {
				hours := int(diff.Hours())
				return fmt.Sprintf("há %d horas", hours)
			} else {
				days := int(diff.Hours() / 24)
				return fmt.Sprintf("há %d dias", days)
			}
		default:
			if diff < time.Hour {
				minutes := int(diff.Minutes())
				return fmt.Sprintf("%d minutes ago", minutes)
			} else if diff < 24*time.Hour {
				hours := int(diff.Hours())
				return fmt.Sprintf("%d hours ago", hours)
			} else {
				days := int(diff.Hours() / 24)
				return fmt.Sprintf("%d days ago", days)
			}
		}
	}
}
