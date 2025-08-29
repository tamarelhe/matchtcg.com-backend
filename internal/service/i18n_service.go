package service

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// SupportedLocale represents supported locales in the system
type SupportedLocale string

const (
	LocalePortuguese SupportedLocale = "pt"
	LocaleEnglish    SupportedLocale = "en"
)

// I18nService provides internationalization and localization services
type I18nService struct {
	printers map[SupportedLocale]*message.Printer
}

// NewI18nService creates a new internationalization service
func NewI18nService() *I18nService {
	service := &I18nService{
		printers: make(map[SupportedLocale]*message.Printer),
	}

	// Initialize message printers for supported locales
	service.printers[LocalePortuguese] = message.NewPrinter(language.Portuguese)
	service.printers[LocaleEnglish] = message.NewPrinter(language.English)

	return service
}

// GetSupportedLocales returns all supported locales
func (s *I18nService) GetSupportedLocales() []SupportedLocale {
	return []SupportedLocale{LocalePortuguese, LocaleEnglish}
}

// IsValidLocale checks if the given locale is supported
func (s *I18nService) IsValidLocale(locale string) bool {
	supportedLocale := SupportedLocale(locale)
	for _, supported := range s.GetSupportedLocales() {
		if supported == supportedLocale {
			return true
		}
	}
	return false
}

// GetDefaultLocale returns the default locale for the system
func (s *I18nService) GetDefaultLocale() SupportedLocale {
	return LocalePortuguese
}

// GetLocaleForCountry returns the appropriate locale based on country code
func (s *I18nService) GetLocaleForCountry(countryCode string) SupportedLocale {
	switch countryCode {
	case "PT", "BR":
		return LocalePortuguese
	default:
		return LocaleEnglish
	}
}

// FormatMessage formats a message using the specified locale
func (s *I18nService) FormatMessage(ctx context.Context, locale SupportedLocale, key string, args ...interface{}) string {
	printer, exists := s.printers[locale]
	if !exists {
		printer = s.printers[s.GetDefaultLocale()]
	}

	return printer.Sprintf(key, args...)
}

// FormatDateTime formats a datetime according to locale preferences
func (s *I18nService) FormatDateTime(ctx context.Context, locale SupportedLocale, t time.Time, timezone string) (string, error) {
	// Load the timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// Convert to user's timezone
	localTime := t.In(loc)

	// Format according to locale
	switch locale {
	case LocalePortuguese:
		return localTime.Format("02/01/2006 15:04"), nil
	case LocaleEnglish:
		return localTime.Format("01/02/2006 3:04 PM"), nil
	default:
		return localTime.Format("2006-01-02 15:04"), nil
	}
}

// FormatDate formats a date according to locale preferences
func (s *I18nService) FormatDate(ctx context.Context, locale SupportedLocale, t time.Time, timezone string) (string, error) {
	// Load the timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// Convert to user's timezone
	localTime := t.In(loc)

	// Format according to locale
	switch locale {
	case LocalePortuguese:
		return localTime.Format("02/01/2006"), nil
	case LocaleEnglish:
		return localTime.Format("01/02/2006"), nil
	default:
		return localTime.Format("2006-01-02"), nil
	}
}

// FormatTime formats a time according to locale preferences
func (s *I18nService) FormatTime(ctx context.Context, locale SupportedLocale, t time.Time, timezone string) (string, error) {
	// Load the timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// Convert to user's timezone
	localTime := t.In(loc)

	// Format according to locale
	switch locale {
	case LocalePortuguese:
		return localTime.Format("15:04"), nil
	case LocaleEnglish:
		return localTime.Format("3:04 PM"), nil
	default:
		return localTime.Format("15:04"), nil
	}
}

// ConvertToUserTimezone converts a UTC time to user's timezone
func (s *I18nService) ConvertToUserTimezone(ctx context.Context, utcTime time.Time, timezone string) (time.Time, error) {
	if timezone == "" {
		timezone = "UTC"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	return utcTime.In(loc), nil
}

// ConvertFromUserTimezone converts a time from user's timezone to UTC
func (s *I18nService) ConvertFromUserTimezone(ctx context.Context, localTime time.Time, timezone string) (time.Time, error) {
	if timezone == "" {
		return localTime.UTC(), nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	// Parse the time in the specified timezone and convert to UTC
	timeInTz := time.Date(
		localTime.Year(), localTime.Month(), localTime.Day(),
		localTime.Hour(), localTime.Minute(), localTime.Second(),
		localTime.Nanosecond(), loc,
	)

	return timeInTz.UTC(), nil
}

// GetTimezoneOffset returns the timezone offset in hours for a given timezone
func (s *I18nService) GetTimezoneOffset(ctx context.Context, timezone string, t time.Time) (int, error) {
	if timezone == "" {
		return 0, nil
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return 0, fmt.Errorf("invalid timezone %s: %w", timezone, err)
	}

	_, offset := t.In(loc).Zone()
	return offset / 3600, nil // Convert seconds to hours
}
