package handler

// Common response types shared across handlers

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"display_name"`
}

// AuthUserInfo represents user information in auth responses (includes email)
type AuthUserInfo struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	Locale      string  `json:"locale"`
}

// Coordinates represents geographic coordinates
type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GroupInfo represents basic group information
type GroupInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
