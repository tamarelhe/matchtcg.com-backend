package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) CreateTx(ctx context.Context, tx pgx.Tx, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) CreateProfile(ctx context.Context, profile *domain.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserRepository) CreateUserWithProfile(ctx context.Context, user *domain.User, profile *domain.Profile) error {
	args := m.Called(ctx, user, profile)
	return args.Error(0)
}

func (m *MockUserRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Profile), args.Error(1)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*domain.UserWithProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserWithProfile), args.Error(1)
}

func (m *MockUserRepository) ExportUserData(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockUserRepository) DeleteUserData(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	args := m.Called(ctx, userID, loginTime)
	return args.Error(0)
}

func (m *MockUserRepository) SetActive(ctx context.Context, userID uuid.UUID, active bool) error {
	args := m.Called(ctx, userID, active)
	return args.Error(0)
}

// MockPasswordHasher is a mock implementation of PasswordHasher
type MockPasswordHasher struct {
	mock.Mock
}

func (m *MockPasswordHasher) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordHasher) VerifyPassword(password, hash string) (bool, error) {
	args := m.Called(password, hash)
	return args.Bool(0), args.Error(1)
}

func TestRegisterUserUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	useCase := NewRegisterUserUseCase(mockUserRepo, mockPasswordHasher)

	req := &RegisterUserRequest{
		Email:       "test@example.com",
		Password:    "password123",
		DisplayName: "Test User",
		Locale:      "en",
		Timezone:    "UTC",
		Country:     "Portugal",
		City:        "Lisbon",
	}

	// Mock expectations
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return((*domain.User)(nil), errors.New("not found"))
	mockPasswordHasher.On("HashPassword", req.Password).Return("hashed_password", nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	mockUserRepo.On("CreateProfile", mock.Anything, mock.AnythingOfType("*domain.Profile")).Return(nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, req.Email, result.User.Email)
	assert.Equal(t, "hashed_password", result.User.PasswordHash)
	assert.Equal(t, req.DisplayName, *result.Profile.DisplayName)
	assert.Equal(t, req.Locale, result.Profile.Locale)
	assert.Equal(t, req.Timezone, result.Profile.Timezone)
	assert.Equal(t, req.Country, *result.Profile.Country)
	assert.Equal(t, req.City, *result.Profile.City)
	assert.True(t, result.User.IsActive)

	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
}

func TestRegisterUserUseCase_Execute_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	useCase := NewRegisterUserUseCase(mockUserRepo, mockPasswordHasher)

	req := &RegisterUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Timezone: "UTC",
	}

	existingUser := &domain.User{
		ID:    uuid.New(),
		Email: req.Email,
	}

	// Mock expectations
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return(existingUser, nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrEmailAlreadyExists, err)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUserUseCase_Execute_InvalidEmail(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	useCase := NewRegisterUserUseCase(mockUserRepo, mockPasswordHasher)

	req := &RegisterUserRequest{
		Email:    "invalid-email",
		Password: "password123",
		Timezone: "UTC",
	}

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidEmail, err)
	assert.Nil(t, result)
}

func TestRegisterUserUseCase_Execute_WeakPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	useCase := NewRegisterUserUseCase(mockUserRepo, mockPasswordHasher)

	req := &RegisterUserRequest{
		Email:    "test@example.com",
		Password: "weak",
		Timezone: "UTC",
	}

	// Mock expectations
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return((*domain.User)(nil), errors.New("not found"))

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrWeakPassword, err)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUserUseCase_Execute_DefaultLocale(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	useCase := NewRegisterUserUseCase(mockUserRepo, mockPasswordHasher)

	req := &RegisterUserRequest{
		DisplayName: "Test",
		Email:       "test@example.com",
		Password:    "password123",
		Timezone:    "UTC",
		// Locale not provided - should default to "pt"
	}

	// Mock expectations
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return((*domain.User)(nil), errors.New("not found"))
	mockPasswordHasher.On("HashPassword", req.Password).Return("hashed_password", nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	mockUserRepo.On("CreateProfile", mock.Anything, mock.AnythingOfType("*domain.Profile")).Return(nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "pt", result.Profile.Locale) // Should default to Portuguese

	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
}

func TestRegisterUserUseCase_Execute_PasswordHashingError(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockPasswordHasher := new(MockPasswordHasher)
	useCase := NewRegisterUserUseCase(mockUserRepo, mockPasswordHasher)

	req := &RegisterUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Timezone: "UTC",
	}

	// Mock expectations
	mockUserRepo.On("GetByEmail", mock.Anything, req.Email).Return((*domain.User)(nil), errors.New("not found"))
	mockPasswordHasher.On("HashPassword", req.Password).Return("", errors.New("hashing failed"))

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "hashing failed", err.Error())
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
	mockPasswordHasher.AssertExpectations(t)
}
