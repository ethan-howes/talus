package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ethan-howes/talus/internal/config"
	"github.com/ethan-howes/talus/internal/store"
	"github.com/ethan-howes/talus/internal/store/models"
	"github.com/ethan-howes/talus/internal/telemetry"
	"github.com/ethan-howes/talus/internal/terrain/subprocess"
	"github.com/ethan-howes/talus/internal/terrain/zones"
)

type tileRequest struct {
	TilePath    string  `json:"tile_path"`
	OutputDir   string  `json:"output_dir"`
	Rows        int     `json:"rows"`
	Cols        int     `json:"cols"`
	CellSize    float64 `json:"cell_size"`
	DemTileID   int     `json:"dem_tile_id"`
	GeologyID   int     `json:"geology_id"`
	OriginLon   float64 `json:"origin_lon"`
	OriginLat   float64 `json:"origin_lat"`
	SlopeThresh float64 `json:"slope_thresh"`
	TriThresh   float64 `json:"tri_thresh"`
}

type tileResponse struct {
	SourceZoneCount int    `json:"source_zone_count"`
	Message         string `json:"message"`
}

func main() {

	// load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// init logger
	logger := telemetry.NewLogger(cfg.LogLevel)
	logger.Info("starting terrain service")

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

	// start an HTTP server to listen to service 1
	http.HandleFunc("/tile", handleTile(pool, logger, cfg))
	logger.Info("terrain service listening", "port", cfg.S2ListenPort)
	err = http.ListenAndServe(":"+cfg.S2ListenPort, nil)
	if err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func handleTile(pool *pgxpool.Pool, logger *slog.Logger, cfg *config.Config) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		var req tileRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		result, err := subprocess.RunTerrain(
			cfg.CudaTerrainBinaryPath,
			req.TilePath,
			req.OutputDir,
			req.Rows,
			req.Cols,
			req.CellSize,
		)
		if err != nil {
			logger.Error("cuda binary failed", "error", err)
			http.Error(w, "terrain processing failed", http.StatusInternalServerError)
			return
		}

		detectedZones, err := zones.DetectSourceZones(
			result.SlopePath,
			result.PlanPath,
			result.TriPath,
			req.Rows,
			req.Cols,
			req.CellSize,
			req.SlopeThresh,
			req.TriThresh,
			req.OriginLon,
			req.OriginLat,
			req.DemTileID,
			req.GeologyID,
		)
		if err != nil {
			logger.Error("source zone detection failed", "error", err)
			http.Error(w, "zone detection failed", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()

		_, err = store.InsertTerrainDerivative(ctx, pool, models.TerrainDerivative{
			DemTileID:     req.DemTileID,
			SlopePath:     result.SlopePath,
			AspectPath:    result.AspectPath,
			CurvaturePath: result.PlanPath,
			TriPath:       result.TriPath,
		})
		if err != nil {
			logger.Error("failed to insert terrain derivative", "error", err)
			http.Error(w, "database write failed", http.StatusInternalServerError)
			return
		}

		for _, z := range detectedZones {
			_, err = store.InsertSourceZone(ctx, pool, z)
			if err != nil {
				logger.Error("failed to insert source pool", "error", err)
				http.Error(w, "database write failed", http.StatusInternalServerError)
				return
			}
		}

		kernels := []struct {
			name string
			gpu  float64
			cpu  float64
		}{
			{"slope_aspect", result.GpuTimeMsSlopeAspect, result.CpuTimeMsSlopeAspect},
			{"curvature", result.GpuTimeMsCurvature, result.CpuTimeMsCurvature},
			{"tri", result.GpuTimeMsTri, result.CpuTimeMsTri},
		}

		cells := int64(req.Rows * req.Cols)

		for _, k := range kernels {
			_, err = store.InsertTerrainMetric(ctx, pool, models.TerrainMetric{
				DemTileID:      req.DemTileID,
				KernelName:     k.name,
				CellsProcessed: cells,
				GpuTimeMs:      k.gpu,
				CpuTimeMs:      k.cpu,
				ThroughputMcps: float64(cells) / k.gpu / 1000.0,
			})
			if err != nil {
				logger.Error("failed to insert terrain metric", "error", err)
				http.Error(w, "database failed to write", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tileResponse{
			SourceZoneCount: len(detectedZones),
			Message:         "terrain processing complete",
		})
	}
}
