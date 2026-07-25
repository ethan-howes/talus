package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ethan-howes/talus/internal/config"
)

type Handlers struct {
	Pool   *pgxpool.Pool
	Logger *slog.Logger
	Cfg    *config.Config
}

// /dem to s1
func (h *Handlers) HandleDem(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post(h.Cfg.S1IngestionEndpoint+"/dem", "application/json", r.Body)
	if err != nil {
		h.Logger.Error("service 1 unavailable", "error", err)
		http.Error(w, "upstream service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// /routes to s1
func (h *Handlers) HandleRoutes(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post(h.Cfg.S1IngestionEndpoint+"/routes", "application/json", r.Body)
	if err != nil {
		h.Logger.Error("service 1 unavailable", "error", err)
		http.Error(w, "upstream service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// /analyze to s4
func (h *Handlers) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post(h.Cfg.S4HazardEndpoint+"/analyze", "application/json", r.Body)
	if err != nil {
		h.Logger.Error("service 4 unavailable", "error", err)
		http.Error(w, "upstream service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// queries db for alerts and returns json from the table
func (h *Handlers) HandleGetAlerts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Pool.Query(r.Context(), `SELECT id, route_id, risk_score, summary, triggered_at FROM alert_events ORDER BY triggered_at DESC`)
	if err != nil {
		h.Logger.Error("query failed", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id, routeID int
		var riskScore float64
		var summary string
		var triggeredAt time.Time
		rows.Scan(&id, &routeID, &riskScore, &summary, &triggeredAt)
		events = append(events, map[string]interface{}{
			"id":           id,
			"route_id":     routeID,
			"risk_score":   riskScore,
			"summary":      summary,
			"triggered_at": triggeredAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// returns json values for route_risk_assessments db
func (h *Handlers) HandleGetRouteRisk(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Pool.Query(r.Context(), `SELECT id, route_id, source_zone_id, nearest_source_m, risk_score, assessed_at FROM route_risk_assessments ORDER BY assessed_at DESC`)
	if err != nil {
		h.Logger.Error("query failed", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id, routeID, sourceZoneID int
		var nearestSourceM, riskScore float64
		var assessedAt time.Time
		rows.Scan(&id, &routeID, &sourceZoneID, &nearestSourceM, &riskScore, &assessedAt)
		events = append(events, map[string]interface{}{
			"id":               id,
			"route_id":         routeID,
			"source_zone_id":   sourceZoneID,
			"nearest_source_m": nearestSourceM,
			"risk_score":       riskScore,
			"assessed_at":      assessedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// returns json values for freeze_thaw_windows db
func (h *Handlers) HandleGetFreezeThaw(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Pool.Query(r.Context(), `SELECT id, source_zone_id, forecast_date, overnight_low_c, sun_exposure_time, freeze_thaw_active, risk_level FROM freeze_thaw_windows`)
	if err != nil {
		h.Logger.Error("query failed", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id, sourceZoneID int
		var forecastDate time.Time
		var overnightLowC float64
		var sunExposureTime time.Time
		var freezeThawActive bool
		var riskLevel string
		rows.Scan(&id, &sourceZoneID, &forecastDate, &overnightLowC, &sunExposureTime, &freezeThawActive, &riskLevel)
		events = append(events, map[string]interface{}{
			"id":                 id,
			"source_zone_id":     sourceZoneID,
			"forecast_date":      forecastDate,
			"overnight_low_c":    overnightLowC,
			"sun_exposure_time":  sunExposureTime,
			"freeze_thaw_active": freezeThawActive,
			"risk_level":         riskLevel,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// returns json values for terrain_metrics db
func (h *Handlers) HandleGetTerrainMetrics(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Pool.Query(r.Context(), `SELECT id, dem_tile_id, kernel_name, cells_processed, gpu_time_ms, cpu_time_ms, throughput_mcps, recorded_at FROM terrain_metrics`)
	if err != nil {
		h.Logger.Error("query failed", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id, demTileID int
		var kernelName string
		var cellsProcessed int64
		var gpuTimeMs, cpuTimeMs, throughputMcps float64
		var recordedAt time.Time
		rows.Scan(&id, &demTileID, &kernelName, &cellsProcessed, &gpuTimeMs, &cpuTimeMs, &throughputMcps, &recordedAt)
		events = append(events, map[string]interface{}{
			"id":              id,
			"dem_tile_id":     demTileID,
			"kernel_name":     kernelName,
			"cells_processed": cellsProcessed,
			"gpu_time_ms":     gpuTimeMs,
			"cpu_time_ms":     cpuTimeMs,
			"throughput_mcps": throughputMcps,
			"recorded_at":     recordedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (h *Handlers) HandlePostAlertConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string  `json:"name"`
		RouteID           int     `json:"route_id"`
		RiskThreshold     float64 `json:"risk_threshold"`
		FreezeThawTrigger bool    `json:"freeze_thaw_trigger"`
		WebhookURL        string  `json:"webhook_url"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.Logger.Error("invalid request body", "error", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var id int
	err = h.Pool.QueryRow(r.Context(),
		`INSERT INTO alert_configs (name, route_id, risk_threshold, freeze_thaw_trigger, webhook_url)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		req.Name, req.RouteID, req.RiskThreshold, req.FreezeThawTrigger, req.WebhookURL,
	).Scan(&id)
	if err != nil {
		h.Logger.Error("failed to insert alert config", "error", err)
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      id,
		"message": "alert config created",
	})
}
