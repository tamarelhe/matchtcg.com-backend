package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matchtcg/backend/internal/domain"
)

var (
	ErrGeocodingFailed     = errors.New("geocoding failed")
	ErrAddressNotFound     = errors.New("address not found")
	ErrRateLimitExceeded   = errors.New("rate limit exceeded")
	ErrInvalidAddress      = errors.New("invalid address")
	ErrGeocodingTimeout    = errors.New("geocoding request timeout")
	ErrProviderUnavailable = errors.New("geocoding provider unavailable")
)

// GeocodingResult represents the result of a geocoding operation
type GeocodingResult struct {
	Coordinates domain.Coordinates `json:"coordinates"`
	Address     domain.Address     `json:"address"`
	Confidence  float64            `json:"confidence"`
	Source      string             `json:"source"`
}

// GeocodingProvider defines the interface for geocoding providers
type GeocodingProvider interface {
	Geocode(ctx context.Context, address string) (*GeocodingResult, error)
	ReverseGeocode(ctx context.Context, lat, lon float64) (*GeocodingResult, error)
	GetName() string
}

// GeocodingService provides geocoding functionality with provider abstraction
type GeocodingService struct {
	provider GeocodingProvider
	cache    *geocodingCache
}

// NewGeocodingService creates a new geocoding service with the specified provider
func NewGeocodingService(provider GeocodingProvider) *GeocodingService {
	return &GeocodingService{
		provider: provider,
		cache:    newGeocodingCache(),
	}
}

// Geocode converts an address to coordinates
func (s *GeocodingService) Geocode(ctx context.Context, address string) (*GeocodingResult, error) {
	if strings.TrimSpace(address) == "" {
		return nil, ErrInvalidAddress
	}

	// Check cache first
	if result := s.cache.get(address); result != nil {
		return result, nil
	}

	// Call provider
	result, err := s.provider.Geocode(ctx, address)
	if err != nil {
		return nil, err
	}

	// Validate result
	if err := result.Coordinates.Validate(); err != nil {
		return nil, fmt.Errorf("invalid coordinates from provider: %w", err)
	}

	// Cache the result
	s.cache.set(address, result)

	return result, nil
}

// ReverseGeocode converts coordinates to an address
func (s *GeocodingService) ReverseGeocode(ctx context.Context, lat, lon float64) (*GeocodingResult, error) {
	if !domain.IsValidCoordinates(lat, lon) {
		return nil, domain.ErrInvalidCoordinates
	}

	cacheKey := fmt.Sprintf("reverse_%f_%f", lat, lon)

	// Check cache first
	if result := s.cache.get(cacheKey); result != nil {
		return result, nil
	}

	// Call provider
	result, err := s.provider.ReverseGeocode(ctx, lat, lon)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cache.set(cacheKey, result)

	return result, nil
}

// ValidateAndNormalizeCoordinates validates and normalizes coordinate values
func (s *GeocodingService) ValidateAndNormalizeCoordinates(lat, lon float64) (domain.Coordinates, error) {
	// Check for NaN or Inf values first
	if math.IsNaN(lat) || math.IsNaN(lon) || math.IsInf(lat, 0) || math.IsInf(lon, 0) {
		return domain.Coordinates{}, domain.ErrInvalidCoordinates
	}

	// Check latitude range: -90 to 90 (cannot be normalized)
	if lat < -90 || lat > 90 {
		return domain.Coordinates{}, domain.ErrInvalidCoordinates
	}

	// Normalize longitude to -180 to 180 range
	normalizedLon := lon
	for normalizedLon > 180 {
		normalizedLon -= 360
	}
	for normalizedLon < -180 {
		normalizedLon += 360
	}

	return domain.Coordinates{
		Latitude:  lat,
		Longitude: normalizedLon,
	}, nil
}

// GetProviderName returns the name of the current geocoding provider
func (s *GeocodingService) GetProviderName() string {
	return s.provider.GetName()
}

// geocodingCache provides in-memory caching for geocoding results
type geocodingCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

type cacheEntry struct {
	result    *GeocodingResult
	timestamp time.Time
}

const (
	cacheExpiration = 24 * time.Hour
	maxCacheSize    = 1000
)

func newGeocodingCache() *geocodingCache {
	cache := &geocodingCache{
		entries: make(map[string]*cacheEntry),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *geocodingCache) get(key string) *GeocodingResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil
	}

	// Check if entry is expired
	if time.Since(entry.timestamp) > cacheExpiration {
		return nil
	}

	return entry.result
}

func (c *geocodingCache) set(key string, result *GeocodingResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If cache is full, remove oldest entries
	if len(c.entries) >= maxCacheSize {
		c.evictOldest()
	}

	c.entries[key] = &cacheEntry{
		result:    result,
		timestamp: time.Now(),
	}
}

func (c *geocodingCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

func (c *geocodingCache) cleanup() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if time.Since(entry.timestamp) > cacheExpiration {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// NominatimProvider implements the GeocodingProvider interface using OpenStreetMap's Nominatim service
type NominatimProvider struct {
	baseURL     string
	userAgent   string
	httpClient  *http.Client
	rateLimiter *rateLimiter
}

// NominatimConfig holds configuration for the Nominatim provider
type NominatimConfig struct {
	BaseURL   string
	UserAgent string
	Timeout   time.Duration
}

// NewNominatimProvider creates a new Nominatim geocoding provider
func NewNominatimProvider(config NominatimConfig) *NominatimProvider {
	if config.BaseURL == "" {
		config.BaseURL = "https://nominatim.openstreetmap.org"
	}
	if config.UserAgent == "" {
		config.UserAgent = "MatchTCG/1.0"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	return &NominatimProvider{
		baseURL:   config.BaseURL,
		userAgent: config.UserAgent,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		rateLimiter: newRateLimiter(1, time.Second), // 1 request per second
	}
}

// GetName returns the provider name
func (p *NominatimProvider) GetName() string {
	return "Nominatim"
}

// Geocode converts an address to coordinates using Nominatim
func (p *NominatimProvider) Geocode(ctx context.Context, address string) (*GeocodingResult, error) {
	if err := p.rateLimiter.wait(ctx); err != nil {
		return nil, ErrRateLimitExceeded
	}

	params := url.Values{}
	params.Set("q", address)
	params.Set("format", "json")
	params.Set("limit", "1")
	params.Set("addressdetails", "1")

	reqURL := fmt.Sprintf("%s/search?%s", p.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrGeocodingTimeout
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, ErrRateLimitExceeded
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, ErrProviderUnavailable)
	}

	var results []nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrAddressNotFound
	}

	result := results[0]
	lat, err := strconv.ParseFloat(result.Lat, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %w", err)
	}

	lon, err := strconv.ParseFloat(result.Lon, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %w", err)
	}

	return &GeocodingResult{
		Coordinates: domain.Coordinates{
			Latitude:  lat,
			Longitude: lon,
		},
		Address: domain.Address{
			Street:     result.Address.Road,
			City:       result.Address.City,
			State:      result.Address.State,
			PostalCode: result.Address.Postcode,
			Country:    result.Address.Country,
			Formatted:  result.DisplayName,
		},
		Confidence: parseImportance(result.Importance),
		Source:     "Nominatim",
	}, nil
}

// ReverseGeocode converts coordinates to an address using Nominatim
func (p *NominatimProvider) ReverseGeocode(ctx context.Context, lat, lon float64) (*GeocodingResult, error) {
	if err := p.rateLimiter.wait(ctx); err != nil {
		return nil, ErrRateLimitExceeded
	}

	params := url.Values{}
	params.Set("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	params.Set("lon", strconv.FormatFloat(lon, 'f', 6, 64))
	params.Set("format", "json")
	params.Set("addressdetails", "1")

	reqURL := fmt.Sprintf("%s/reverse?%s", p.baseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", p.userAgent)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrGeocodingTimeout
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, ErrRateLimitExceeded
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, ErrProviderUnavailable)
	}

	var result nominatimResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &GeocodingResult{
		Coordinates: domain.Coordinates{
			Latitude:  lat,
			Longitude: lon,
		},
		Address: domain.Address{
			Street:     result.Address.Road,
			City:       result.Address.City,
			State:      result.Address.State,
			PostalCode: result.Address.Postcode,
			Country:    result.Address.Country,
			Formatted:  result.DisplayName,
		},
		Confidence: parseImportance(result.Importance),
		Source:     "Nominatim",
	}, nil
}

// nominatimResult represents the response structure from Nominatim API
type nominatimResult struct {
	PlaceID     int                    `json:"place_id"`
	Licence     string                 `json:"licence"`
	OSMType     string                 `json:"osm_type"`
	OSMId       int                    `json:"osm_id"`
	Lat         string                 `json:"lat"`
	Lon         string                 `json:"lon"`
	DisplayName string                 `json:"display_name"`
	Class       string                 `json:"class"`
	Type        string                 `json:"type"`
	Importance  float64                `json:"importance"`
	Address     nominatimAddress       `json:"address"`
	BoundingBox []string               `json:"boundingbox"`
	ExtraData   map[string]interface{} `json:"extratags"`
}

type nominatimAddress struct {
	Road        string `json:"road"`
	House       string `json:"house_number"`
	City        string `json:"city"`
	Town        string `json:"town"`
	Village     string `json:"village"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	Country     string `json:"country"`
	CountryCode string `json:"country_code"`
}

func parseImportance(importance float64) float64 {
	// Nominatim importance is typically 0-1, convert to 0-100 scale
	if importance < 0 {
		return 0
	}
	if importance > 1 {
		return 100
	}
	return importance * 100
}

// rateLimiter implements a simple rate limiter
type rateLimiter struct {
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

func newRateLimiter(requests int, duration time.Duration) *rateLimiter {
	return &rateLimiter{
		interval: duration / time.Duration(requests),
	}
}

func (rl *rateLimiter) wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if elapsed := now.Sub(rl.lastCall); elapsed < rl.interval {
		waitTime := rl.interval - elapsed

		select {
		case <-time.After(waitTime):
			rl.lastCall = time.Now()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	rl.lastCall = now
	return nil
}
