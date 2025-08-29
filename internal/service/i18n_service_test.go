package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewI18nService(t *testing.T) {
	service := NewI18nService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.printers)
	assert.Len(t, service.printers, 2)
	assert.Contains(t, service.printers, LocalePortuguese)
	assert.Contains(t, service.printers, LocaleEnglish)
}

func TestGetSupportedLocales(t *testing.T) {
	service := NewI18nService()
	locales := service.GetSupportedLocales()

	assert.Len(t, locales, 2)
	assert.Contains(t, locales, LocalePortuguese)
	assert.Contains(t, locales, LocaleEnglish)
}

func TestIsValidLocale(t *testing.T) {
	service := NewI18nService()

	tests := []struct {
		locale   string
		expected bool
	}{
		{"pt", true},
		{"en", true},
		{"fr", false},
		{"es", false},
		{"", false},
		{"PT", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			result := service.IsValidLocale(tt.locale)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDefaultLocale(t *testing.T) {
	service := NewI18nService()
	defaultLocale := service.GetDefaultLocale()

	assert.Equal(t, LocalePortuguese, defaultLocale)
}

func TestGetLocaleForCountry(t *testing.T) {
	service := NewI18nService()

	tests := []struct {
		country  string
		expected SupportedLocale
	}{
		{"PT", LocalePortuguese},
		{"BR", LocalePortuguese},
		{"US", LocaleEnglish},
		{"GB", LocaleEnglish},
		{"FR", LocaleEnglish}, // fallback to English
		{"", LocaleEnglish},   // fallback to English
	}

	for _, tt := range tests {
		t.Run(tt.country, func(t *testing.T) {
			result := service.GetLocaleForCountry(tt.country)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatMessage(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	// Test with valid locale
	result := service.FormatMessage(ctx, LocalePortuguese, "Hello %s", "World")
	assert.Contains(t, result, "World")

	// Test with invalid locale (should fallback to default)
	result = service.FormatMessage(ctx, SupportedLocale("invalid"), "Hello %s", "World")
	assert.Contains(t, result, "World")
}

func TestFormatDateTime(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	// Test time
	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		locale   SupportedLocale
		timezone string
		expected string
	}{
		{
			name:     "Portuguese format with UTC",
			locale:   LocalePortuguese,
			timezone: "UTC",
			expected: "15/01/2024 14:30",
		},
		{
			name:     "English format with UTC",
			locale:   LocaleEnglish,
			timezone: "UTC",
			expected: "01/15/2024 2:30 PM",
		},
		{
			name:     "Portuguese format with Lisbon timezone",
			locale:   LocalePortuguese,
			timezone: "Europe/Lisbon",
			expected: "15/01/2024 14:30", // Same as UTC in January
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FormatDateTime(ctx, tt.locale, testTime, tt.timezone)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test invalid timezone
	_, err := service.FormatDateTime(ctx, LocalePortuguese, testTime, "Invalid/Timezone")
	assert.Error(t, err)
}

func TestFormatDate(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		locale   SupportedLocale
		timezone string
		expected string
	}{
		{
			name:     "Portuguese date format",
			locale:   LocalePortuguese,
			timezone: "UTC",
			expected: "15/01/2024",
		},
		{
			name:     "English date format",
			locale:   LocaleEnglish,
			timezone: "UTC",
			expected: "01/15/2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FormatDate(ctx, tt.locale, testTime, tt.timezone)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTime(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		locale   SupportedLocale
		timezone string
		expected string
	}{
		{
			name:     "Portuguese time format",
			locale:   LocalePortuguese,
			timezone: "UTC",
			expected: "14:30",
		},
		{
			name:     "English time format",
			locale:   LocaleEnglish,
			timezone: "UTC",
			expected: "2:30 PM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FormatTime(ctx, tt.locale, testTime, tt.timezone)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToUserTimezone(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	utcTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{
			name:     "Valid timezone - Europe/Lisbon",
			timezone: "Europe/Lisbon",
			wantErr:  false,
		},
		{
			name:     "Valid timezone - America/New_York",
			timezone: "America/New_York",
			wantErr:  false,
		},
		{
			name:     "Empty timezone defaults to UTC",
			timezone: "",
			wantErr:  false,
		},
		{
			name:     "Invalid timezone",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ConvertToUserTimezone(ctx, utcTime, tt.timezone)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, result)
			}
		})
	}
}

func TestConvertFromUserTimezone(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	localTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.Local)

	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{
			name:     "Valid timezone - Europe/Lisbon",
			timezone: "Europe/Lisbon",
			wantErr:  false,
		},
		{
			name:     "Empty timezone",
			timezone: "",
			wantErr:  false,
		},
		{
			name:     "Invalid timezone",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ConvertFromUserTimezone(ctx, localTime, tt.timezone)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, time.UTC, result.Location())
			}
		})
	}
}

func TestGetTimezoneOffset(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	testTime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{
			name:     "UTC timezone",
			timezone: "UTC",
			wantErr:  false,
		},
		{
			name:     "Europe/Lisbon timezone",
			timezone: "Europe/Lisbon",
			wantErr:  false,
		},
		{
			name:     "Empty timezone",
			timezone: "",
			wantErr:  false,
		},
		{
			name:     "Invalid timezone",
			timezone: "Invalid/Timezone",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset, err := service.GetTimezoneOffset(ctx, tt.timezone, testTime)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.IsType(t, int(0), offset)
			}
		})
	}
}

func TestGetMessage(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	// Test getting a message
	result := service.GetMessage(ctx, LocalePortuguese, MsgWelcome)
	assert.NotEmpty(t, result)

	// Test with arguments
	result = service.GetMessage(ctx, LocaleEnglish, MsgEventReminder, "Test Event", "1 hour")
	assert.Contains(t, result, "Test Event")
	assert.Contains(t, result, "1 hour")
}

func TestGetNotificationTemplate(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	data := NotificationTemplateData{
		UserName:   "John Doe",
		EventTitle: "Friday Night Magic",
		EventDate:  "15/01/2024",
		EventTime:  "19:00",
		EventVenue: "Local Game Store",
		HostName:   "Jane Smith",
		AppName:    "MatchTCG",
		ActionURL:  "https://matchtcg.com/events/123",
	}

	tests := []struct {
		name         string
		locale       SupportedLocale
		templateType NotificationTemplateType
	}{
		{
			name:         "Portuguese RSVP confirmation",
			locale:       LocalePortuguese,
			templateType: TemplateEventRSVPConfirmation,
		},
		{
			name:         "English RSVP confirmation",
			locale:       LocaleEnglish,
			templateType: TemplateEventRSVPConfirmation,
		},
		{
			name:         "Portuguese event reminder",
			locale:       LocalePortuguese,
			templateType: TemplateEventReminder,
		},
		{
			name:         "English event reminder",
			locale:       LocaleEnglish,
			templateType: TemplateEventReminder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := service.GetNotificationTemplate(ctx, tt.locale, tt.templateType, data)

			assert.NotEmpty(t, template.Subject)
			assert.NotEmpty(t, template.Body)
			assert.Contains(t, template.Body, data.UserName)
			assert.Contains(t, template.Body, data.EventTitle)
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	service := NewI18nService()
	ctx := context.Background()

	now := time.Now()

	tests := []struct {
		name     string
		locale   SupportedLocale
		time     time.Time
		contains string
	}{
		{
			name:     "Portuguese - 30 minutes ago",
			locale:   LocalePortuguese,
			time:     now.Add(-30 * time.Minute),
			contains: "há",
		},
		{
			name:     "English - 30 minutes ago",
			locale:   LocaleEnglish,
			time:     now.Add(-30 * time.Minute),
			contains: "ago",
		},
		{
			name:     "Portuguese - in 2 hours",
			locale:   LocalePortuguese,
			time:     now.Add(2 * time.Hour),
			contains: "em",
		},
		{
			name:     "English - in 2 hours",
			locale:   LocaleEnglish,
			time:     now.Add(2 * time.Hour),
			contains: "in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.FormatRelativeTime(ctx, tt.locale, tt.time)
			assert.Contains(t, result, tt.contains)
		})
	}
}
