package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for validation
type TestRegisterRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name,omitempty" validate:"max=100"`
	Locale      string `json:"locale,omitempty" validate:"locale"`
	Timezone    string `json:"timezone" validate:"required,timezone"`
}

type TestEventRequest struct {
	Title      string   `json:"title" validate:"required,min=1,max=200"`
	Game       string   `json:"game" validate:"required,game_type"`
	Visibility string   `json:"visibility" validate:"required,event_visibility"`
	Tags       []string `json:"tags,omitempty"`
}

type TestVenueRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=200"`
	Type    string `json:"type" validate:"required,venue_type"`
	Address string `json:"address" validate:"required,min=1,max=500"`
}

func TestNewValidationMiddleware(t *testing.T) {
	vm := NewValidationMiddleware()
	assert.NotNil(t, vm)
	assert.NotNil(t, vm.validator)
}

func TestValidateStruct_ValidInput(t *testing.T) {
	vm := NewValidationMiddleware()

	req := TestRegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Locale:   "en",
		Timezone: "Europe/Lisbon",
	}

	errors := vm.ValidateStruct(req)
	assert.Empty(t, errors)
}

func TestValidateStruct_InvalidInput(t *testing.T) {
	vm := NewValidationMiddleware()

	req := TestRegisterRequest{
		Email:    "invalid-email",
		Password: "short",
		Locale:   "invalid",
		Timezone: "invalid",
	}

	errors := vm.ValidateStruct(req)
	assert.NotEmpty(t, errors)
	assert.Len(t, errors, 4) // email, password, locale, timezone

	// Check specific error messages
	errorFields := make(map[string]string)
	for _, err := range errors {
		errorFields[err.Field] = err.Message
	}

	assert.Contains(t, errorFields["email"], "valid email")
	assert.Contains(t, errorFields["password"], "at least 8")
	assert.Contains(t, errorFields["locale"], "en, pt")
	assert.Contains(t, errorFields["timezone"], "valid timezone")
}

func TestValidateStruct_MissingRequiredFields(t *testing.T) {
	vm := NewValidationMiddleware()

	req := TestRegisterRequest{
		// Missing required fields
	}

	errors := vm.ValidateStruct(req)
	assert.NotEmpty(t, errors)

	// Check that required fields are reported
	errorFields := make(map[string]bool)
	for _, err := range errors {
		errorFields[err.Field] = true
	}

	assert.True(t, errorFields["email"])
	assert.True(t, errorFields["password"])
	assert.True(t, errorFields["timezone"])
}

func TestCustomValidators_GameType(t *testing.T) {
	vm := NewValidationMiddleware()

	tests := []struct {
		name     string
		game     string
		expected bool
	}{
		{"valid mtg", "mtg", true},
		{"valid lorcana", "lorcana", true},
		{"valid pokemon", "pokemon", true},
		{"valid other", "other", true},
		{"invalid game", "invalid", false},
		{"empty game", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := TestEventRequest{
				Title:      "Test Event",
				Game:       tt.game,
				Visibility: "public",
			}

			errors := vm.ValidateStruct(req)
			hasGameError := false
			for _, err := range errors {
				if err.Field == "game" {
					hasGameError = true
					break
				}
			}

			if tt.expected {
				assert.False(t, hasGameError, "Expected no game validation error")
			} else {
				assert.True(t, hasGameError, "Expected game validation error")
			}
		})
	}
}

func TestCustomValidators_EventVisibility(t *testing.T) {
	vm := NewValidationMiddleware()

	tests := []struct {
		name       string
		visibility string
		expected   bool
	}{
		{"valid public", "public", true},
		{"valid private", "private", true},
		{"valid group", "group", true},
		{"invalid visibility", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := TestEventRequest{
				Title:      "Test Event",
				Game:       "mtg",
				Visibility: tt.visibility,
			}

			errors := vm.ValidateStruct(req)
			hasVisibilityError := false
			for _, err := range errors {
				if err.Field == "visibility" {
					hasVisibilityError = true
					break
				}
			}

			if tt.expected {
				assert.False(t, hasVisibilityError, "Expected no visibility validation error")
			} else {
				assert.True(t, hasVisibilityError, "Expected visibility validation error")
			}
		})
	}
}

func TestCustomValidators_VenueType(t *testing.T) {
	vm := NewValidationMiddleware()

	tests := []struct {
		name      string
		venueType string
		expected  bool
	}{
		{"valid store", "store", true},
		{"valid home", "home", true},
		{"valid other", "other", true},
		{"invalid type", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := TestVenueRequest{
				Name:    "Test Venue",
				Type:    tt.venueType,
				Address: "123 Test St",
			}

			errors := vm.ValidateStruct(req)
			hasTypeError := false
			for _, err := range errors {
				if err.Field == "type" {
					hasTypeError = true
					break
				}
			}

			if tt.expected {
				assert.False(t, hasTypeError, "Expected no type validation error")
			} else {
				assert.True(t, hasTypeError, "Expected type validation error")
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	vm := NewValidationMiddleware()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic HTML escape",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "javascript protocol",
			input:    "javascript:alert('xss')",
			expected: "alert(&#39;xss&#39;)",
		},
		{
			name:     "event handlers",
			input:    "onload=alert('xss')",
			expected: "alert(&#39;xss&#39;)",
		},
		{
			name:     "normal text",
			input:    "This is normal text",
			expected: "This is normal text",
		},
		{
			name:     "special characters",
			input:    "Test & Co. <company>",
			expected: "Test &amp; Co. &lt;company&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeStruct(t *testing.T) {
	vm := NewValidationMiddleware()

	type TestStruct struct {
		Name        string   `json:"name"`
		Description *string  `json:"description"`
		Tags        []string `json:"tags"`
	}

	desc := "<script>alert('xss')</script>"
	req := TestStruct{
		Name:        "<script>alert('name')</script>",
		Description: &desc,
		Tags:        []string{"<script>alert('tag')</script>", "normal tag"},
	}

	vm.SanitizeStruct(&req)

	assert.Contains(t, req.Name, "&lt;script&gt;")
	assert.NotContains(t, req.Name, "<script>")

	assert.Contains(t, *req.Description, "&lt;script&gt;")
	assert.NotContains(t, *req.Description, "<script>")

	assert.Contains(t, req.Tags[0], "&lt;script&gt;")
	assert.NotContains(t, req.Tags[0], "<script>")
	assert.Equal(t, "normal tag", req.Tags[1])
}

func TestValidateAndSanitize(t *testing.T) {
	vm := NewValidationMiddleware()

	req := TestRegisterRequest{
		Email:    "test@example.com",
		Password: "password123<script>",
		Locale:   "en", // Add valid locale
		Timezone: "Europe/Lisbon",
	}

	errors := vm.ValidateAndSanitize(&req)
	assert.Empty(t, errors)

	// Check that sanitization occurred
	assert.Contains(t, req.Password, "&lt;script&gt;")
	assert.NotContains(t, req.Password, "<script>")
}

func TestWriteValidationError(t *testing.T) {
	vm := NewValidationMiddleware()

	errors := []ValidationError{
		{
			Field:   "email",
			Message: "email is required",
		},
		{
			Field:   "password",
			Message: "password must be at least 8 characters long",
		},
	}

	w := httptest.NewRecorder()
	vm.WriteValidationError(w, errors)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ValidationErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "validation_error", response.Error)
	assert.Equal(t, "Request validation failed", response.Message)
	assert.Len(t, response.Details, 2)
	assert.Equal(t, "email", response.Details[0].Field)
	assert.Equal(t, "password", response.Details[1].Field)
}

func TestValidateJSON_Middleware(t *testing.T) {
	vm := NewValidationMiddleware()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with validation middleware
	middleware := vm.ValidateJSON(TestRegisterRequest{})
	wrappedHandler := middleware(handler)

	t.Run("valid JSON", func(t *testing.T) {
		req := TestRegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Locale:   "en", // Add valid locale
			Timezone: "Europe/Lisbon",
		}

		body, _ := json.Marshal(req)
		r := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})

	t.Run("invalid JSON", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("invalid json")))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "validation_error", response.Error)
	})

	t.Run("validation errors", func(t *testing.T) {
		req := TestRegisterRequest{
			Email:    "invalid-email",
			Password: "short",
		}

		body, _ := json.Marshal(req)
		r := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "validation_error", response.Error)
		assert.NotEmpty(t, response.Details)
	})

	t.Run("non-JSON request", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		wrappedHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})
}

func TestGetErrorMessage(t *testing.T) {
	vm := NewValidationMiddleware()

	// Test with a struct that will generate validation errors
	type TestStruct struct {
		Email string `validate:"required,email"`
		Name  string `validate:"min=2,max=10"`
		Game  string `validate:"game_type"`
	}

	req := TestStruct{
		Email: "invalid",
		Name:  "a", // too short
		Game:  "invalid",
	}

	errors := vm.ValidateStruct(req)
	assert.NotEmpty(t, errors)

	// Check that error messages are human-readable
	for _, err := range errors {
		assert.NotEmpty(t, err.Message)
		assert.Contains(t, err.Message, err.Field)
	}
}

func TestEdgeCases(t *testing.T) {
	vm := NewValidationMiddleware()

	t.Run("empty struct", func(t *testing.T) {
		type EmptyStruct struct{}
		req := EmptyStruct{}
		errors := vm.ValidateStruct(req)
		assert.Empty(t, errors)
	})

	t.Run("nil pointer", func(t *testing.T) {
		var req *TestRegisterRequest
		errors := vm.ValidateStruct(req)
		assert.NotEmpty(t, errors) // Should handle nil gracefully
	})

	t.Run("struct with no validation tags", func(t *testing.T) {
		type NoValidationStruct struct {
			Name string
			Age  int
		}
		req := NoValidationStruct{Name: "test", Age: 25}
		errors := vm.ValidateStruct(req)
		assert.Empty(t, errors)
	})
}
