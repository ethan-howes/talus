package proximity

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ethan-howes/talus/internal/store/models"
)

func FindProximateSourceZones(ctx context.Context, pool *pgxpool.Pool, routeID int, radiusM float64, freezeThawActive bool) ([]models.RouteRiskAssessment, error) {

	// postGIS query
	sql := `
		SELECT 
			sz.id,
			sz.mean_slope_deg,
			ST_Distance(
				ST_Transform(sz.centroid, 32616),
				ST_Transform(r.geometry, 32616)
			) AS distance_m
		FROM source_zones sz
		JOIN routes r ON r.id = $1
		WHERE ST_DWithin(
			ST_Transform(sz.centroid, 32616),
			ST_Transform(r.geometry, 32616),
			$2
		)
		ORDER BY distance_m ASC
	`

	rows, err := pool.Query(ctx, sql, routeID, radiusM)
	if err != nil {
		return nil, fmt.Errorf("proximity query failed: %w", err)
	}
	defer rows.Close()

	var assessments []models.RouteRiskAssessment

	for rows.Next() {
		var sourceZoneID int
		var meanSlopeDeg float64
		var distanceM float64

		err := rows.Scan(&sourceZoneID, &meanSlopeDeg, &distanceM)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		proximityWeight := 1.0 - (distanceM / radiusM)
		slopeWeight := (meanSlopeDeg - 45.0) / (70.0 - 45.0)

		if slopeWeight < 0 {
			slopeWeight = 0
		}
		if slopeWeight > 1 {
			slopeWeight = 1
		}

		freezeThawMultiplier := 1.0
		if freezeThawActive {
			freezeThawMultiplier = 1.5
		}

		riskScore := proximityWeight * slopeWeight * freezeThawMultiplier

		assessments = append(assessments, models.RouteRiskAssessment{
			RouteID:        routeID,
			SourceZoneID:   sourceZoneID,
			NearestSourceM: distanceM,
			RiskScore:      riskScore,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return assessments, nil
}
