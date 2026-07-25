package store

import (
	"context"
	"fmt"

	"github.com/ethan-howes/talus/internal/store/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertTerrainDerivative(ctx context.Context, pool *pgxpool.Pool, d models.TerrainDerivative) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		"INSERT INTO terrain_derivatives (dem_tile_id, slope_path, aspect_path, curvature_path, tri_path) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		d.DemTileID, d.SlopePath, d.AspectPath, d.CurvaturePath, d.TriPath).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert terrain derivative: %w", err)
	}
	return id, nil
}

func InsertSourceZone(ctx context.Context, pool *pgxpool.Pool, z models.SourceZone) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		"INSERT INTO source_zones (dem_tile_id, geometry, centroid, mean_slope_deg, mean_aspect_deg, area_m2, geology_id) VALUES ($1, ST_GeomFromText($2, 4326), ST_GeomFromText($3, 4326), $4, $5, $6, $7) RETURNING id",
		z.DemTileID, z.Geometry, z.Centroid, z.MeanSlopeDeg, z.MeanAspectDeg, z.AreaM2, z.GeologyID).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert source zone: %w", err)
	}
	return id, nil
}

func InsertTerrainMetric(ctx context.Context, pool *pgxpool.Pool, m models.TerrainMetric) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		"INSERT INTO terrain_metrics (dem_tile_id, kernel_name, cells_processed, gpu_time_ms, cpu_time_ms, throughput_mcps) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		m.DemTileID, m.KernelName, m.CellsProcessed, m.GpuTimeMs, m.CpuTimeMs, m.ThroughputMcps).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert terrain metric: %w", err)
	}
	return id, nil
}
