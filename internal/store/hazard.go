package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ethan-howes/talus/internal/store/models"
)

func GetSourceZones(ctx context.Context, pool *pgxpool.Pool, routeID int, radiusM float64) ([]models.SourceZone, error) {

	sql := `
		SELECT sz.id, sz.dem_tile_id, sz.mean_slope_deg, sz.mean_aspect_deg, sz.area_m2, sz.geology_id
		FROM source_zones sz
		JOIN routes r ON r.id = $1
		WHERE ST_DWithin(
		ST_Transform(sz.centroid, 32616),
		ST_Transform(r.geometry, 32616),
		$2
	)
	`

	rows, err := pool.Query(ctx, sql, routeID, radiusM)
	if err != nil {
		return nil, fmt.Errorf("failed to query source zone: %w", err)
	}
	defer rows.Close()

	var zones []models.SourceZone

	for rows.Next() {
		var z models.SourceZone
		err := rows.Scan(&z.ID, &z.DemTileID, &z.MeanSlopeDeg, &z.MeanAspectDeg, &z.AreaM2, &z.GeologyID)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		zones = append(zones, z)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return zones, nil
}

func InsertRouteRiskAssessment(ctx context.Context, pool *pgxpool.Pool, a models.RouteRiskAssessment) (int, error) {

	var id int
	err := pool.QueryRow(ctx,
		`INSERT INTO route_risk_assessments (route_id, source_zone_id, nearest_source_m, risk_score)
	VALUES ($1, $2, $3, $4) RETURNING id`,
		a.RouteID, a.SourceZoneID, a.NearestSourceM, a.RiskScore).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert route risk assessment: %w", err)
	}
	return id, nil
}

func InsertFreezeThawWindow(ctx context.Context, pool *pgxpool.Pool, f models.FreezeThawWindow) (int, error) {

	var id int
	err := pool.QueryRow(ctx,
		`INSERT INTO freeze_thaw_windows (source_zone_id, forecast_date, overnight_low_c, sun_exposure_time, freeze_thaw_active, risk_level)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		f.SourceZoneID, time.Now(), f.OvernightLowC, f.SunExposureTime, f.FreezeThawActive, f.RiskLevel).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert freeze thaw window: %w", err)
	}
	return id, nil
}
