package handler

import (
	"encoding/json"
	"net/http"

	"github.com/matchtcg/backend/internal/middleware"
)

// ValidationHelper provides validation utilities for handlers
type ValidationHelper struct {
	validator *middleware.ValidationMiddleware
}

// NewValidationHelper creates a new validation helper
func NewValidationHelper() *ValidationHelper {
	return &ValidationHelper{
		validator: middleware.NewValidationMiddleware(),
	}
}

// ValidateAndDecodeJSON validates and decodes JSON request body
func (vh *ValidationHelper) ValidateAndDecodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		vh.writeValidationError(w, []middleware.ValidationError{
			{
				Field:   "body",
				Message: "Invalid JSON format",
			},
		})
		return false
	}

	// Validate and sanitize
	if errors := vh.validator.ValidateAndSanitize(dst); len(errors) > 0 {
		vh.writeValidationError(w, errors)
		return false
	}

	return true
}

// ValidateStruct validates a struct and writes errors if validation fails
func (vh *ValidationHelper) ValidateStruct(w http.ResponseWriter, s interface{}) bool {
	if errors := vh.validator.ValidateAndSanitize(s); len(errors) > 0 {
		vh.writeValidationError(w, errors)
		return false
	}
	return true
}

// writeValidationError writes validation errors in the standard format
func (vh *ValidationHelper) writeValidationError(w http.ResponseWriter, errors []middleware.ValidationError) {
	response := middleware.ValidationErrorResponse{
		Error:   "validation_error",
		Message: "Request validation failed",
		Details: errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

// SanitizeInput sanitizes a string input
func (vh *ValidationHelper) SanitizeInput(input string) string {
	return vh.validator.SanitizeInput(input)
}
