package middleware

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/matchtcg/backend/internal/constant"
)

// ValidationMiddleware provides request validation functionality
type ValidationMiddleware struct {
	validator *validator.Validate
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse represents the validation error response
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details"`
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	v := validator.New()

	// Register custom tag name function to use JSON field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators(v)

	return &ValidationMiddleware{
		validator: v,
	}
}

// registerCustomValidators registers custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Custom validator for game types
	v.RegisterValidation("game_type", func(fl validator.FieldLevel) bool {
		gameType := fl.Field().String()
		validGames := []string{"mtg", "lorcana", "pokemon", "other"}
		for _, valid := range validGames {
			if gameType == valid {
				return true
			}
		}
		return false
	})

	// Custom validator for event visibility
	v.RegisterValidation("event_visibility", func(fl validator.FieldLevel) bool {
		visibility := fl.Field().String()
		validVisibilities := []string{"public", "private", "group"}
		for _, valid := range validVisibilities {
			if visibility == valid {
				return true
			}
		}
		return false
	})

	// Custom validator for RSVP status
	v.RegisterValidation("rsvp_status", func(fl validator.FieldLevel) bool {
		status := fl.Field().String()
		validStatuses := []string{"going", "interested", "declined"}
		for _, valid := range validStatuses {
			if status == valid {
				return true
			}
		}
		return false
	})

	// Custom validator for venue type
	v.RegisterValidation("venue_type", func(fl validator.FieldLevel) bool {
		venueType := fl.Field().String()
		validTypes := []string{"store", "home", "other"}
		for _, valid := range validTypes {
			if venueType == valid {
				return true
			}
		}
		return false
	})

	// Custom validator for group role
	v.RegisterValidation("group_role", func(fl validator.FieldLevel) bool {
		role := fl.Field().String()
		validRoles := []string{"member", "admin", "owner"}
		for _, valid := range validRoles {
			if role == valid {
				return true
			}
		}
		return false
	})

	// Custom validator for locale
	v.RegisterValidation("locale", func(fl validator.FieldLevel) bool {
		locale := fl.Field().String()
		// Allow empty values (will be handled by omitempty tag)
		if locale == "" {
			return true
		}
		validLocales := []string{"en", "pt"}
		for _, valid := range validLocales {
			if locale == valid {
				return true
			}
		}
		return false
	})

	// Custom validator for country
	v.RegisterValidation("country", func(fl validator.FieldLevel) bool {
		country := fl.Field().String()
		// Allow empty values (will be handled by omitempty tag)
		if country == "" {
			return true
		}

		// Validate country code
		_, ok := constant.ISO3166Alpha2[country]
		if !ok {
			return false
		}

		return true
	})

	// Custom validator for timezone (basic validation)
	v.RegisterValidation("timezone", func(fl validator.FieldLevel) bool {
		timezone := fl.Field().String()
		// Allow empty values (will be handled by omitempty tag)
		if timezone == "" {
			return true
		}
		// Basic timezone validation - should contain a slash for region/city format
		return strings.Contains(timezone, "/") && len(timezone) > 3
	})

	// Custom validator for coordinates format "lat,lon"
	v.RegisterValidation("coordinates", func(fl validator.FieldLevel) bool {
		coords := fl.Field().String()
		parts := strings.Split(coords, ",")
		if len(parts) != 2 {
			return false
		}
		// Additional validation could parse floats here
		return len(strings.TrimSpace(parts[0])) > 0 && len(strings.TrimSpace(parts[1])) > 0
	})
}

// ValidateStruct validates a struct and returns validation errors
func (vm *ValidationMiddleware) ValidateStruct(s interface{}) []ValidationError {
	var validationErrors []ValidationError

	err := vm.validator.Struct(s)
	if err != nil {
		// Handle different types of validation errors
		if valErrors, ok := err.(validator.ValidationErrors); ok {
			for _, valErr := range valErrors {
				validationErrors = append(validationErrors, ValidationError{
					Field:   valErr.Field(),
					Message: vm.getErrorMessage(valErr),
					Value:   fmt.Sprintf("%v", valErr.Value()),
				})
			}
		} else {
			// Handle InvalidValidationError (e.g., nil pointer)
			validationErrors = append(validationErrors, ValidationError{
				Field:   "struct",
				Message: "Invalid validation input",
				Value:   "",
			})
		}
	}

	return validationErrors
}

// getErrorMessage returns a human-readable error message for validation errors
func (vm *ValidationMiddleware) getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s cannot exceed %s characters", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", fe.Field())
	case "datetime":
		return fmt.Sprintf("%s must be a valid date-time", fe.Field())
	case "game_type":
		return fmt.Sprintf("%s must be one of: mtg, lorcana, pokemon, other", fe.Field())
	case "event_visibility":
		return fmt.Sprintf("%s must be one of: public, private, group", fe.Field())
	case "rsvp_status":
		return fmt.Sprintf("%s must be one of: going, interested, declined", fe.Field())
	case "venue_type":
		return fmt.Sprintf("%s must be one of: store, home, other", fe.Field())
	case "group_role":
		return fmt.Sprintf("%s must be one of: member, admin, owner", fe.Field())
	case "locale":
		return fmt.Sprintf("%s must be one of: en, pt", fe.Field())
	case "timezone":
		return fmt.Sprintf("%s must be a valid timezone (e.g., Europe/Lisbon)", fe.Field())
	case "coordinates":
		return fmt.Sprintf("%s must be in format 'latitude,longitude'", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}

// SanitizeInput sanitizes input to prevent XSS attacks
func (vm *ValidationMiddleware) SanitizeInput(input string) string {
	// HTML escape to prevent XSS
	sanitized := html.EscapeString(input)

	// Remove potentially dangerous characters/sequences
	sanitized = strings.ReplaceAll(sanitized, "<script", "&lt;script")
	sanitized = strings.ReplaceAll(sanitized, "</script>", "&lt;/script&gt;")
	sanitized = strings.ReplaceAll(sanitized, "javascript:", "")
	sanitized = strings.ReplaceAll(sanitized, "vbscript:", "")
	sanitized = strings.ReplaceAll(sanitized, "onload=", "")
	sanitized = strings.ReplaceAll(sanitized, "onerror=", "")

	return sanitized
}

// SanitizeStruct recursively sanitizes string fields in a struct
func (vm *ValidationMiddleware) SanitizeStruct(s interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			if field.String() != "" {
				field.SetString(vm.SanitizeInput(field.String()))
			}
		case reflect.Ptr:
			if !field.IsNil() && field.Elem().Kind() == reflect.String {
				str := field.Elem().String()
				if str != "" {
					field.Elem().SetString(vm.SanitizeInput(str))
				}
			}
		case reflect.Slice:
			if field.Type().Elem().Kind() == reflect.String {
				for j := 0; j < field.Len(); j++ {
					elem := field.Index(j)
					if elem.String() != "" {
						elem.SetString(vm.SanitizeInput(elem.String()))
					}
				}
			}
		case reflect.Struct:
			// Recursively sanitize nested structs
			if field.CanAddr() {
				vm.SanitizeStruct(field.Addr().Interface())
			}
		}

		// Handle special cases based on field tags
		jsonTag := fieldType.Tag.Get("json")
		if strings.Contains(jsonTag, "omitempty") && field.IsZero() {
			continue
		}
	}
}

// ValidateAndSanitize validates and sanitizes a struct
func (vm *ValidationMiddleware) ValidateAndSanitize(s interface{}) []ValidationError {
	// First sanitize the input
	vm.SanitizeStruct(s)

	// Then validate
	return vm.ValidateStruct(s)
}

// WriteValidationError writes a standardized validation error response
func (vm *ValidationMiddleware) WriteValidationError(w http.ResponseWriter, errors []ValidationError) {
	response := ValidationErrorResponse{
		Error:   "validation_error",
		Message: "Request validation failed",
		Details: errors,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(response)
}

// ValidateJSON is a middleware that validates JSON request bodies
func (vm *ValidationMiddleware) ValidateJSON(requestType interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate for requests with JSON content
			contentType := r.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				next.ServeHTTP(w, r)
				return
			}

			// Create a new instance of the request type
			reqValue := reflect.New(reflect.TypeOf(requestType))
			req := reqValue.Interface()

			// Decode JSON
			if err := json.NewDecoder(r.Body).Decode(req); err != nil {
				vm.WriteValidationError(w, []ValidationError{
					{
						Field:   "body",
						Message: "Invalid JSON format",
					},
				})
				return
			}

			// Validate and sanitize
			if errors := vm.ValidateAndSanitize(req); len(errors) > 0 {
				vm.WriteValidationError(w, errors)
				return
			}

			// Store validated request in context for handlers to use
			// This would require context modification, for now just continue
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateQueryParams validates query parameters against a struct
func (vm *ValidationMiddleware) ValidateQueryParams(r *http.Request, s interface{}) []ValidationError {
	// This would populate the struct from query parameters and validate
	// For now, return empty slice - full implementation would require
	// query parameter parsing and struct field mapping
	return []ValidationError{}
}
