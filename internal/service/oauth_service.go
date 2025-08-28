package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	ErrInvalidProvider   = errors.New("invalid OAuth provider")
	ErrInvalidState      = errors.New("invalid OAuth state")
	ErrInvalidCode       = errors.New("invalid authorization code")
	ErrUserInfoFetch     = errors.New("failed to fetch user info")
	ErrAccountLinking    = errors.New("account linking failed")
	ErrInvalidAppleToken = errors.New("invalid Apple ID token")
)

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	ProviderGoogle OAuthProvider = "google"
	ProviderApple  OAuthProvider = "apple"
)

// OAuthUserInfo represents user information from OAuth providers
type OAuthUserInfo struct {
	ID            string        `json:"id"`
	Email         string        `json:"email"`
	EmailVerified bool          `json:"email_verified"`
	Name          string        `json:"name"`
	GivenName     string        `json:"given_name"`
	FamilyName    string        `json:"family_name"`
	Picture       string        `json:"picture"`
	Locale        string        `json:"locale"`
	Provider      OAuthProvider `json:"provider"`
}

// PKCEChallenge represents PKCE challenge data
type PKCEChallenge struct {
	CodeVerifier  string    `json:"code_verifier"`
	CodeChallenge string    `json:"code_challenge"`
	State         string    `json:"state"`
	CreatedAt     time.Time `json:"created_at"`
}

// OAuthConfig holds configuration for OAuth providers
type OAuthConfig struct {
	Google GoogleConfig `json:"google"`
	Apple  AppleConfig  `json:"apple"`
}

// GoogleConfig holds Google OAuth configuration
type GoogleConfig struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
}

// AppleConfig holds Apple OAuth configuration
type AppleConfig struct {
	ClientID    string `json:"client_id"`
	TeamID      string `json:"team_id"`
	KeyID       string `json:"key_id"`
	PrivateKey  string `json:"private_key"`
	RedirectURL string `json:"redirect_url"`
}

// OAuthService handles OAuth authentication flows
type OAuthService struct {
	config     OAuthConfig
	stateStore StateStore
	userLinker UserLinker
	httpClient *http.Client
}

// StateStore interface for storing PKCE challenges and state
type StateStore interface {
	StorePKCEChallenge(state string, challenge PKCEChallenge) error
	GetPKCEChallenge(state string) (*PKCEChallenge, error)
	DeletePKCEChallenge(state string) error
}

// UserLinker interface for linking OAuth accounts with existing users
type UserLinker interface {
	FindUserByEmail(ctx context.Context, email string) (userID string, exists bool, err error)
	LinkOAuthAccount(ctx context.Context, userID string, provider OAuthProvider, providerUserID string) error
	CreateUserFromOAuth(ctx context.Context, userInfo *OAuthUserInfo) (userID string, err error)
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(config OAuthConfig, stateStore StateStore, userLinker UserLinker) *OAuthService {
	return &OAuthService{
		config:     config,
		stateStore: stateStore,
		userLinker: userLinker,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateAuthURL generates OAuth authorization URL with PKCE
func (o *OAuthService) GenerateAuthURL(provider OAuthProvider, usePKCE bool) (string, string, error) {
	var oauthConfig *oauth2.Config
	var authURL string

	switch provider {
	case ProviderGoogle:
		oauthConfig = &oauth2.Config{
			ClientID:     o.config.Google.ClientID,
			ClientSecret: o.config.Google.ClientSecret,
			RedirectURL:  o.config.Google.RedirectURL,
			Scopes:       o.config.Google.Scopes,
			Endpoint:     google.Endpoint,
		}
	case ProviderApple:
		// Apple Sign In uses a different flow
		return o.generateAppleAuthURL(usePKCE)
	default:
		return "", "", ErrInvalidProvider
	}

	// Generate state parameter
	state, err := generateRandomString(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	var opts []oauth2.AuthCodeOption

	if usePKCE {
		// Generate PKCE challenge
		codeVerifier, err := generateRandomString(128)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate code verifier: %w", err)
		}

		codeChallenge := generateCodeChallenge(codeVerifier)

		// Store PKCE challenge
		challenge := PKCEChallenge{
			CodeVerifier:  codeVerifier,
			CodeChallenge: codeChallenge,
			State:         state,
			CreatedAt:     time.Now(),
		}

		err = o.stateStore.StorePKCEChallenge(state, challenge)
		if err != nil {
			return "", "", fmt.Errorf("failed to store PKCE challenge: %w", err)
		}

		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	authURL = oauthConfig.AuthCodeURL(state, opts...)
	return authURL, state, nil
}

// HandleCallback processes OAuth callback and returns user information
func (o *OAuthService) HandleCallback(ctx context.Context, provider OAuthProvider, code, state string) (*OAuthUserInfo, error) {
	switch provider {
	case ProviderGoogle:
		return o.handleGoogleCallback(ctx, code, state)
	case ProviderApple:
		return o.handleAppleCallback(ctx, code, state)
	default:
		return nil, ErrInvalidProvider
	}
}

// LinkOrCreateUser links OAuth account to existing user or creates new user
func (o *OAuthService) LinkOrCreateUser(ctx context.Context, userInfo *OAuthUserInfo) (userID string, isNewUser bool, err error) {
	// Try to find existing user by email
	existingUserID, exists, err := o.userLinker.FindUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return "", false, fmt.Errorf("failed to find user by email: %w", err)
	}

	if exists {
		// Link OAuth account to existing user
		err = o.userLinker.LinkOAuthAccount(ctx, existingUserID, userInfo.Provider, userInfo.ID)
		if err != nil {
			return "", false, fmt.Errorf("failed to link OAuth account: %w", err)
		}
		return existingUserID, false, nil
	}

	// Create new user from OAuth info
	newUserID, err := o.userLinker.CreateUserFromOAuth(ctx, userInfo)
	if err != nil {
		return "", false, fmt.Errorf("failed to create user from OAuth: %w", err)
	}

	return newUserID, true, nil
}

// handleGoogleCallback processes Google OAuth callback
func (o *OAuthService) handleGoogleCallback(ctx context.Context, code, state string) (*OAuthUserInfo, error) {
	oauthConfig := &oauth2.Config{
		ClientID:     o.config.Google.ClientID,
		ClientSecret: o.config.Google.ClientSecret,
		RedirectURL:  o.config.Google.RedirectURL,
		Scopes:       o.config.Google.Scopes,
		Endpoint:     google.Endpoint,
	}

	// Check if PKCE was used
	challenge, err := o.stateStore.GetPKCEChallenge(state)
	var opts []oauth2.AuthCodeOption

	if err == nil && challenge != nil {
		// PKCE was used, include code verifier
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", challenge.CodeVerifier))
		// Clean up stored challenge
		defer o.stateStore.DeletePKCEChallenge(state)
	}

	// Exchange code for token
	token, err := oauthConfig.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Fetch user info
	userInfo, err := o.fetchGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	userInfo.Provider = ProviderGoogle
	return userInfo, nil
}

// handleAppleCallback processes Apple OAuth callback
func (o *OAuthService) handleAppleCallback(ctx context.Context, code, state string) (*OAuthUserInfo, error) {
	// Check PKCE challenge
	challenge, err := o.stateStore.GetPKCEChallenge(state)
	if err != nil {
		return nil, fmt.Errorf("invalid state or PKCE challenge: %w", err)
	}
	defer o.stateStore.DeletePKCEChallenge(state)

	// Exchange code for token with Apple
	token, err := o.exchangeAppleCode(ctx, code, challenge.CodeVerifier)
	if err != nil {
		return nil, err
	}

	// Parse Apple ID token
	userInfo, err := o.parseAppleIDToken(token)
	if err != nil {
		return nil, err
	}

	userInfo.Provider = ProviderApple
	return userInfo, nil
}

// fetchGoogleUserInfo fetches user information from Google
func (o *OAuthService) fetchGoogleUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo OAuthUserInfo
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}

// generateAppleAuthURL generates Apple Sign In authorization URL
func (o *OAuthService) generateAppleAuthURL(usePKCE bool) (string, string, error) {
	state, err := generateRandomString(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	params := url.Values{}
	params.Set("client_id", o.config.Apple.ClientID)
	params.Set("redirect_uri", o.config.Apple.RedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "name email")
	params.Set("response_mode", "form_post")
	params.Set("state", state)

	if usePKCE {
		codeVerifier, err := generateRandomString(128)
		if err != nil {
			return "", "", fmt.Errorf("failed to generate code verifier: %w", err)
		}

		codeChallenge := generateCodeChallenge(codeVerifier)

		challenge := PKCEChallenge{
			CodeVerifier:  codeVerifier,
			CodeChallenge: codeChallenge,
			State:         state,
			CreatedAt:     time.Now(),
		}

		err = o.stateStore.StorePKCEChallenge(state, challenge)
		if err != nil {
			return "", "", fmt.Errorf("failed to store PKCE challenge: %w", err)
		}

		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	authURL := "https://appleid.apple.com/auth/authorize?" + params.Encode()
	return authURL, state, nil
}

// exchangeAppleCode exchanges authorization code for Apple ID token
func (o *OAuthService) exchangeAppleCode(ctx context.Context, code, codeVerifier string) (string, error) {
	// This is a simplified implementation
	// In production, you would need to create a proper JWT client secret for Apple
	// using the team ID, key ID, and private key

	data := url.Values{}
	data.Set("client_id", o.config.Apple.ClientID)
	data.Set("client_secret", o.generateAppleClientSecret()) // This needs proper JWT implementation
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", o.config.Apple.RedirectURL)

	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://appleid.apple.com/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Apple token endpoint returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}

	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	return tokenResp.IDToken, nil
}

// parseAppleIDToken parses Apple ID token and extracts user info
func (o *OAuthService) parseAppleIDToken(idToken string) (*OAuthUserInfo, error) {
	// This is a simplified implementation
	// In production, you would need to properly verify the JWT signature
	// using Apple's public keys

	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidAppleToken
	}

	// Decode payload (base64url)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
	}

	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	return &OAuthUserInfo{
		ID:            claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified == "true",
		Name:          claims.Name,
	}, nil
}

// generateAppleClientSecret generates JWT client secret for Apple
func (o *OAuthService) generateAppleClientSecret() string {
	// This is a placeholder - in production you would generate a proper JWT
	// using the Apple private key, team ID, and key ID
	return "placeholder-client-secret"
}

// generateRandomString generates a cryptographically secure random string
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

// generateCodeChallenge generates PKCE code challenge from verifier
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
