package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenBlacklisted = errors.New("token blacklisted")
)

// TokenClaims represents the JWT claims structure
type TokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// JWTService handles JWT token operations
type JWTService struct {
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
	accessTTL      time.Duration
	refreshTTL     time.Duration
	blacklistStore BlacklistStore
}

// BlacklistStore interface for token blacklisting
type BlacklistStore interface {
	IsBlacklisted(tokenID string) (bool, error)
	BlacklistToken(tokenID string, expiresAt time.Time) error
}

// JWTConfig holds configuration for JWT service
type JWTConfig struct {
	PrivateKeyPEM  string
	PublicKeyPEM   string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	BlacklistStore BlacklistStore
}

// NewJWTService creates a new JWT service instance
func NewJWTService(config JWTConfig) (*JWTService, error) {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	var err error

	if config.PrivateKeyPEM != "" && config.PublicKeyPEM != "" {
		// Parse provided keys
		privateKey, err = parsePrivateKey(config.PrivateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		publicKey, err = parsePublicKey(config.PublicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	} else {
		// Generate new key pair for development
		privateKey, publicKey, err = generateKeyPair()
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}
	}

	// Set default TTLs if not provided
	accessTTL := config.AccessTTL
	if accessTTL == 0 {
		accessTTL = 15 * time.Minute
	}

	refreshTTL := config.RefreshTTL
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour // 7 days
	}

	return &JWTService{
		privateKey:     privateKey,
		publicKey:      publicKey,
		accessTTL:      accessTTL,
		refreshTTL:     refreshTTL,
		blacklistStore: config.BlacklistStore,
	}, nil
}

// GenerateTokenPair creates a new access and refresh token pair
func (j *JWTService) GenerateTokenPair(userID, email string) (*TokenPair, error) {
	now := time.Now()
	accessTokenID := uuid.New().String()
	refreshTokenID := uuid.New().String()

	// Create access token
	accessClaims := TokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessTokenID,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTTL)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "matchtcg-backend",
			Audience:  []string{"matchtcg-app"},
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token
	refreshClaims := TokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshTokenID,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTTL)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "matchtcg-backend",
			Audience:  []string{"matchtcg-refresh"},
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(j.accessTTL),
	}, nil
}

// ValidateAccessToken validates an access token and returns claims
func (j *JWTService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	return j.validateToken(tokenString, "matchtcg-app")
}

// ValidateRefreshToken validates a refresh token and returns claims
func (j *JWTService) ValidateRefreshToken(tokenString string) (*TokenClaims, error) {
	return j.validateToken(tokenString, "matchtcg-refresh")
}

// RefreshTokens creates a new token pair using a valid refresh token
func (j *JWTService) RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	// Validate refresh token
	claims, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Blacklist the old refresh token
	if j.blacklistStore != nil {
		err = j.blacklistStore.BlacklistToken(claims.ID, claims.ExpiresAt.Time)
		if err != nil {
			return nil, fmt.Errorf("failed to blacklist old refresh token: %w", err)
		}
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Email)
}

// BlacklistToken adds a token to the blacklist
func (j *JWTService) BlacklistToken(tokenString string) error {
	if j.blacklistStore == nil {
		return nil // No blacklist store configured
	}

	// Parse token to get ID and expiration
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token for blacklisting: %w", err)
	}

	if claims, ok := token.Claims.(*TokenClaims); ok {
		return j.blacklistStore.BlacklistToken(claims.ID, claims.ExpiresAt.Time)
	}

	return errors.New("invalid token claims")
}

// validateToken validates a token and checks blacklist
func (j *JWTService) validateToken(tokenString, expectedAudience string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify audience
	audienceFound := false
	for _, aud := range claims.Audience {
		if aud == expectedAudience {
			audienceFound = true
			break
		}
	}
	if !audienceFound {
		return nil, ErrInvalidToken
	}

	// Check blacklist
	if j.blacklistStore != nil {
		blacklisted, err := j.blacklistStore.IsBlacklisted(claims.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check blacklist: %w", err)
		}
		if blacklisted {
			return nil, ErrTokenBlacklisted
		}
	}

	return claims, nil
}

// generateKeyPair generates a new RSA key pair for development
func generateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

// parsePrivateKey parses a PEM-encoded RSA private key
func parsePrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8 format
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rsaKey, ok := parsedKey.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not an RSA private key")
		}
		return rsaKey, nil
	}
	return key, nil
}

// parsePublicKey parses a PEM-encoded RSA public key
func parsePublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}
	return rsaKey, nil
}
