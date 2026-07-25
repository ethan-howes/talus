// S1 main file

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ethan-howes/talus/internal/config"
	"github.com/ethan-howes/talus/internal/ingestion/geology"
	"github.com/ethan-howes/talus/internal/ingestion/geotiff"
	"github.com/ethan-howes/talus/internal/ingestion/gpx"
	"github.com/ethan-howes/talus/internal/store"
	"github.com/ethan-howes/talus/internal/store/models"
	"github.com/ethan-howes/talus/internal/telemetry"
)

type demRequest struct {
	FilePath string `json:"file_path"`
	Source   string `json:"source"`
}

type demResponse struct {
	DemTileID int    `json:"dem_tile_id"`
	TileCount int    `json:"tile_count"`
	Message   string `json:"message"`
}

type routeRequest struct {
	FilePath string `json:"file_path"`
	Name     string `json:"name"`
	Source   string `json:"source"`
}

type geologyRequest struct {
	DemTileID int    `json:"dem_tile_id"`
	RockType  string `json:"rock_type"`
	Geometry  string `json:"geometry"`
}

func main() {

	// load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// init logger
	logger := telemetry.NewLogger(cfg.LogLevel)
	logger.Info("starting ingestion service")

	// connect to db
	pool, err := store.Connect(cfg)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// run migrations
	err = store.RunMigrations(pool, "db/migrations")
	if err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	logger.Info("migrations complete")

	// start HTTP server
	http.HandleFunc("/dem", handleDem(pool, logger, cfg))
	http.HandleFunc("/routes", handleRoutes(pool, logger, cfg))
	http.HandleFunc("/geology", handleGeology(pool, logger, cfg))

	logger.Info("ingestion service lisening", "port", cfg.S1ListenPort)
	err = http.ListenAndServe(":"+cfg.S1ListenPort, nil)
	if err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func handleDem(pool *pgxpool.Pool, logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req demRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		result, err := geotiff.Inspect(req.FilePath)
		if err != nil {
			logger.Error("geotiff inspect failed", "error", err)
			http.Error(w, "geotiff processing failed", http.StatusInternalServerError)
			return
		}

		tiles := geotiff.ComputeTiles(result, cfg.TerrainTileMaxCells, cfg.TerrainTileHaloMeters)

		bbox := fmt.Sprintf(
			"POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
			result.MinLon, result.MinLat,
			result.MaxLon, result.MinLat,
			result.MaxLon, result.MaxLat,
			result.MinLon, result.MaxLat,
			result.MinLon, result.MinLat,
		)

		ctx := r.Context()
		demTileID, err := store.InsertDemTile(ctx, pool, models.DemTile{
			Filename:    result.Filename,
			BoundingBox: bbox,
			Crs:         result.CRS,
			ResolutionM: result.ResolutionM,
			Rows:        result.Rows,
			Cols:        result.Cols,
			FilePath:    req.FilePath,
			Source:      req.Source,
		})
		if err != nil {
			logger.Error("failed to insert dem tile", "error", err)
			http.Error(w, "database write failed", http.StatusInternalServerError)
			return
		}

		// notify s2 for each tile
		for _, tile := range tiles {
			tileBody, _ := json.Marshal(map[string]interface{}{
				"tile_path":    req.FilePath,
				"output_dir":   cfg.DemStoragePath,
				"rows":         tile.Rows,
				"cols":         tile.Cols,
				"cell_size":    result.ResolutionM,
				"dem_tile_id":  demTileID,
				"geology_id":   1,
				"origin_lon":   tile.OriginLon,
				"origin_lat":   tile.OriginLat,
				"slope_thresh": cfg.TerrainSlopeThreshDeg,
				"tri_thresh":   cfg.TerrainTriThreshold,
			})
			http.Post(cfg.S2TerrainEndpoint+"/tile", "application/json", bytes.NewReader(tileBody))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(demResponse{
			DemTileID: demTileID,
			TileCount: len(tiles),
			Message:   "DEM ingested successfully",
		})

	}
}

func handleRoutes(pool *pgxpool.Pool, logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req routeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		result, err := gpx.Parse(req.FilePath)
		if err != nil {
			logger.Error("gpx parse failed", "error", err)
			http.Error(w, "gpx processing failed", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		routeID, err := store.InsertRoute(ctx, pool, models.Route{
			Name:     req.Name,
			Geometry: result,
			Source:   req.Source,
		})
		if err != nil {
			logger.Error("failed to insert route", "error", err)
			http.Error(w, "databse write failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"route_id": routeID,
			"message":  "route ingested successfully",
		})
	}
}

func handleGeology(pool *pgxpool.Pool, logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req geologyRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		params, ok := geology.Lookup(req.RockType)
		if !ok {
			http.Error(w, "unknown rock type", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		geologyID, err := store.InsertGeology(ctx, pool, models.Geology{
			DemTileID:      req.DemTileID,
			Geometry:       req.Geometry,
			RockType:       req.RockType,
			BounceCoeff:    params.BounceCoeff,
			FrictionCoeff:  params.FrictionCoeff,
			FragmentationK: params.FragmentationK,
		})
		if err != nil {
			logger.Error("failed to insert geology", "error", err)
			http.Error(w, "database write failed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"geology_id": geologyID,
			"message":    "geology ingested successfully",
		})
	}
}
