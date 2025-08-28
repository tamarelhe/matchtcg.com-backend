package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	LastLogin    *time.Time `json:"last_login,omitempty" db:"last_login"`
}

// Profile represents a user's profile information
type Profile struct {
	UserID                   uuid.UUID              `json:"user_id" db:"user_id"`
	DisplayName              *string                `json:"display_name,omitempty" db:"display_name"`
	Locale                   string                 `json:"locale" db:"locale"`
	Timezone                 string                 `json:"timezone" db:"timezone"`
	Country                  *string                `json:"country,omitempty" db:"country"`
	City                     *string                `json:"city,omitempty" db:"city"`
	PreferredGames           []string               `json:"preferred_games" db:"preferred_games"`
	CommunicationPreferences map[string]interface{} `json:"communication_preferences" db:"communication_preferences"`
	VisibilitySettings       map[string]interface{} `json:"visibility_settings" db:"visibility_settings"`
	UpdatedAt                time.Time              `json:"updated_at" db:"updated_at"`
}

// UserWithProfile represents a user with their profile
type UserWithProfile struct {
	User
	Profile *Profile `json:"profile,omitempty"`
}

var (
	ErrInvalidEmail       = errors.New("invalid email format")
	ErrEmptyEmail         = errors.New("email cannot be empty")
	ErrEmptyPasswordHash  = errors.New("password hash cannot be empty")
	ErrInvalidDisplayName = errors.New("display name must be between 1 and 100 characters")
	ErrInvalidLocale      = errors.New("locale must be 'en' or 'pt'")
	ErrInvalidTimezone    = errors.New("timezone cannot be empty")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validate validates the User entity
func (u *User) Validate() error {
	if u.Email == "" {
		return ErrEmptyEmail
	}

	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmail
	}

	if u.PasswordHash == "" {
		return ErrEmptyPasswordHash
	}

	return nil
}

// Validate validates the Profile entity
func (p *Profile) Validate() error {
	if p.DisplayName != nil {
		displayName := strings.TrimSpace(*p.DisplayName)
		if len(displayName) == 0 || len(displayName) > 100 {
			return ErrInvalidDisplayName
		}
	}

	if p.Locale != "en" && p.Locale != "pt" {
		return ErrInvalidLocale
	}

	if p.Timezone == "" {
		return ErrInvalidTimezone
	}

	return nil
}

// IsValidEmail checks if an email address is valid
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
