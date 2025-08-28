package domain

import (
	"math"
	"testing"
)

func TestGeospatialService_CalculateDistance(t *testing.T) {
	service := NewGeospatialService()

	lisbon := Coordinates{Latitude: 38.7223, Longitude: -9.1393}
	porto := Coordinates{Latitude: 41.1579, Longitude: -8.6291}

	distance := service.CalculateDistance(lisbon, porto)

	// Distance between Lisbon and Porto is approximately 274 km
	expectedDistance := 274.0
	tolerance := 10.0

	if math.Abs(distance-expectedDistance) > tolerance {
		t.Errorf("Distance calculation incorrect. Got %f, expected approximately %f", distance, expectedDistance)
	}

	// Test distance to self should be 0
	selfDistance := service.CalculateDistance(lisbon, lisbon)
	if selfDistance > 0.001 {
		t.Errorf("Distance to self should be 0, got %f", selfDistance)
	}
}

func TestGeospatialService_IsWithinRadius(t *testing.T) {
	service := NewGeospatialService()

	center := Coordinates{Latitude: 38.7223, Longitude: -9.1393}
	nearby := Coordinates{Latitude: 38.7323, Longitude: -9.1293}  // ~1.5 km away
	faraway := Coordinates{Latitude: 41.1579, Longitude: -8.6291} // ~274 km away

	tests := []struct {
		name     string
		point    Coordinates
		radiusKm float64
		want     bool
	}{
		{"point within radius", nearby, 5.0, true},
		{"point outside radius", faraway, 5.0, false},
		{"point exactly at radius", nearby, 1.5, true},
		{"zero radius", nearby, 0.0, false},
		{"negative radius", nearby, -1.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.IsWithinRadius(center, tt.point, tt.radiusKm)
			if got != tt.want {
				t.Errorf("IsWithinRadius() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeospatialService_ValidateSearchRadius(t *testing.T) {
	service := NewGeospatialService()

	tests := []struct {
		name     string
		radiusKm int
		wantErr  error
	}{
		{"valid radius", 50, nil},
		{"minimum valid radius", 1, nil},
		{"maximum valid radius", 1000, nil},
		{"zero radius", 0, ErrInvalidRadius},
		{"negative radius", -10, ErrInvalidRadius},
		{"radius too large", 1001, ErrRadiusTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSearchRadius(tt.radiusKm)
			if err != tt.wantErr {
				t.Errorf("ValidateSearchRadius() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeospatialService_CalculateBoundingBox(t *testing.T) {
	service := NewGeospatialService()

	center := Coordinates{Latitude: 38.7223, Longitude: -9.1393}
	radiusKm := 10.0

	bbox := service.CalculateBoundingBox(center, radiusKm)

	// Check that bounding box contains the center
	if !bbox.Contains(center) {
		t.Error("Bounding box should contain the center point")
	}

	// Check that bounding box is reasonable (not too small or too large)
	latDelta := bbox.NorthEast.Latitude - bbox.SouthWest.Latitude
	lonDelta := bbox.NorthEast.Longitude - bbox.SouthWest.Longitude

	if latDelta <= 0 || lonDelta <= 0 {
		t.Error("Bounding box should have positive dimensions")
	}

	// Rough check - for 10km radius, lat delta should be around 0.18 degrees
	expectedLatDelta := 0.18
	tolerance := 0.05
	if math.Abs(latDelta-expectedLatDelta) > tolerance {
		t.Errorf("Latitude delta %f seems incorrect for 10km radius", latDelta)
	}
}

func TestGeospatialService_GetNearbyPoints(t *testing.T) {
	service := NewGeospatialService()

	center := Coordinates{Latitude: 38.7223, Longitude: -9.1393}
	points := []Coordinates{
		{Latitude: 38.7323, Longitude: -9.1293}, // ~1.5 km away
		{Latitude: 38.7123, Longitude: -9.1493}, // ~1.5 km away
		{Latitude: 41.1579, Longitude: -8.6291}, // ~274 km away
		{Latitude: 38.7223, Longitude: -9.1393}, // Same as center
	}

	nearby := service.GetNearbyPoints(center, points, 5.0)

	// Should find 3 points within 5km (including the center point)
	if len(nearby) != 3 {
		t.Errorf("Expected 3 nearby points, got %d", len(nearby))
	}

	// Check that all returned points have distance <= 5km
	for _, point := range nearby {
		if point.Distance > 5.0 {
			t.Errorf("Point distance %f exceeds radius 5.0", point.Distance)
		}
	}
}

func TestGeospatialService_SortByDistance(t *testing.T) {
	service := NewGeospatialService()

	center := Coordinates{Latitude: 38.7223, Longitude: -9.1393}
	points := []Coordinates{
		{Latitude: 41.1579, Longitude: -8.6291}, // ~274 km away (farthest)
		{Latitude: 38.7223, Longitude: -9.1393}, // Same as center (closest)
		{Latitude: 38.7323, Longitude: -9.1293}, // ~1.5 km away (middle)
	}

	sorted := service.SortByDistance(center, points)

	if len(sorted) != 3 {
		t.Errorf("Expected 3 sorted points, got %d", len(sorted))
	}

	// Check that points are sorted by distance (ascending)
	for i := 1; i < len(sorted); i++ {
		if sorted[i-1].Distance > sorted[i].Distance {
			t.Error("Points are not sorted by distance")
		}
	}

	// First point should be the center (distance ~0)
	if sorted[0].Distance > 0.001 {
		t.Errorf("First point should be at center, distance = %f", sorted[0].Distance)
	}
}

func TestGeospatialService_CalculateCenterPoint(t *testing.T) {
	service := NewGeospatialService()

	tests := []struct {
		name        string
		coordinates []Coordinates
		want        *Coordinates
	}{
		{
			name:        "empty slice",
			coordinates: []Coordinates{},
			want:        nil,
		},
		{
			name: "single point",
			coordinates: []Coordinates{
				{Latitude: 38.7223, Longitude: -9.1393},
			},
			want: &Coordinates{Latitude: 38.7223, Longitude: -9.1393},
		},
		{
			name: "two points",
			coordinates: []Coordinates{
				{Latitude: 38.0, Longitude: -9.0},
				{Latitude: 40.0, Longitude: -8.0},
			},
			want: &Coordinates{Latitude: 39.0, Longitude: -8.5},
		},
		{
			name: "multiple points",
			coordinates: []Coordinates{
				{Latitude: 38.0, Longitude: -9.0},
				{Latitude: 39.0, Longitude: -8.0},
				{Latitude: 40.0, Longitude: -7.0},
			},
			want: &Coordinates{Latitude: 39.0, Longitude: -8.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CalculateCenterPoint(tt.coordinates)

			if tt.want == nil {
				if got != nil {
					t.Errorf("Expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("Expected %v, got nil", tt.want)
				return
			}

			tolerance := 0.0001
			if math.Abs(got.Latitude-tt.want.Latitude) > tolerance ||
				math.Abs(got.Longitude-tt.want.Longitude) > tolerance {
				t.Errorf("CalculateCenterPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeospatialService_GetCompassBearing(t *testing.T) {
	service := NewGeospatialService()

	center := Coordinates{Latitude: 38.7223, Longitude: -9.1393}

	// Point directly north
	north := Coordinates{Latitude: 39.7223, Longitude: -9.1393}
	bearingNorth := service.GetCompassBearing(center, north)

	// Should be close to 0 degrees (north)
	if math.Abs(bearingNorth) > 1.0 {
		t.Errorf("Bearing to north should be ~0°, got %f°", bearingNorth)
	}

	// Point directly east
	east := Coordinates{Latitude: 38.7223, Longitude: -8.1393}
	bearingEast := service.GetCompassBearing(center, east)

	// Should be close to 90 degrees (east)
	if math.Abs(bearingEast-90.0) > 1.0 {
		t.Errorf("Bearing to east should be ~90°, got %f°", bearingEast)
	}
}

func TestBoundingBox_Contains(t *testing.T) {
	bbox := BoundingBox{
		NorthEast: Coordinates{Latitude: 40.0, Longitude: -8.0},
		SouthWest: Coordinates{Latitude: 38.0, Longitude: -10.0},
	}

	tests := []struct {
		name  string
		coord Coordinates
		want  bool
	}{
		{"point inside", Coordinates{Latitude: 39.0, Longitude: -9.0}, true},
		{"point on boundary", Coordinates{Latitude: 40.0, Longitude: -8.0}, true},
		{"point outside north", Coordinates{Latitude: 41.0, Longitude: -9.0}, false},
		{"point outside south", Coordinates{Latitude: 37.0, Longitude: -9.0}, false},
		{"point outside east", Coordinates{Latitude: 39.0, Longitude: -7.0}, false},
		{"point outside west", Coordinates{Latitude: 39.0, Longitude: -11.0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bbox.Contains(tt.coord)
			if got != tt.want {
				t.Errorf("BoundingBox.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchArea_Validate(t *testing.T) {
	tests := []struct {
		name       string
		searchArea SearchArea
		wantErr    error
	}{
		{
			name: "valid search area",
			searchArea: SearchArea{
				Center:   Coordinates{Latitude: 38.7223, Longitude: -9.1393},
				RadiusKm: 50.0,
			},
			wantErr: nil,
		},
		{
			name: "invalid coordinates",
			searchArea: SearchArea{
				Center:   Coordinates{Latitude: 91.0, Longitude: 0.0},
				RadiusKm: 50.0,
			},
			wantErr: ErrInvalidCoordinates,
		},
		{
			name: "zero radius",
			searchArea: SearchArea{
				Center:   Coordinates{Latitude: 38.7223, Longitude: -9.1393},
				RadiusKm: 0.0,
			},
			wantErr: ErrInvalidRadius,
		},
		{
			name: "radius too large",
			searchArea: SearchArea{
				Center:   Coordinates{Latitude: 38.7223, Longitude: -9.1393},
				RadiusKm: 1001.0,
			},
			wantErr: ErrRadiusTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.searchArea.Validate()
			if err != tt.wantErr {
				t.Errorf("SearchArea.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSearchArea_Contains(t *testing.T) {
	searchArea := SearchArea{
		Center:   Coordinates{Latitude: 38.7223, Longitude: -9.1393},
		RadiusKm: 5.0,
	}

	nearby := Coordinates{Latitude: 38.7323, Longitude: -9.1293}  // ~1.5 km away
	faraway := Coordinates{Latitude: 41.1579, Longitude: -8.6291} // ~274 km away

	if !searchArea.Contains(nearby) {
		t.Error("Search area should contain nearby point")
	}

	if searchArea.Contains(faraway) {
		t.Error("Search area should not contain faraway point")
	}
}
