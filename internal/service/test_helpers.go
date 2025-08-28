package service

import (
	"context"
)

// MockGeocodingProvider implements GeocodingProvider for testing
type MockGeocodingProvider struct {
	GeocodeFunc        func(ctx context.Context, address string) (*GeocodingResult, error)
	ReverseGeocodeFunc func(ctx context.Context, lat, lon float64) (*GeocodingResult, error)
	Name               string
}

func (m *MockGeocodingProvider) Geocode(ctx context.Context, address string) (*GeocodingResult, error) {
	if m.GeocodeFunc != nil {
		return m.GeocodeFunc(ctx, address)
	}
	return nil, ErrGeocodingFailed
}

func (m *MockGeocodingProvider) ReverseGeocode(ctx context.Context, lat, lon float64) (*GeocodingResult, error) {
	if m.ReverseGeocodeFunc != nil {
		return m.ReverseGeocodeFunc(ctx, lat, lon)
	}
	return nil, ErrGeocodingFailed
}

func (m *MockGeocodingProvider) GetName() string {
	if m.Name != "" {
		return m.Name
	}
	return "MockProvider"
}
