package domain

import (
	"errors"
	"math"
)

var (
	ErrInvalidRadius  = errors.New("radius must be greater than 0")
	ErrRadiusTooLarge = errors.New("radius cannot exceed 1000 km")
)

const (
	// EarthRadiusKm is the Earth's radius in kilometers
	EarthRadiusKm = 6371.0

	// MaxSearchRadiusKm is the maximum allowed search radius
	MaxSearchRadiusKm = 1000

	// MinSearchRadiusKm is the minimum allowed search radius
	MinSearchRadiusKm = 1
)

// GeospatialService provides geospatial calculations and utilities
type GeospatialService struct{}

// NewGeospatialService creates a new GeospatialService
func NewGeospatialService() *GeospatialService {
	return &GeospatialService{}
}

// CalculateDistance calculates the distance between two coordinates using the Haversine formula
func (s *GeospatialService) CalculateDistance(coord1, coord2 Coordinates) float64 {
	return coord1.DistanceTo(coord2)
}

// IsWithinRadius checks if two coordinates are within the specified radius (in kilometers)
func (s *GeospatialService) IsWithinRadius(center, point Coordinates, radiusKm float64) bool {
	if radiusKm <= 0 {
		return false
	}

	distance := s.CalculateDistance(center, point)
	return distance <= radiusKm
}

// ValidateSearchRadius validates that a search radius is within acceptable limits
func (s *GeospatialService) ValidateSearchRadius(radiusKm int) error {
	if radiusKm <= 0 {
		return ErrInvalidRadius
	}

	if radiusKm > MaxSearchRadiusKm {
		return ErrRadiusTooLarge
	}

	return nil
}

// CalculateBoundingBox calculates the bounding box for a given center point and radius
func (s *GeospatialService) CalculateBoundingBox(center Coordinates, radiusKm float64) *BoundingBox {
	// Convert radius from kilometers to degrees (approximate)
	latDelta := radiusKm / 111.0 // 1 degree latitude ≈ 111 km
	lonDelta := radiusKm / (111.0 * math.Cos(center.Latitude*math.Pi/180))

	return &BoundingBox{
		NorthEast: Coordinates{
			Latitude:  center.Latitude + latDelta,
			Longitude: center.Longitude + lonDelta,
		},
		SouthWest: Coordinates{
			Latitude:  center.Latitude - latDelta,
			Longitude: center.Longitude - lonDelta,
		},
	}
}

// GetNearbyPoints filters a list of coordinates to only include those within the specified radius
func (s *GeospatialService) GetNearbyPoints(center Coordinates, points []Coordinates, radiusKm float64) []CoordinateWithDistance {
	var nearby []CoordinateWithDistance

	for _, point := range points {
		distance := s.CalculateDistance(center, point)
		if distance <= radiusKm {
			nearby = append(nearby, CoordinateWithDistance{
				Coordinates: point,
				Distance:    distance,
			})
		}
	}

	return nearby
}

// SortByDistance sorts coordinates by their distance from a center point
func (s *GeospatialService) SortByDistance(center Coordinates, points []Coordinates) []CoordinateWithDistance {
	var withDistances []CoordinateWithDistance

	for _, point := range points {
		distance := s.CalculateDistance(center, point)
		withDistances = append(withDistances, CoordinateWithDistance{
			Coordinates: point,
			Distance:    distance,
		})
	}

	// Sort by distance (ascending)
	for i := 0; i < len(withDistances)-1; i++ {
		for j := i + 1; j < len(withDistances); j++ {
			if withDistances[i].Distance > withDistances[j].Distance {
				withDistances[i], withDistances[j] = withDistances[j], withDistances[i]
			}
		}
	}

	return withDistances
}

// CalculateCenterPoint calculates the center point of multiple coordinates
func (s *GeospatialService) CalculateCenterPoint(coordinates []Coordinates) *Coordinates {
	if len(coordinates) == 0 {
		return nil
	}

	if len(coordinates) == 1 {
		return &coordinates[0]
	}

	var totalLat, totalLon float64

	for _, coord := range coordinates {
		totalLat += coord.Latitude
		totalLon += coord.Longitude
	}

	return &Coordinates{
		Latitude:  totalLat / float64(len(coordinates)),
		Longitude: totalLon / float64(len(coordinates)),
	}
}

// ConvertDegreesToRadians converts degrees to radians
func (s *GeospatialService) ConvertDegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// ConvertRadiansToDegrees converts radians to degrees
func (s *GeospatialService) ConvertRadiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// GetCompassBearing calculates the compass bearing from one coordinate to another
func (s *GeospatialService) GetCompassBearing(from, to Coordinates) float64 {
	lat1 := s.ConvertDegreesToRadians(from.Latitude)
	lat2 := s.ConvertDegreesToRadians(to.Latitude)
	deltaLon := s.ConvertDegreesToRadians(to.Longitude - from.Longitude)

	y := math.Sin(deltaLon) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(deltaLon)

	bearing := math.Atan2(y, x)
	bearing = s.ConvertRadiansToDegrees(bearing)

	// Normalize to 0-360 degrees
	if bearing < 0 {
		bearing += 360
	}

	return bearing
}

// BoundingBox represents a rectangular geographic area
type BoundingBox struct {
	NorthEast Coordinates `json:"north_east"`
	SouthWest Coordinates `json:"south_west"`
}

// Contains checks if a coordinate is within the bounding box
func (bb *BoundingBox) Contains(coord Coordinates) bool {
	return coord.Latitude >= bb.SouthWest.Latitude &&
		coord.Latitude <= bb.NorthEast.Latitude &&
		coord.Longitude >= bb.SouthWest.Longitude &&
		coord.Longitude <= bb.NorthEast.Longitude
}

// CoordinateWithDistance represents a coordinate with its distance from a reference point
type CoordinateWithDistance struct {
	Coordinates Coordinates `json:"coordinates"`
	Distance    float64     `json:"distance_km"`
}

// SearchArea represents a geographic search area
type SearchArea struct {
	Center   Coordinates `json:"center"`
	RadiusKm float64     `json:"radius_km"`
}

// Validate validates the search area
func (sa *SearchArea) Validate() error {
	if err := sa.Center.Validate(); err != nil {
		return err
	}

	if sa.RadiusKm <= 0 {
		return ErrInvalidRadius
	}

	if sa.RadiusKm > MaxSearchRadiusKm {
		return ErrRadiusTooLarge
	}

	return nil
}

// Contains checks if a coordinate is within the search area
func (sa *SearchArea) Contains(coord Coordinates) bool {
	distance := sa.Center.DistanceTo(coord)
	return distance <= sa.RadiusKm
}
