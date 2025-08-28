package domain

import (
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
