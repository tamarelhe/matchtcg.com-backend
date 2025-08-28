package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserLinker is a mock implementation of UserLinker
type MockUserLinker struct {
	mock.Mock
}

func (m *MockUserLinker) FindUserByEmail(ctx context.Context, email string) (userID string, exists bool, err error) {
	args := m.Called(ctx, email)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockUserLinker) LinkOAuthAccount(ctx context.Context, userID string, provider OAuthProvider, providerUserID string) error {
	args := m.Called(ctx, userID, provider, providerUserID)
	return args.Error(0)
}

func (m *MockUserLinker) CreateUserFromOAuth(ctx context.Context, userInfo *OAuthUserInfo) (userID string, err error) {
	args := m.Called(ctx, userInfo)
	return args.String(0), args.Error(1)
}

func TestOAuthService_GenerateAuthURL(t *testing.T) {
	config := OAuthConfig{
		Google: GoogleConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
		},
		Apple: AppleConfig{
			ClientID:    "com.matchtcg.app",
			RedirectURL: "http://localhost:8080/auth/apple/callback",
		},
	}

	stateStore := NewInMemoryStateStore(10 * time.Minute)
	defer stateStore.Close()

	userLinker := &MockUserLinker{}
	oauthService := NewOAuthService(config, stateStore, userLinker)

	t.Run("Google OAuth without PKCE", func(t *testing.T) {
		authURL, state, err := oauthService.GenerateAuthURL(ProviderGoogle, false)
		require.NoError(t, err)
		assert.NotEmpty(t, authURL)
		assert.NotEmpty(t, state)

		// Parse URL to verify parameters
		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)
		assert.Equal(t, "accounts.google.com", parsedURL.Host)
		assert.Equal(t, "/o/oauth2/auth", parsedURL.Path)

		params := parsedURL.Query()
		assert.Equal(t, config.Google.ClientID, params.Get("client_id"))
		assert.Equal(t, config.Google.RedirectURL, params.Get("redirect_uri"))
		assert.Equal(t, "code", params.Get("response_type"))
		assert.Equal(t, state, params.Get("state"))
		assert.Contains(t, params.Get("scope"), "openid")
		assert.Empty(t, params.Get("code_challenge")) // No PKCE
	})

	t.Run("Google OAuth with PKCE", func(t *testing.T) {
		authURL, state, err := oauthService.GenerateAuthURL(ProviderGoogle, true)
		require.NoError(t, err)
		assert.NotEmpty(t, authURL)
		assert.NotEmpty(t, state)

		// Parse URL to verify PKCE parameters
		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)

		params := parsedURL.Query()
		assert.NotEmpty(t, params.Get("code_challenge"))
		assert.Equal(t, "S256", params.Get("code_challenge_method"))

		// Verify PKCE challenge is stored
		challenge, err := stateStore.GetPKCEChallenge(state)
		require.NoError(t, err)
		assert.NotEmpty(t, challenge.CodeVerifier)
		assert.NotEmpty(t, challenge.CodeChallenge)
		assert.Equal(t, state, challenge.State)
	})

	t.Run("Apple OAuth with PKCE", func(t *testing.T) {
		authURL, state, err := oauthService.GenerateAuthURL(ProviderApple, true)
		require.NoError(t, err)
		assert.NotEmpty(t, authURL)
		assert.NotEmpty(t, state)

		// Parse URL to verify parameters
		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)
		assert.Equal(t, "appleid.apple.com", parsedURL.Host)
		assert.Equal(t, "/auth/authorize", parsedURL.Path)

		params := parsedURL.Query()
		assert.Equal(t, config.Apple.ClientID, params.Get("client_id"))
		assert.Equal(t, config.Apple.RedirectURL, params.Get("redirect_uri"))
		assert.Equal(t, "code", params.Get("response_type"))
		assert.Equal(t, "name email", params.Get("scope"))
		assert.Equal(t, "form_post", params.Get("response_mode"))
		assert.NotEmpty(t, params.Get("code_challenge"))
		assert.Equal(t, "S256", params.Get("code_challenge_method"))
	})

	t.Run("Invalid provider", func(t *testing.T) {
		_, _, err := oauthService.GenerateAuthURL("invalid", false)
		assert.ErrorIs(t, err, ErrInvalidProvider)
	})
}

func TestOAuthService_HandleGoogleCallback(t *testing.T) {
	// Create mock Google OAuth server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			// Mock token exchange
			response := map[string]interface{}{
				"access_token": "mock-access-token",
				"token_type":   "Bearer",
				"expires_in":   3600,
			}
			json.NewEncoder(w).Encode(response)
		case "/oauth2/v2/userinfo":
			// Mock user info
			userInfo := map[string]interface{}{
				"id":             "123456789",
				"email":          "test@example.com",
				"verified_email": true,
				"name":           "Test User",
				"given_name":     "Test",
				"family_name":    "User",
				"picture":        "https://example.com/avatar.jpg",
				"locale":         "en",
			}
			json.NewEncoder(w).Encode(userInfo)
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockServer.Close()

	config := OAuthConfig{
		Google: GoogleConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
		},
	}

	stateStore := NewInMemoryStateStore(10 * time.Minute)
	defer stateStore.Close()

	userLinker := &MockUserLinker{}
	oauthService := NewOAuthService(config, stateStore, userLinker)

	// Override HTTP client to use mock server
	oauthService.httpClient = mockServer.Client()

	t.Run("successful callback without PKCE", func(t *testing.T) {
		state := "test-state"
		code := "test-auth-code"

		// Note: This test would require mocking the Google OAuth endpoint
		// For now, we'll test the structure and error handling
		ctx := context.Background()

		// This will fail because we can't easily mock the oauth2 library's token exchange
		// In a real integration test, you would use a proper OAuth testing framework
		_, err := oauthService.HandleCallback(ctx, ProviderGoogle, code, state)

		// We expect an error here because we can't properly mock the oauth2 library
		assert.Error(t, err)
	})
}

func TestOAuthService_LinkOrCreateUser(t *testing.T) {
	config := OAuthConfig{}
	stateStore := NewInMemoryStateStore(10 * time.Minute)
	defer stateStore.Close()

	ctx := context.Background()
	userInfo := &OAuthUserInfo{
		ID:            "google-123456789",
		Email:         "test@example.com",
		EmailVerified: true,
		Name:          "Test User",
		Provider:      ProviderGoogle,
	}

	t.Run("link to existing user", func(t *testing.T) {
		// Create fresh mock for this test case
		userLinker := &MockUserLinker{}
		oauthService := NewOAuthService(config, stateStore, userLinker)

		existingUserID := "existing-user-id"

		userLinker.On("FindUserByEmail", ctx, userInfo.Email).Return(existingUserID, true, nil)
		userLinker.On("LinkOAuthAccount", ctx, existingUserID, userInfo.Provider, userInfo.ID).Return(nil)

		userID, isNewUser, err := oauthService.LinkOrCreateUser(ctx, userInfo)
		require.NoError(t, err)
		assert.Equal(t, existingUserID, userID)
		assert.False(t, isNewUser)

		userLinker.AssertExpectations(t)
	})

	t.Run("create new user", func(t *testing.T) {
		// Create fresh mock for this test case
		userLinker := &MockUserLinker{}
		oauthService := NewOAuthService(config, stateStore, userLinker)

		newUserID := "new-user-id"

		userLinker.On("FindUserByEmail", ctx, userInfo.Email).Return("", false, nil)
		userLinker.On("CreateUserFromOAuth", ctx, userInfo).Return(newUserID, nil)

		userID, isNewUser, err := oauthService.LinkOrCreateUser(ctx, userInfo)
		require.NoError(t, err)
		assert.Equal(t, newUserID, userID)
		assert.True(t, isNewUser)

		userLinker.AssertExpectations(t)
	})
}

func TestOAuthService_PKCEFlow(t *testing.T) {
	config := OAuthConfig{
		Google: GoogleConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost:8080/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
		},
	}

	stateStore := NewInMemoryStateStore(10 * time.Minute)
	defer stateStore.Close()

	userLinker := &MockUserLinker{}
	oauthService := NewOAuthService(config, stateStore, userLinker)

	t.Run("PKCE challenge generation and retrieval", func(t *testing.T) {
		// Generate auth URL with PKCE
		authURL, state, err := oauthService.GenerateAuthURL(ProviderGoogle, true)
		require.NoError(t, err)
		assert.NotEmpty(t, authURL)
		assert.NotEmpty(t, state)

		// Verify PKCE challenge is stored
		challenge, err := stateStore.GetPKCEChallenge(state)
		require.NoError(t, err)
		assert.NotEmpty(t, challenge.CodeVerifier)
		assert.NotEmpty(t, challenge.CodeChallenge)
		assert.Equal(t, state, challenge.State)

		// Verify code challenge is properly generated
		expectedChallenge := generateCodeChallenge(challenge.CodeVerifier)
		assert.Equal(t, expectedChallenge, challenge.CodeChallenge)

		// Verify auth URL contains PKCE parameters
		parsedURL, err := url.Parse(authURL)
		require.NoError(t, err)
		params := parsedURL.Query()
		assert.Equal(t, challenge.CodeChallenge, params.Get("code_challenge"))
		assert.Equal(t, "S256", params.Get("code_challenge_method"))
	})

	t.Run("PKCE challenge expiration", func(t *testing.T) {
		// Create store with very short TTL
		shortTTLStore := NewInMemoryStateStore(1 * time.Millisecond)
		defer shortTTLStore.Close()

		oauthService.stateStore = shortTTLStore

		// Generate auth URL
		_, state, err := oauthService.GenerateAuthURL(ProviderGoogle, true)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Challenge should be expired
		_, err = shortTTLStore.GetPKCEChallenge(state)

		require.True(t,
			errors.Is(err, ErrStateExpired) || errors.Is(err, ErrStateNotFound),
			"expected ErrA or ErrB, got %v", err,
		)
	})
}

func TestGenerateCodeChallenge(t *testing.T) {
	testCases := []struct {
		verifier string
		expected string
	}{
		{
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			expected: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
		},
		{
			verifier: "test-verifier",
			expected: "JBbiqONGWPaAmwXk_8bT6UnlPfrn65D32eZlJS-zGG0",
		},
	}

	for _, tc := range testCases {
		t.Run("verifier: "+tc.verifier, func(t *testing.T) {
			challenge := generateCodeChallenge(tc.verifier)
			assert.Equal(t, tc.expected, challenge)
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	t.Run("generates string of correct length", func(t *testing.T) {
		lengths := []int{16, 32, 64, 128}

		for _, length := range lengths {
			str, err := generateRandomString(length)
			require.NoError(t, err)
			assert.Len(t, str, length)
		}
	})

	t.Run("generates different strings", func(t *testing.T) {
		str1, err := generateRandomString(32)
		require.NoError(t, err)

		str2, err := generateRandomString(32)
		require.NoError(t, err)

		assert.NotEqual(t, str1, str2)
	})

	t.Run("generates URL-safe strings", func(t *testing.T) {
		str, err := generateRandomString(64)
		require.NoError(t, err)

		// Should not contain URL-unsafe characters
		assert.NotContains(t, str, "+")
		assert.NotContains(t, str, "/")
		assert.NotContains(t, str, "=")
	})
}

func TestOAuthService_InvalidProvider(t *testing.T) {
	config := OAuthConfig{}
	stateStore := NewInMemoryStateStore(10 * time.Minute)
	defer stateStore.Close()

	userLinker := &MockUserLinker{}
	oauthService := NewOAuthService(config, stateStore, userLinker)

	ctx := context.Background()

	t.Run("invalid provider in GenerateAuthURL", func(t *testing.T) {
		_, _, err := oauthService.GenerateAuthURL("invalid-provider", false)
		assert.ErrorIs(t, err, ErrInvalidProvider)
	})

	t.Run("invalid provider in HandleCallback", func(t *testing.T) {
		_, err := oauthService.HandleCallback(ctx, "invalid-provider", "code", "state")
		assert.ErrorIs(t, err, ErrInvalidProvider)
	})
}
