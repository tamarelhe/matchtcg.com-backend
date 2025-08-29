package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrWeakPassword       = errors.New("password does not meet security requirements")
)

// PasswordHasher defines the interface for password hashing operations
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hash string) (bool, error)
}

// RegisterUserRequest represents the request to register a new user
type RegisterUserRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name,omitempty" validate:"max=100"`
	Locale      string `json:"locale,omitempty" validate:"oneof=en pt"`
	Timezone    string `json:"timezone" validate:"required"`
	Country     string `json:"country,omitempty"`
	City        string `json:"city,omitempty"`
}

// RegisterUserResponse represents the response after successful registration
type RegisterUserResponse struct {
	User    *domain.User    `json:"user"`
	Profile *domain.Profile `json:"profile"`
}

// RegisterUserUseCase handles user registration
type RegisterUserUseCase struct {
	userRepo       repository.UserRepository
	passwordHasher PasswordHasher
	db             *pgxpool.Pool
}

// NewRegisterUserUseCase creates a new RegisterUserUseCase
func NewRegisterUserUseCase(userRepo repository.UserRepository, passwordHasher PasswordHasher) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

// Execute registers a new user with the provided information
func (uc *RegisterUserUseCase) Execute(ctx context.Context, req *RegisterUserRequest) (*RegisterUserResponse, error) {
	// Validate email format
	if !domain.IsValidEmail(req.Email) {
		return nil, domain.ErrInvalidEmail
	}

	// Check if user already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Validate password strength
	if err := uc.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := uc.passwordHasher.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	user := &domain.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		IsActive:     true,
	}

	// Validate user entity
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Set default values for profile
	locale := req.Locale
	if locale == "" {
		locale = "pt" // Default to Portuguese for MVP
	}

	// Create profile
	profile := &domain.Profile{
		UserID:                   user.ID,
		DisplayName:              &req.DisplayName,
		Locale:                   locale,
		Timezone:                 req.Timezone,
		Country:                  &req.Country,
		City:                     &req.City,
		PreferredGames:           []string{},
		CommunicationPreferences: make(map[string]interface{}),
		VisibilitySettings:       uc.getDefaultVisibilitySettings(),
		UpdatedAt:                time.Now().UTC(),
	}

	// Set default communication preferences
	profile.CommunicationPreferences = map[string]interface{}{
		"email_notifications": true,
		"event_reminders":     true,
		"group_invitations":   true,
		"rsvp_confirmations":  true,
		"event_updates":       true,
		"marketing_emails":    false,
	}

	// Validate profile
	if err := profile.Validate(); err != nil {
		return nil, err
	}

	// Create user + profile in repository
	if err := uc.userRepo.CreateUserWithProfile(ctx, user, profile); err != nil {
		return nil, err
	}

	return &RegisterUserResponse{
		User:    user,
		Profile: profile,
	}, nil
}

// validatePassword checks if password meets security requirements
func (uc *RegisterUserUseCase) validatePassword(password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	// Additional password complexity checks can be added here
	// For now, we just check minimum length
	return nil
}

// getDefaultVisibilitySettings returns default privacy settings for new users
func (uc *RegisterUserUseCase) getDefaultVisibilitySettings() map[string]interface{} {
	return map[string]interface{}{
		"profile_visibility":    "public",
		"show_email":            false,
		"show_real_name":        true,
		"show_location":         true,
		"show_preferred_games":  true,
		"show_event_attendance": true,
		"allow_group_invites":   true,
		"allow_event_invites":   true,
	}
}
