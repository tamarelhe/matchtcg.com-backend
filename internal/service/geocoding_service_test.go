package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/matchtcg/backend/internal/domain"
)

func TestGeocodingService_Geocode(t *testing.T) {
	tests := []struct {
		name           string
		address        string
		providerResult *GeocodingResult
		providerError  error
		expectedError  error
		expectCached   bool
	}{
		{
			name:    "successful geocoding",
			address: "123 Main St, Lisbon, Portugal",
			providerResult: &GeocodingResult{
				Coordinates: domain.Coordinates{Latitude: 38.7223, Longitude: -9.1393},
				Address: domain.Address{
					Street:    "123 Main St",
					City:      "Lisbon",
					Country:   "Portugal",
					Formatted: "123 Main St, Lisbon, Portugal",
				},
				Confidence: 95.0,
				Source:     "MockProvider",
			},
			expectCached: true,
		},
		{
			name:          "empty address",
			address:       "",
			expectedError: ErrInvalidAddress,
		},
		{
			name:          "whitespace only address",
			address:       "   ",
			expectedError: ErrInvalidAddress,
		},
		{
			name:          "provider error",
			address:       "123 Main St",
			providerError: ErrAddressNotFound,
			expectedError: ErrAddressNotFound,
		},
		{
			name:    "invalid coordinates from provider",
			address: "123 Main St",
			providerResult: &GeocodingResult{
				Coordinates: domain.Coordinates{Latitude: 91.0, Longitude: -9.1393}, // Invalid latitude
			},
			expectedError: domain.ErrInvalidCoordinates,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockGeocodingProvider{
				GeocodeFunc: func(ctx context.Context, address string) (*GeocodingResult, error) {
					if tt.providerError != nil {
						return nil, tt.providerError
					}
					return tt.providerResult, nil
				},
			}

			service := NewGeocodingService(mockProvider)

			result, err := service.Geocode(context.Background(), tt.address)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result, got nil")
				return
			}

			if tt.expectCached {
				// Test cache hit
				result2, err2 := service.Geocode(context.Background(), tt.address)
				if err2 != nil {
					t.Errorf("unexpected error on cached request: %v", err2)
				}
				if result2 == nil {
					t.Error("expected cached result, got nil")
				}
			}
		})
	}
}

func TestGeocodingService_ReverseGeocode(t *testing.T) {
	tests := []struct {
		name           string
		lat            float64
		lon            float64
		providerResult *GeocodingResult
		providerError  error
		expectedError  error
	}{
		{
			name: "successful reverse geocoding",
			lat:  38.7223,
			lon:  -9.1393,
			providerResult: &GeocodingResult{
				Coordinates: domain.Coordinates{Latitude: 38.7223, Longitude: -9.1393},
				Address: domain.Address{
					City:      "Lisbon",
					Country:   "Portugal",
					Formatted: "Lisbon, Portugal",
				},
				Confidence: 90.0,
				Source:     "MockProvider",
			},
		},
		{
			name:          "invalid latitude",
			lat:           91.0,
			lon:           -9.1393,
			expectedError: domain.ErrInvalidCoordinates,
		},
		{
			name:          "invalid longitude",
			lat:           38.7223,
			lon:           181.0,
			expectedError: domain.ErrInvalidCoordinates,
		},
		{
			name:          "provider error",
			lat:           38.7223,
			lon:           -9.1393,
			providerError: ErrAddressNotFound,
			expectedError: ErrAddressNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProvider := &MockGeocodingProvider{
				ReverseGeocodeFunc: func(ctx context.Context, lat, lon float64) (*GeocodingResult, error) {
					if tt.providerError != nil {
						return nil, tt.providerError
					}
					return tt.providerResult, nil
				},
			}

			service := NewGeocodingService(mockProvider)

			result, err := service.ReverseGeocode(context.Background(), tt.lat, tt.lon)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result, got nil")
			}
		})
	}
}

func TestGeocodingService_ValidateAndNormalizeCoordinates(t *testing.T) {
	service := NewGeocodingService(&MockGeocodingProvider{})

	tests := []struct {
		name        string
		lat         float64
		lon         float64
		expectedLat float64
		expectedLon float64
		expectError bool
	}{
		{
			name:        "valid coordinates",
			lat:         38.7223,
			lon:         -9.1393,
			expectedLat: 38.7223,
			expectedLon: -9.1393,
		},
		{
			name:        "normalize longitude > 180",
			lat:         38.7223,
			lon:         190.0,
			expectedLat: 38.7223,
			expectedLon: -170.0,
		},
		{
			name:        "normalize longitude < -180",
			lat:         38.7223,
			lon:         -190.0,
			expectedLat: 38.7223,
			expectedLon: 170.0,
		},
		{
			name:        "invalid latitude > 90",
			lat:         91.0,
			lon:         -9.1393,
			expectError: true,
		},
		{
			name:        "invalid latitude < -90",
			lat:         -91.0,
			lon:         -9.1393,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.ValidateAndNormalizeCoordinates(tt.lat, tt.lon)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Latitude != tt.expectedLat {
				t.Errorf("expected latitude %f, got %f", tt.expectedLat, result.Latitude)
			}

			if result.Longitude != tt.expectedLon {
				t.Errorf("expected longitude %f, got %f", tt.expectedLon, result.Longitude)
			}
		})
	}
}

func TestNominatimProvider_Geocode(t *testing.T) {
	// Create a test server that mimics Nominatim API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")

		if query == "rate_limit_test" {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		if query == "not_found" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "[]")
			return
		}

		if query == "server_error" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Return a mock successful response
		response := []nominatimResult{
			{
				PlaceID:     123456,
				Lat:         "38.7223",
				Lon:         "-9.1393",
				DisplayName: "Lisbon, Portugal",
				Importance:  0.95,
				Address: nominatimAddress{
					Road:        "Rua Augusta",
					City:        "Lisbon",
					Country:     "Portugal",
					CountryCode: "pt",
					Postcode:    "1100-048",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NominatimConfig{
		BaseURL:   server.URL,
		UserAgent: "TestAgent/1.0",
		Timeout:   5 * time.Second,
	}

	provider := NewNominatimProvider(config)

	tests := []struct {
		name          string
		address       string
		expectedError error
		expectResult  bool
	}{
		{
			name:         "successful geocoding",
			address:      "Lisbon, Portugal",
			expectResult: true,
		},
		{
			name:          "address not found",
			address:       "not_found",
			expectedError: ErrAddressNotFound,
		},
		{
			name:          "rate limit exceeded",
			address:       "rate_limit_test",
			expectedError: ErrRateLimitExceeded,
		},
		{
			name:          "server error",
			address:       "server_error",
			expectedError: ErrProviderUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := provider.Geocode(ctx, tt.address)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectResult && result == nil {
				t.Error("expected result, got nil")
			}

			if result != nil {
				if result.Source != "Nominatim" {
					t.Errorf("expected source 'Nominatim', got %s", result.Source)
				}
				if result.Coordinates.Latitude != 38.7223 {
					t.Errorf("expected latitude 38.7223, got %f", result.Coordinates.Latitude)
				}
				if result.Coordinates.Longitude != -9.1393 {
					t.Errorf("expected longitude -9.1393, got %f", result.Coordinates.Longitude)
				}
			}
		})
	}
}

func TestNominatimProvider_ReverseGeocode(t *testing.T) {
	// Create a test server that mimics Nominatim reverse API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lat := r.URL.Query().Get("lat")
		lon := r.URL.Query().Get("lon")

		if lat == "0" && lon == "0" {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		// Return a mock successful response
		response := nominatimResult{
			PlaceID:     123456,
			Lat:         lat,
			Lon:         lon,
			DisplayName: "Lisbon, Portugal",
			Importance:  0.90,
			Address: nominatimAddress{
				City:        "Lisbon",
				Country:     "Portugal",
				CountryCode: "pt",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := NominatimConfig{
		BaseURL:   server.URL,
		UserAgent: "TestAgent/1.0",
		Timeout:   5 * time.Second,
	}

	provider := NewNominatimProvider(config)

	tests := []struct {
		name          string
		lat           float64
		lon           float64
		expectedError error
		expectResult  bool
	}{
		{
			name:         "successful reverse geocoding",
			lat:          38.7223,
			lon:          -9.1393,
			expectResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := provider.ReverseGeocode(ctx, tt.lat, tt.lon)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError.Error()) {
					t.Errorf("expected error containing %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectResult && result == nil {
				t.Error("expected result, got nil")
			}

			if result != nil {
				if result.Source != "Nominatim" {
					t.Errorf("expected source 'Nominatim', got %s", result.Source)
				}
			}
		})
	}
}

func TestNominatimProvider_GetName(t *testing.T) {
	provider := NewNominatimProvider(NominatimConfig{})

	if provider.GetName() != "Nominatim" {
		t.Errorf("expected provider name 'Nominatim', got %s", provider.GetName())
	}
}

func TestGeocodingCache(t *testing.T) {
	cache := newGeocodingCache()

	result := &GeocodingResult{
		Coordinates: domain.Coordinates{Latitude: 38.7223, Longitude: -9.1393},
		Address: domain.Address{
			City:      "Lisbon",
			Country:   "Portugal",
			Formatted: "Lisbon, Portugal",
		},
		Confidence: 95.0,
		Source:     "Test",
	}

	// Test cache miss
	if cached := cache.get("test_key"); cached != nil {
		t.Error("expected cache miss, got result")
	}

	// Test cache set and hit
	cache.set("test_key", result)
	if cached := cache.get("test_key"); cached == nil {
		t.Error("expected cache hit, got nil")
	} else {
		if cached.Coordinates.Latitude != result.Coordinates.Latitude {
			t.Errorf("expected latitude %f, got %f", result.Coordinates.Latitude, cached.Coordinates.Latitude)
		}
	}
}

/* TODO
func TestRateLimiter(t *testing.T) {
	limiter := newRateLimiter(2, time.Second) // 2 requests per second

	ctx := context.Background()

	// First request should pass immediately
	start := time.Now()
	err := limiter.wait(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if time.Since(start) > 10*time.Millisecond {
		t.Error("first request should not be delayed")
	}

	// Second request should pass immediately
	start = time.Now()
	err = limiter.wait(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if time.Since(start) > 10*time.Millisecond {
		t.Error("second request should not be delayed")
	}

	// Third request should be delayed
	start = time.Now()
	err = limiter.wait(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 400*time.Millisecond { // Should wait ~500ms
		t.Errorf("third request should be delayed, but only waited %v", elapsed)
	}
}
*/

func TestRateLimiter_ContextCancellation(t *testing.T) {
	limiter := newRateLimiter(1, time.Second)

	// Make first request to trigger rate limiting
	err := limiter.wait(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after a short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	// Second request should be cancelled
	err = limiter.wait(ctx)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
