package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordService_HashPassword(t *testing.T) {
	passwordService := NewPasswordService(nil)
	password := "TestPassword123!"

	hash, err := passwordService.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.True(t, strings.HasPrefix(hash, "$argon2id$"))

	// Hash should be different each time due to random salt
	hash2, err := passwordService.HashPassword(password)
	require.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestPasswordService_VerifyPassword(t *testing.T) {
	passwordService := NewPasswordService(nil)
	password := "TestPassword123!"

	hash, err := passwordService.HashPassword(password)
	require.NoError(t, err)

	t.Run("correct password", func(t *testing.T) {
		valid, err := passwordService.VerifyPassword(password, hash)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("incorrect password", func(t *testing.T) {
		valid, err := passwordService.VerifyPassword("WrongPassword", hash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("empty password", func(t *testing.T) {
		valid, err := passwordService.VerifyPassword("", hash)
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("invalid hash format", func(t *testing.T) {
		_, err := passwordService.VerifyPassword(password, "invalid-hash")
		assert.ErrorIs(t, err, ErrInvalidHash)
	})
}

func TestPasswordService_CustomConfig(t *testing.T) {
	config := &PasswordConfig{
		Memory:      32 * 1024, // 32 MB
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  8,
		KeyLength:   16,
	}

	passwordService := NewPasswordService(config)
	password := "TestPassword123!"

	hash, err := passwordService.HashPassword(password)
	require.NoError(t, err)

	// Verify the hash contains the custom parameters
	assert.Contains(t, hash, "m=32768,t=2,p=1")

	// Verify password works with custom config
	valid, err := passwordService.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestPasswordService_IsHashUpToDate(t *testing.T) {
	config := &PasswordConfig{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	passwordService := NewPasswordService(config)
	password := "TestPassword123!"

	t.Run("current hash", func(t *testing.T) {
		hash, err := passwordService.HashPassword(password)
		require.NoError(t, err)

		upToDate, err := passwordService.IsHashUpToDate(hash)
		require.NoError(t, err)
		assert.True(t, upToDate)
	})

	t.Run("outdated hash", func(t *testing.T) {
		// Create hash with different parameters
		oldService := NewPasswordService(&PasswordConfig{
			Memory:      32 * 1024, // Different memory
			Iterations:  2,         // Different iterations
			Parallelism: 1,         // Different parallelism
			SaltLength:  8,         // Different salt length
			KeyLength:   16,        // Different key length
		})

		oldHash, err := oldService.HashPassword(password)
		require.NoError(t, err)

		upToDate, err := passwordService.IsHashUpToDate(oldHash)
		require.NoError(t, err)
		assert.False(t, upToDate)
	})
}

func TestPasswordService_ValidatePasswordStrength(t *testing.T) {
	passwordService := NewPasswordService(nil)

	testCases := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid strong password",
			password: "TestPassword123!",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Test1!",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "no uppercase",
			password: "testpassword123!",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "no lowercase",
			password: "TESTPASSWORD123!",
			wantErr:  true,
			errMsg:   "lowercase letter",
		},
		{
			name:     "no number",
			password: "TestPassword!",
			wantErr:  true,
			errMsg:   "number",
		},
		{
			name:     "no special character",
			password: "TestPassword123",
			wantErr:  true,
			errMsg:   "special character",
		},
		{
			name:     "minimum valid password",
			password: "Test123!",
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := passwordService.ValidatePasswordStrength(tc.password)
			if tc.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPasswordService_HashVerificationWithDifferentVersions(t *testing.T) {
	passwordService := NewPasswordService(nil)
	password := "TestPassword123!"

	// Create hash with current service
	hash, err := passwordService.HashPassword(password)
	require.NoError(t, err)

	// Create new service instance (simulating restart)
	newPasswordService := NewPasswordService(nil)

	// Should still be able to verify with new instance
	valid, err := newPasswordService.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestPasswordService_EdgeCases(t *testing.T) {
	passwordService := NewPasswordService(nil)

	t.Run("empty password hash", func(t *testing.T) {
		hash, err := passwordService.HashPassword("")
		require.NoError(t, err)
		assert.NotEmpty(t, hash)

		valid, err := passwordService.VerifyPassword("", hash)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("very long password", func(t *testing.T) {
		longPassword := strings.Repeat("a", 1000) + "A1!"
		hash, err := passwordService.HashPassword(longPassword)
		require.NoError(t, err)

		valid, err := passwordService.VerifyPassword(longPassword, hash)
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("unicode password", func(t *testing.T) {
		unicodePassword := "Tëst🔒Pässwörd123!"
		hash, err := passwordService.HashPassword(unicodePassword)
		require.NoError(t, err)

		valid, err := passwordService.VerifyPassword(unicodePassword, hash)
		require.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestPasswordService_InvalidHashFormats(t *testing.T) {
	passwordService := NewPasswordService(nil)
	password := "TestPassword123!"

	invalidHashes := []string{
		"",
		"invalid",
		"$argon2id$",
		"$argon2id$v=19$",
		"$argon2id$v=19$m=65536$",
		"$argon2id$v=19$m=65536,t=3,p=2$",
		"$argon2id$v=18$m=65536,t=3,p=2$salt$hash",             // Wrong version
		"$argon2id$v=19$m=invalid,t=3,p=2$salt$hash",           // Invalid memory
		"$argon2id$v=19$m=65536,t=3,p=2$invalid-base64$hash",   // Invalid salt
		"$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$invalid-base64", // Invalid hash
	}

	for _, invalidHash := range invalidHashes {
		t.Run("invalid hash: "+invalidHash, func(t *testing.T) {
			_, err := passwordService.VerifyPassword(password, invalidHash)
			assert.Error(t, err)
		})
	}
}

func BenchmarkPasswordService_HashPassword(b *testing.B) {
	passwordService := NewPasswordService(nil)
	password := "TestPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := passwordService.HashPassword(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordService_VerifyPassword(b *testing.B) {
	passwordService := NewPasswordService(nil)
	password := "TestPassword123!"

	hash, err := passwordService.HashPassword(password)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := passwordService.VerifyPassword(password, hash)
		if err != nil {
			b.Fatal(err)
		}
	}
}
