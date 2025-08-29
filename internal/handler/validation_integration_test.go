package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matchtcg/backend/internal/middleware"
)

func TestValidationHelper_ValidateAndDecodeJSON(t *testing.T) {
	vh := NewValidationHelper()

	t.Run("valid JSON and validation", func(t *testing.T) {
		req := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Timezone: "Europe/Lisbon",
		}

		body, _ := json.Marshal(req)
		r := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		var decoded RegisterRequest
		result := vh.ValidateAndDecodeJSON(w, r, &decoded)

		assert.True(t, result)
		assert.Equal(t, req.Email, decoded.Email)
		assert.Equal(t, req.Password, decoded.Password)
		assert.Equal(t, req.Timezone, decoded.Timezone)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		r := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()

		var decoded RegisterRequest
		result := vh.ValidateAndDecodeJSON(w, r, &decoded)

		assert.False(t, result)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "validation_error", response.Error)
	})

	t.Run("validation errors", func(t *testing.T) {
		req := RegisterRequest{
			Email:    "invalid-email",
			Password: "short",
			Timezone: "invalid",
		}

		body, _ := json.Marshal(req)
		r := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		var decoded RegisterRequest
		result := vh.ValidateAndDecodeJSON(w, r, &decoded)

		assert.False(t, result)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "validation_error", response.Error)
		assert.NotEmpty(t, response.Details)

		// Check that we have validation errors for the expected fields
		errorFields := make(map[string]bool)
		for _, detail := range response.Details {
			errorFields[detail.Field] = true
		}
		assert.True(t, errorFields["email"])
		assert.True(t, errorFields["password"])
		assert.True(t, errorFields["timezone"])
	})

	t.Run("XSS sanitization", func(t *testing.T) {
		req := RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "<script>alert('xss')</script>",
			Timezone:    "Europe/Lisbon",
		}

		body, _ := json.Marshal(req)
		r := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
		w := httptest.NewRecorder()

		var decoded RegisterRequest
		result := vh.ValidateAndDecodeJSON(w, r, &decoded)

		assert.True(t, result)
		assert.Contains(t, decoded.DisplayName, "&lt;script&gt;")
		assert.NotContains(t, decoded.DisplayName, "<script>")
	})
}

func TestValidationHelper_ValidateStruct(t *testing.T) {
	vh := NewValidationHelper()

	t.Run("valid struct", func(t *testing.T) {
		req := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Timezone: "Europe/Lisbon",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.True(t, result)
		assert.Equal(t, http.StatusOK, w.Code) // No response written
	})

	t.Run("invalid struct", func(t *testing.T) {
		req := RegisterRequest{
			Email:    "invalid-email",
			Password: "short",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.False(t, result)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "validation_error", response.Error)
		assert.NotEmpty(t, response.Details)
	})
}

func TestValidationHelper_SanitizeInput(t *testing.T) {
	vh := NewValidationHelper()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "XSS script tag",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "normal text",
			input:    "This is normal text",
			expected: "This is normal text",
		},
		{
			name:     "HTML entities",
			input:    "Test & Co.",
			expected: "Test &amp; Co.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vh.SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationTags_EventRequest(t *testing.T) {
	vh := NewValidationHelper()

	t.Run("valid event request", func(t *testing.T) {
		req := CreateEventRequest{
			Title:      "Test Event",
			Game:       "mtg",
			Visibility: "public",
			Timezone:   "Europe/Lisbon",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		// This will fail because StartAt and EndAt are required but zero values
		// But we can check that game and visibility validation passes
		assert.False(t, result) // Due to missing required fields

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that game and visibility are not in the error list
		errorFields := make(map[string]bool)
		for _, detail := range response.Details {
			errorFields[detail.Field] = true
		}
		assert.False(t, errorFields["game"])       // Should be valid
		assert.False(t, errorFields["visibility"]) // Should be valid
	})

	t.Run("invalid game type", func(t *testing.T) {
		req := CreateEventRequest{
			Title:      "Test Event",
			Game:       "invalid_game",
			Visibility: "public",
			Timezone:   "Europe/Lisbon",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.False(t, result)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that game validation failed
		errorFields := make(map[string]string)
		for _, detail := range response.Details {
			errorFields[detail.Field] = detail.Message
		}
		assert.Contains(t, errorFields["game"], "mtg, lorcana, pokemon, other")
	})

	t.Run("invalid visibility", func(t *testing.T) {
		req := CreateEventRequest{
			Title:      "Test Event",
			Game:       "mtg",
			Visibility: "invalid_visibility",
			Timezone:   "Europe/Lisbon",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.False(t, result)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that visibility validation failed
		errorFields := make(map[string]string)
		for _, detail := range response.Details {
			errorFields[detail.Field] = detail.Message
		}
		assert.Contains(t, errorFields["visibility"], "public, private, group")
	})
}

func TestValidationTags_VenueRequest(t *testing.T) {
	vh := NewValidationHelper()

	t.Run("valid venue request", func(t *testing.T) {
		req := CreateVenueRequest{
			Name:    "Test Venue",
			Type:    "store",
			Address: "123 Test St",
			City:    "Test City",
			Country: "Test Country",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.True(t, result)
	})

	t.Run("invalid venue type", func(t *testing.T) {
		req := CreateVenueRequest{
			Name:    "Test Venue",
			Type:    "invalid_type",
			Address: "123 Test St",
			City:    "Test City",
			Country: "Test Country",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.False(t, result)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that type validation failed
		errorFields := make(map[string]string)
		for _, detail := range response.Details {
			errorFields[detail.Field] = detail.Message
		}
		assert.Contains(t, errorFields["type"], "store, home, other")
	})
}

func TestValidationTags_GroupRequest(t *testing.T) {
	vh := NewValidationHelper()

	t.Run("valid add member request", func(t *testing.T) {
		req := AddMemberRequest{
			UserID: "550e8400-e29b-41d4-a716-446655440000",
			Role:   "member",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.True(t, result)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		req := AddMemberRequest{
			UserID: "invalid-uuid",
			Role:   "member",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.False(t, result)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that UUID validation failed
		errorFields := make(map[string]string)
		for _, detail := range response.Details {
			errorFields[detail.Field] = detail.Message
		}
		assert.Contains(t, errorFields["user_id"], "valid UUID")
	})

	t.Run("invalid role", func(t *testing.T) {
		req := AddMemberRequest{
			UserID: "550e8400-e29b-41d4-a716-446655440000",
			Role:   "invalid_role",
		}

		w := httptest.NewRecorder()
		result := vh.ValidateStruct(w, &req)

		assert.False(t, result)

		var response middleware.ValidationErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Check that role validation failed
		errorFields := make(map[string]string)
		for _, detail := range response.Details {
			errorFields[detail.Field] = detail.Message
		}
		assert.Contains(t, errorFields["role"], "member, admin, owner")
	})
}
