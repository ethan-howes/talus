package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	PostgresHost          string
	PostgresPort          int
	PostgresDB            string
	PostgresUser          string
	PostgresPassword      string
	PostgresSSLMode       string
	LogLevel              string
	S1ListenPort          string
	S2ListenPort          string
	S4ListenPort          string
	S5ListenPort          string
	CudaTerrainBinaryPath string
	S1IngestionEndpoint   string
	S2TerrainEndpoint     string
	S4HazardEndpoint      string
	DemStoragePath        string
	TerrainSlopeThreshDeg float64
	TerrainTriThreshold   float64
	TerrainTileHaloMeters float64
	TerrainTileMaxCells   int
	ProximityRadiusM      float64
	WebStaticDir          string
}

func Load() (*Config, error) {

	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		return nil, fmt.Errorf("POSTGRES_HOST is required but not set")
	}

	db := os.Getenv("POSTGRES_DB")
	if db == "" {
		return nil, fmt.Errorf("POSTGRES_DB is required but not set")
	}

	user := os.Getenv("POSTGRES_USER")
	if user == "" {
		return nil, fmt.Errorf("POSTGRES_USER is required but not set")
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("POSTGRES_PASSWORD is required but not set")
	}

	sslmode := os.Getenv("POSTGRES_SSLMODE")
	if sslmode == "" {
		return nil, fmt.Errorf("POSTGRES_SSLMODE is required but not set")
	}

	portStr := os.Getenv("POSTGRES_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("POSTGRES_PORT must be a number: %w", err)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	s1Port := os.Getenv("S1_LISTEN_PORT")
	if s1Port == "" {
		return nil, fmt.Errorf("S1_LISTEN_PORT is required but not set")
	}

	s2Port := os.Getenv("S2_LISTEN_PORT")
	if s2Port == "" {
		return nil, fmt.Errorf("S2_LISTEN_PORT is required but not set")
	}

	s4Port := os.Getenv("S4_LISTEN_PORT")
	if s4Port == "" {
		return nil, fmt.Errorf("S4_LISTEN_PORT is required but not set")
	}

	s5Port := os.Getenv("S5_LISTEN_PORT")
	if s5Port == "" {
		return nil, fmt.Errorf("S5_LISTEN_PORT is required but not set")
	}

	cudaBinary := os.Getenv("CUDA_TERRAIN_BINARY_PATH")
	if cudaBinary == "" {
		return nil, fmt.Errorf("CUDA_TERRAIN_BINARY_PATH is required but not set")
	}

	s1Endpoint := os.Getenv("S1_INGESTION_ENDPOINT")
	if s1Endpoint == "" {
		return nil, fmt.Errorf("S2_INGESTION_ENDPOINT is required but not set")
	}

	s2Endpoint := os.Getenv("S2_TERRAIN_ENDPOINT")
	if s2Endpoint == "" {
		return nil, fmt.Errorf("S2_TERRAIN_ENDPOINT is required but not set")
	}

	s4Endpoint := os.Getenv("S4_HAZARD_ENDPOINT")
	if s4Endpoint == "" {
		return nil, fmt.Errorf("S4_HAZARD_ENDPOINT is required but not set")
	}

	demStoragePath := os.Getenv("DEM_STORAGE_PATH")
	if demStoragePath == "" {
		return nil, fmt.Errorf("DEM_STORAGE_PATH is required but not set")
	}

	terrainSlopeThreshDeg := os.Getenv("TERRAIN_SLOPE_THRESHOLD_DEG")
	if terrainSlopeThreshDeg == "" {
		return nil, fmt.Errorf("TERRAIN_SLOPE_THRESHOLD_DEG is required but not set")
	}

	terrainTriThreshold := os.Getenv("TERRAIN_TRI_THRESHOLD")
	if terrainTriThreshold == "" {
		return nil, fmt.Errorf("TERRAIN_TRI_THRESHOLD is required but not set")
	}

	terrainTileHaloMeters := os.Getenv("TERRAIN_TILE_HALO_METERS")
	if terrainTileHaloMeters == "" {
		return nil, fmt.Errorf("TERRAIN_TILE_HALO_METERS is required but not set")
	}

	terrainTileMaxCells := os.Getenv("TERRAIN_TILE_MAX_CELLS_PER_SIDE")
	if terrainTileMaxCells == "" {
		return nil, fmt.Errorf("TERRAIN_TILE_MAX_CELLS_PER_SIDE is required but not set")
	}

	proximityRadiusStr := os.Getenv("SOURCE_ZONE_PROXIMITY_RADIUS_M")
	if proximityRadiusStr == "" {
		return nil, fmt.Errorf("SOURCE_ZONE_PROXIMITY_RADIUS_M is required but not set")
	}

	webstaticdir := os.Getenv("WEB_STATIC_DIR")
	if webstaticdir == "" {
		return nil, fmt.Errorf("WEB_STATIC_DIR is required but not set")
	}

	slopeThresh, err := strconv.ParseFloat(terrainSlopeThreshDeg, 64)
	if err != nil {
		return nil, fmt.Errorf("TERRAIN_SLOPE_THRESH_DEG must be a number: %w", err)
	}

	triThresh, err := strconv.ParseFloat(terrainTriThreshold, 64)
	if err != nil {
		return nil, fmt.Errorf("TERRAIN_TRI_THRESH must be a number: %w", err)
	}

	haloMeters, err := strconv.ParseFloat(terrainTileHaloMeters, 64)
	if err != nil {
		return nil, fmt.Errorf("TERRAIN_TILE_HALO_METERS must be a number: %w", err)
	}

	maxCells, err := strconv.Atoi(terrainTileMaxCells)
	if err != nil {
		return nil, fmt.Errorf("TERRAIN_TILE_MAX_CELLS must be a number: %w", err)
	}

	proximityRadius, err := strconv.ParseFloat(proximityRadiusStr, 64)
	if err != nil {
		return nil, fmt.Errorf("SOURCE_ZONE_PROXIMITY_RADIUS_M must be a number: %w", err)
	}

	return &Config{
		PostgresHost:          host,
		PostgresPort:          port,
		PostgresDB:            db,
		PostgresUser:          user,
		PostgresPassword:      password,
		PostgresSSLMode:       sslmode,
		LogLevel:              logLevel,
		S1ListenPort:          s1Port,
		S2ListenPort:          s2Port,
		S4ListenPort:          s4Port,
		S5ListenPort:          s5Port,
		CudaTerrainBinaryPath: cudaBinary,
		S1IngestionEndpoint:   s1Endpoint,
		S2TerrainEndpoint:     s2Endpoint,
		S4HazardEndpoint:      s4Endpoint,
		DemStoragePath:        demStoragePath,
		TerrainSlopeThreshDeg: slopeThresh,
		TerrainTriThreshold:   triThresh,
		TerrainTileHaloMeters: haloMeters,
		TerrainTileMaxCells:   maxCells,
		ProximityRadiusM:      proximityRadius,
		WebStaticDir:          webstaticdir,
	}, nil
}
