package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("invalid hash format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

// PasswordConfig holds configuration for password hashing
type PasswordConfig struct {
	Memory      uint32 // Memory usage in KB
	Iterations  uint32 // Number of iterations
	Parallelism uint8  // Number of threads
	SaltLength  uint32 // Length of salt in bytes
	KeyLength   uint32 // Length of generated key in bytes
}

// PasswordService handles password hashing and verification
type PasswordService struct {
	config PasswordConfig
}

// NewPasswordService creates a new password service with secure defaults
func NewPasswordService(config *PasswordConfig) *PasswordService {
	// Use secure defaults if no config provided
	if config == nil {
		config = &PasswordConfig{
			Memory:      64 * 1024, // 64 MB
			Iterations:  3,         // 3 iterations
			Parallelism: 2,         // 2 threads
			SaltLength:  16,        // 16 bytes salt
			KeyLength:   32,        // 32 bytes key
		}
	}

	return &PasswordService{
		config: *config,
	}
}

// HashPassword generates a hash for the given password
func (p *PasswordService) HashPassword(password string) (string, error) {
	// Generate random salt
	salt, err := generateRandomBytes(p.config.SaltLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate hash using Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.config.Iterations,
		p.config.Memory,
		p.config.Parallelism,
		p.config.KeyLength,
	)

	// Encode hash in format: $argon2id$v=19$m=memory,t=iterations,p=parallelism$salt$hash
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.config.Memory,
		p.config.Iterations,
		p.config.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifyPassword verifies a password against its hash
func (p *PasswordService) VerifyPassword(password, encodedHash string) (bool, error) {
	// Parse the encoded hash
	salt, hash, config, err := p.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Generate hash for the provided password using the same parameters
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		config.Iterations,
		config.Memory,
		config.Parallelism,
		config.KeyLength,
	)

	// Compare hashes using constant-time comparison
	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

// IsHashUpToDate checks if a hash was created with current parameters
func (p *PasswordService) IsHashUpToDate(encodedHash string) (bool, error) {
	_, _, config, err := p.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	return config.Memory == p.config.Memory &&
		config.Iterations == p.config.Iterations &&
		config.Parallelism == p.config.Parallelism &&
		config.SaltLength == p.config.SaltLength &&
		config.KeyLength == p.config.KeyLength, nil
}

// decodeHash parses an encoded hash and extracts parameters
func (p *PasswordService) decodeHash(encodedHash string) (salt, hash []byte, config PasswordConfig, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, PasswordConfig{}, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, PasswordConfig{}, err
	}
	if version != argon2.Version {
		return nil, nil, PasswordConfig{}, ErrIncompatibleVersion
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &config.Memory, &config.Iterations, &config.Parallelism)
	if err != nil {
		return nil, nil, PasswordConfig{}, err
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, PasswordConfig{}, err
	}
	config.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, PasswordConfig{}, err
	}
	config.KeyLength = uint32(len(hash))

	return salt, hash, config, nil
}

// generateRandomBytes generates cryptographically secure random bytes
func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ValidatePasswordStrength validates password meets minimum requirements
func (p *PasswordService) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
