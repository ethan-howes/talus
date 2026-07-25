package models

import "time"

type DemTile struct {
	ID          int
	Filename    string
	BoundingBox string
	Crs         string
	ResolutionM float64
	Rows        int
	Cols        int
	FilePath    string
	Source      string
	IngestedAt  time.Time
}

type Geology struct {
	ID             int
	DemTileID      int
	Geometry       string
	RockType       string
	BounceCoeff    float64
	FrictionCoeff  float64
	FragmentationK float64
}

type Route struct {
	ID         int
	Name       string
	Geometry   string
	Source     string
	UploadedAt time.Time
}

type TerrainDerivative struct {
	ID            int
	DemTileID     int
	SlopePath     string
	AspectPath    string
	CurvaturePath string
	TriPath       string
	ComputedAt    time.Time
}

type SourceZone struct {
	ID            int
	DemTileID     int
	Geometry      string
	Centroid      string
	MeanSlopeDeg  float64
	MeanAspectDeg float64
	AreaM2        float64
	GeologyID     int
}

type TerrainMetric struct {
	ID             int
	DemTileID      int
	KernelName     string
	CellsProcessed int64
	GpuTimeMs      float64
	CpuTimeMs      float64
	ThroughputMcps float64
	RecordedAt     time.Time
}

type RouteRiskAssessment struct {
	ID               int
	RouteID          int
	SourceZoneID     int
	NearestSourceM   float64
	ExposedLengthM   float64
	MaxPassageProb   float64
	IntersectionGeom string
	RiskScore        float64
	AssessedAt       time.Time
}

type FreezeThawWindow struct {
	ID               int
	SourceZoneID     int
	ForecastDate     time.Time
	OvernightLowC    float64
	SunExposureTime  time.Time
	FreezeThawActive bool
	RiskLevel        string
}

type AlertConfig struct {
	ID                int
	Name              string
	RouteID           int
	RiskThreshold     float64
	FreezeThawTrigger bool
	WebhookURL        string
	Enabled           bool
	CreatedAt         time.Time
}

type AlertEvent struct {
	ID               int
	ConfigID         int
	RouteID          int
	TriggeredAt      time.Time
	RiskScore        float64
	Summary          string
	FreezeThawActive bool
}
