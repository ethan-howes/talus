package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ethan-howes/talus/internal/config"
	"github.com/ethan-howes/talus/internal/hazard/alerts"
	"github.com/ethan-howes/talus/internal/hazard/freezethaw"
	"github.com/ethan-howes/talus/internal/hazard/proximity"
	"github.com/ethan-howes/talus/internal/store"
	"github.com/ethan-howes/talus/internal/store/models"
	"github.com/ethan-howes/talus/internal/telemetry"
)

type analyzeRequest struct {
	RouteID int `json:"route_id"`
}

func main() {
	// load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// init logger
	logger := telemetry.NewLogger(cfg.LogLevel)
	logger.Info("starting hazard service")

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
	http.HandleFunc("/analyze", handleAnalyze(pool, logger, cfg))
	logger.Info("hazard service listening", "port", cfg.S4ListenPort)
	err = http.ListenAndServe(":"+cfg.S4ListenPort, nil)
	if err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func handleAnalyze(pool *pgxpool.Pool, logger *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req analyzeRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		sourceZones, err := store.GetSourceZones(ctx, pool, req.RouteID, cfg.ProximityRadiusM)
		if err != nil {
			logger.Error("failed to get source zones", "error", err)
			http.Error(w, "database query failed", http.StatusInternalServerError)
			return
		}

		for _, sz := range sourceZones {
			prediction, err := freezethaw.PredictFreezeThaw(
				ctx,
				-86.0, -85.0, 35.0, 36.0,
				553.0,
				sz.MeanAspectDeg,
			)
			if err != nil {
				logger.Error("freeze thaw prediction failed", "error", err)
				continue
			}
			_, err = store.InsertFreezeThawWindow(ctx, pool, models.FreezeThawWindow{
				SourceZoneID:     sz.ID,
				OvernightLowC:    prediction.OvernightLowC,
				SunExposureTime:  time.Now(),
				FreezeThawActive: prediction.FreezeThawActive,
				RiskLevel:        prediction.RiskLevel,
			})
			if err != nil {
				logger.Error("failed to insert freeze thaw window", "error", err)
			}
		}

		freezeThawActive := false
		if len(sourceZones) > 0 {
			prediction, _ := freezethaw.PredictFreezeThaw(ctx, -86.0, -85.0, 35.0, 36.0, 553.0, sourceZones[0].MeanAspectDeg)
			if prediction != nil {
				freezeThawActive = prediction.FreezeThawActive
			}
		}

		assessments, err := proximity.FindProximateSourceZones(ctx, pool, req.RouteID, cfg.ProximityRadiusM, freezeThawActive)
		if err != nil {
			logger.Error("proximity query failed", "error", err)
			http.Error(w, "proximity query failed", http.StatusInternalServerError)
			return
		}

		for _, a := range assessments {
			_, err = store.InsertRouteRiskAssessment(ctx, pool, a)
			if err != nil {
				logger.Error("failed to insert risk assessment", "error", err)
				continue
			}
			summary := fmt.Sprintf("route %d risk score %.2f", req.RouteID, a.RiskScore)
			err = alerts.EvaluateAndFire(ctx, pool, req.RouteID, a.RiskScore, freezeThawActive, summary)
			if err != nil {
				logger.Error("alert evaluation failed", "error", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"route_id":           req.RouteID,
			"assessments":        len(assessments),
			"freeze_thaw_active": freezeThawActive,
			"message":            "hazard analysis complete",
		})
	}
}
