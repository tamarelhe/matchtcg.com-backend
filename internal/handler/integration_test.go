package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/service"
	"github.com/matchtcg/backend/internal/usecase"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// MockPasswordService is a mock implementation of PasswordService
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) VerifyPassword(password, hash string) (bool, error) {
	args := m.Called(password, hash)
	return args.Bool(0), args.Error(1)
}

func (m *MockPasswordService) ValidatePasswordStrength(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

// MockJWTService is a mock implementation of JWTService
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(userID, email string) (*service.TokenPair, error) {
	args := m.Called(userID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(token string) (*service.TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenClaims), args.Error(1)
}

func (m *MockJWTService) BlacklistToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockJWTService) RefreshTokens(refreshToken string) (*service.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenPair), args.Error(1)
}

// MockOAuthService is a mock implementation of OAuthService
type MockOAuthService struct {
	mock.Mock
}

func (m *MockOAuthService) GetGoogleUserInfo(code string) (*service.OAuthUserInfo, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.OAuthUserInfo), args.Error(1)
}

func (m *MockOAuthService) GetAppleUserInfo(code string) (*service.OAuthUserInfo, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.OAuthUserInfo), args.Error(1)
}

func (m *MockOAuthService) GenerateAuthURL(provider service.OAuthProvider, includeState bool) (string, string, error) {
	args := m.Called(provider, includeState)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockOAuthService) HandleCallback(ctx context.Context, provider service.OAuthProvider, code, state string) (*service.OAuthUserInfo, error) {
	args := m.Called(ctx, provider, code, state)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.OAuthUserInfo), args.Error(1)
}

func (m *MockOAuthService) LinkOrCreateUser(ctx context.Context, userInfo *service.OAuthUserInfo) (string, bool, error) {
	args := m.Called(ctx, userInfo)
	return args.String(0), args.Bool(1), args.Error(2)
}

// MockRegisterUserUseCase is a mock implementation of RegisterUserUseCase
type MockRegisterUserUseCase struct {
	mock.Mock
}

func (m *MockRegisterUserUseCase) Execute(ctx context.Context, req *usecase.RegisterUserRequest) (*usecase.RegisterUserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.RegisterUserResponse), args.Error(1)
}

// TestAuthHandlerIntegration tests the authentication handler endpoints
func TestAuthHandlerIntegration(t *testing.T) {
	// Setup mocks
	mockUserRepo := &MockUserRepository{}
	mockPasswordService := &MockPasswordService{}
	mockJWTService := &MockJWTService{}
	mockOAuthService := &MockOAuthService{}
	mockRegisterUseCase := &MockRegisterUserUseCase{}

	// Create handler
	authHandler := NewAuthHandler(
		mockRegisterUseCase,
		mockJWTService,
		mockOAuthService,
		mockPasswordService,
		mockUserRepo,
	)

	// Setup router
	router := mux.NewRouter()
	authHandler.RegisterRoutes(router)

	t.Run("Register User Success", func(t *testing.T) {
		// Setup expectations
		mockPasswordService.On("ValidatePasswordStrength", "TestPass123!").Return(nil)

		user := &domain.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}
		profile := &domain.Profile{
			UserID:      user.ID,
			DisplayName: stringPtr("Test User"),
			Locale:      "en",
		}

		mockRegisterUseCase.On("Execute", mock.Anything, mock.AnythingOfType("*usecase.RegisterUserRequest")).Return(
			&usecase.RegisterUserResponse{
				User:    user,
				Profile: profile,
			}, nil)

		tokenPair := &service.TokenPair{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresAt:    time.Now().Add(15 * time.Minute),
		}
		mockJWTService.On("GenerateTokenPair", user.ID.String(), user.Email).Return(tokenPair, nil)

		// Create request
		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "TestPass123!",
			Timezone: "UTC",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusCreated, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "access_token", response.AccessToken)
		assert.Equal(t, "refresh_token", response.RefreshToken)
		assert.Equal(t, user.ID.String(), response.User.ID)
		assert.Equal(t, user.Email, response.User.Email)

		// Verify mocks
		mockPasswordService.AssertExpectations(t)
		mockRegisterUseCase.AssertExpectations(t)
		mockJWTService.AssertExpectations(t)
	})

	t.Run("Login User Success", func(t *testing.T) {
		// Reset mocks
		mockUserRepo.ExpectedCalls = nil
		mockPasswordService.ExpectedCalls = nil
		mockJWTService.ExpectedCalls = nil

		// Setup expectations
		user := &domain.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			IsActive:     true,
		}

		mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
		mockPasswordService.On("VerifyPassword", "TestPass123!", "hashed_password").Return(true, nil)

		tokenPair := &service.TokenPair{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresAt:    time.Now().Add(15 * time.Minute),
		}
		mockJWTService.On("GenerateTokenPair", user.ID.String(), user.Email).Return(tokenPair, nil)

		// Create request
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "TestPass123!",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "access_token", response.AccessToken)
		assert.Equal(t, "refresh_token", response.RefreshToken)
		assert.Equal(t, user.ID.String(), response.User.ID)
		assert.Equal(t, user.Email, response.User.Email)

		// Verify mocks
		mockUserRepo.AssertExpectations(t)
		mockPasswordService.AssertExpectations(t)
		mockJWTService.AssertExpectations(t)
	})

	t.Run("Login Invalid Credentials", func(t *testing.T) {
		// Reset mocks
		mockUserRepo.ExpectedCalls = nil
		mockPasswordService.ExpectedCalls = nil

		// Setup expectations - user not found
		mockUserRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)

		// Create request
		reqBody := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "TestPass123!",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_credentials", response.Error)

		// Verify mocks
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Logout Success", func(t *testing.T) {
		// Reset mocks
		mockJWTService.ExpectedCalls = nil

		// Setup expectations
		mockJWTService.On("BlacklistToken", "valid_token").Return(nil)

		// Create request
		req := httptest.NewRequest("POST", "/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer valid_token")
		w := httptest.NewRecorder()

		// Execute request
		router.ServeHTTP(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Successfully logged out", response["message"])

		// Verify mocks
		mockJWTService.AssertExpectations(t)
	})
}

// TestErrorHandling tests error handling across handlers
func TestErrorHandling(t *testing.T) {
	// Setup mocks
	mockUserRepo := &MockUserRepository{}
	mockPasswordService := &MockPasswordService{}
	mockJWTService := &MockJWTService{}
	mockOAuthService := &MockOAuthService{}
	mockRegisterUseCase := &MockRegisterUserUseCase{}

	// Create handler
	authHandler := NewAuthHandler(
		mockRegisterUseCase,
		mockJWTService,
		mockOAuthService,
		mockPasswordService,
		mockUserRepo,
	)

	// Setup router
	router := mux.NewRouter()
	authHandler.RegisterRoutes(router)

	t.Run("Invalid JSON Body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		reqBody := RegisterRequest{
			// Missing required fields
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "validation_error", response.Error)
	})

	t.Run("Weak Password", func(t *testing.T) {
		mockPasswordService.On("ValidatePasswordStrength", "weak").Return(assert.AnError)

		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "weak",
			Timezone: "UTC",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "weak_password", response.Error)

		mockPasswordService.AssertExpectations(t)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
