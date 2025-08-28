package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr error
	}{
		{
			name: "valid user",
			user: User{
				ID:           uuid.New(),
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
				IsActive:     true,
			},
			wantErr: nil,
		},
		{
			name: "empty email",
			user: User{
				ID:           uuid.New(),
				Email:        "",
				PasswordHash: "hashed_password",
			},
			wantErr: ErrEmptyEmail,
		},
		{
			name: "invalid email format",
			user: User{
				ID:           uuid.New(),
				Email:        "invalid-email",
				PasswordHash: "hashed_password",
			},
			wantErr: ErrInvalidEmail,
		},
		{
			name: "empty password hash",
			user: User{
				ID:    uuid.New(),
				Email: "test@example.com",
			},
			wantErr: ErrEmptyPasswordHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if err != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfile_Validate(t *testing.T) {
	validDisplayName := "John Doe"
	longDisplayName := "This is a very long display name that exceeds the maximum allowed length of 100 characters and should fail validation"
	emptyDisplayName := ""

	tests := []struct {
		name    string
		profile Profile
		wantErr error
	}{
		{
			name: "valid profile",
			profile: Profile{
				UserID:      uuid.New(),
				DisplayName: &validDisplayName,
				Locale:      "en",
				Timezone:    "UTC",
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid profile with Portuguese locale",
			profile: Profile{
				UserID:      uuid.New(),
				DisplayName: &validDisplayName,
				Locale:      "pt",
				Timezone:    "Europe/Lisbon",
				UpdatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "invalid locale",
			profile: Profile{
				UserID:      uuid.New(),
				DisplayName: &validDisplayName,
				Locale:      "fr",
				Timezone:    "UTC",
			},
			wantErr: ErrInvalidLocale,
		},
		{
			name: "empty timezone",
			profile: Profile{
				UserID:      uuid.New(),
				DisplayName: &validDisplayName,
				Locale:      "en",
				Timezone:    "",
			},
			wantErr: ErrInvalidTimezone,
		},
		{
			name: "display name too long",
			profile: Profile{
				UserID:      uuid.New(),
				DisplayName: &longDisplayName,
				Locale:      "en",
				Timezone:    "UTC",
			},
			wantErr: ErrInvalidDisplayName,
		},
		{
			name: "empty display name",
			profile: Profile{
				UserID:      uuid.New(),
				DisplayName: &emptyDisplayName,
				Locale:      "en",
				Timezone:    "UTC",
			},
			wantErr: ErrInvalidDisplayName,
		},
		{
			name: "nil display name (valid)",
			profile: Profile{
				UserID:    uuid.New(),
				Locale:    "en",
				Timezone:  "UTC",
				UpdatedAt: time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if err != tt.wantErr {
				t.Errorf("Profile.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"valid email with numbers", "user123@example123.com", true},
		{"valid email with special chars", "user.name+tag@example.com", true},
		{"invalid email - no @", "testexample.com", false},
		{"invalid email - no domain", "test@", false},
		{"invalid email - no local part", "@example.com", false},
		{"invalid email - spaces", "test @example.com", false},
		{"invalid email - no TLD", "test@example", false},
		{"empty email", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidEmail(tt.email); got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
