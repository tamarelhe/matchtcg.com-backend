package service

import (
	"context"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// MessageKey represents a message key for localization
type MessageKey string

// Message keys for various parts of the application
const (
	// Authentication messages
	MsgWelcome             MessageKey = "welcome"
	MsgLoginSuccess        MessageKey = "login_success"
	MsgLogoutSuccess       MessageKey = "logout_success"
	MsgRegistrationSuccess MessageKey = "registration_success"
	MsgPasswordResetSent   MessageKey = "password_reset_sent"

	// Event messages
	MsgEventCreated              MessageKey = "event_created"
	MsgEventUpdated              MessageKey = "event_updated"
	MsgEventDeleted              MessageKey = "event_deleted"
	MsgEventRSVPConfirmed        MessageKey = "event_rsvp_confirmed"
	MsgEventRSVPCancelled        MessageKey = "event_rsvp_cancelled"
	MsgEventCapacityReached      MessageKey = "event_capacity_reached"
	MsgEventWaitlisted           MessageKey = "event_waitlisted"
	MsgEventPromotedFromWaitlist MessageKey = "event_promoted_from_waitlist"

	// Group messages
	MsgGroupCreated        MessageKey = "group_created"
	MsgGroupUpdated        MessageKey = "group_updated"
	MsgGroupDeleted        MessageKey = "group_deleted"
	MsgGroupMemberAdded    MessageKey = "group_member_added"
	MsgGroupMemberRemoved  MessageKey = "group_member_removed"
	MsgGroupInvitationSent MessageKey = "group_invitation_sent"

	// Notification messages
	MsgEventReminder MessageKey = "event_reminder"
	MsgEventUpdate   MessageKey = "event_update"
	MsgNewGroupEvent MessageKey = "new_group_event"

	// Error messages
	MsgErrorGeneric          MessageKey = "error_generic"
	MsgErrorNotFound         MessageKey = "error_not_found"
	MsgErrorUnauthorized     MessageKey = "error_unauthorized"
	MsgErrorValidation       MessageKey = "error_validation"
	MsgErrorCapacityExceeded MessageKey = "error_capacity_exceeded"

	// Time-related messages
	MsgTimeAgo     MessageKey = "time_ago"
	MsgTimeFromNow MessageKey = "time_from_now"
	MsgDaysAgo     MessageKey = "days_ago"
	MsgHoursAgo    MessageKey = "hours_ago"
	MsgMinutesAgo  MessageKey = "minutes_ago"
)

// initializeMessages sets up the message catalog for all supported languages
func init() {
	// Portuguese messages
	message.SetString(language.Portuguese, string(MsgWelcome), "Bem-vindo ao MatchTCG!")
	message.SetString(language.Portuguese, string(MsgLoginSuccess), "Login realizado com sucesso")
	message.SetString(language.Portuguese, string(MsgLogoutSuccess), "Logout realizado com sucesso")
	message.SetString(language.Portuguese, string(MsgRegistrationSuccess), "Conta criada com sucesso")
	message.SetString(language.Portuguese, string(MsgPasswordResetSent), "Email de recuperação de senha enviado")

	message.SetString(language.Portuguese, string(MsgEventCreated), "Evento criado com sucesso")
	message.SetString(language.Portuguese, string(MsgEventUpdated), "Evento atualizado com sucesso")
	message.SetString(language.Portuguese, string(MsgEventDeleted), "Evento removido com sucesso")
	message.SetString(language.Portuguese, string(MsgEventRSVPConfirmed), "Confirmação de presença registrada")
	message.SetString(language.Portuguese, string(MsgEventRSVPCancelled), "Presença cancelada")
	message.SetString(language.Portuguese, string(MsgEventCapacityReached), "Evento lotado")
	message.SetString(language.Portuguese, string(MsgEventWaitlisted), "Adicionado à lista de espera")
	message.SetString(language.Portuguese, string(MsgEventPromotedFromWaitlist), "Promovido da lista de espera")

	message.SetString(language.Portuguese, string(MsgGroupCreated), "Grupo criado com sucesso")
	message.SetString(language.Portuguese, string(MsgGroupUpdated), "Grupo atualizado com sucesso")
	message.SetString(language.Portuguese, string(MsgGroupDeleted), "Grupo removido com sucesso")
	message.SetString(language.Portuguese, string(MsgGroupMemberAdded), "Membro adicionado ao grupo")
	message.SetString(language.Portuguese, string(MsgGroupMemberRemoved), "Membro removido do grupo")
	message.SetString(language.Portuguese, string(MsgGroupInvitationSent), "Convite enviado")

	message.SetString(language.Portuguese, string(MsgEventReminder), "Lembrete: %s começa em %s")
	message.SetString(language.Portuguese, string(MsgEventUpdate), "O evento %s foi atualizado")
	message.SetString(language.Portuguese, string(MsgNewGroupEvent), "Novo evento no grupo %s: %s")

	message.SetString(language.Portuguese, string(MsgErrorGeneric), "Ocorreu um erro inesperado")
	message.SetString(language.Portuguese, string(MsgErrorNotFound), "Recurso não encontrado")
	message.SetString(language.Portuguese, string(MsgErrorUnauthorized), "Acesso não autorizado")
	message.SetString(language.Portuguese, string(MsgErrorValidation), "Dados inválidos")
	message.SetString(language.Portuguese, string(MsgErrorCapacityExceeded), "Capacidade do evento excedida")

	message.SetString(language.Portuguese, string(MsgTimeAgo), "há %s")
	message.SetString(language.Portuguese, string(MsgTimeFromNow), "em %s")
	message.SetString(language.Portuguese, string(MsgDaysAgo), "%d dias")
	message.SetString(language.Portuguese, string(MsgHoursAgo), "%d horas")
	message.SetString(language.Portuguese, string(MsgMinutesAgo), "%d minutos")

	// English messages
	message.SetString(language.English, string(MsgWelcome), "Welcome to MatchTCG!")
	message.SetString(language.English, string(MsgLoginSuccess), "Login successful")
	message.SetString(language.English, string(MsgLogoutSuccess), "Logout successful")
	message.SetString(language.English, string(MsgRegistrationSuccess), "Account created successfully")
	message.SetString(language.English, string(MsgPasswordResetSent), "Password reset email sent")

	message.SetString(language.English, string(MsgEventCreated), "Event created successfully")
	message.SetString(language.English, string(MsgEventUpdated), "Event updated successfully")
	message.SetString(language.English, string(MsgEventDeleted), "Event deleted successfully")
	message.SetString(language.English, string(MsgEventRSVPConfirmed), "RSVP confirmed")
	message.SetString(language.English, string(MsgEventRSVPCancelled), "RSVP cancelled")
	message.SetString(language.English, string(MsgEventCapacityReached), "Event is at capacity")
	message.SetString(language.English, string(MsgEventWaitlisted), "Added to waitlist")
	message.SetString(language.English, string(MsgEventPromotedFromWaitlist), "Promoted from waitlist")

	message.SetString(language.English, string(MsgGroupCreated), "Group created successfully")
	message.SetString(language.English, string(MsgGroupUpdated), "Group updated successfully")
	message.SetString(language.English, string(MsgGroupDeleted), "Group deleted successfully")
	message.SetString(language.English, string(MsgGroupMemberAdded), "Member added to group")
	message.SetString(language.English, string(MsgGroupMemberRemoved), "Member removed from group")
	message.SetString(language.English, string(MsgGroupInvitationSent), "Invitation sent")

	message.SetString(language.English, string(MsgEventReminder), "Reminder: %s starts in %s")
	message.SetString(language.English, string(MsgEventUpdate), "Event %s has been updated")
	message.SetString(language.English, string(MsgNewGroupEvent), "New event in group %s: %s")

	message.SetString(language.English, string(MsgErrorGeneric), "An unexpected error occurred")
	message.SetString(language.English, string(MsgErrorNotFound), "Resource not found")
	message.SetString(language.English, string(MsgErrorUnauthorized), "Unauthorized access")
	message.SetString(language.English, string(MsgErrorValidation), "Invalid data")
	message.SetString(language.English, string(MsgErrorCapacityExceeded), "Event capacity exceeded")

	message.SetString(language.English, string(MsgTimeAgo), "%s ago")
	message.SetString(language.English, string(MsgTimeFromNow), "in %s")
	message.SetString(language.English, string(MsgDaysAgo), "%d days")
	message.SetString(language.English, string(MsgHoursAgo), "%d hours")
	message.SetString(language.English, string(MsgMinutesAgo), "%d minutes")
}

// GetMessage returns a localized message for the given key and locale
func (s *I18nService) GetMessage(ctx context.Context, locale SupportedLocale, key MessageKey, args ...interface{}) string {
	return s.FormatMessage(ctx, locale, string(key), args...)
}
