package store

import (
	"context"
	"fmt"

	"github.com/ethan-howes/talus/internal/store/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertDemTile(ctx context.Context, pool *pgxpool.Pool, d models.DemTile) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		"INSERT INTO dem_tiles (filename, bounding_box, crs, resolution_m, rows, cols, file_path, source) VALUES ($1, ST_GeomFromText($2, 4326), $3, $4, $5, $6, $7, $8) RETURNING id",
		d.Filename, d.BoundingBox, d.Crs, d.ResolutionM, d.Rows, d.Cols, d.FilePath, d.Source).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert DEM tile: %w", err)
	}
	return id, nil
}

func InsertGeology(ctx context.Context, pool *pgxpool.Pool, g models.Geology) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		"INSERT INTO geology (dem_tile_id, geometry, bounce_coeff, friction_coeff, fragmentation_k, rock_type) VALUES ($1, ST_GeomFromText($2, 4326), $3, $4, $5, $6) RETURNING id",
		g.DemTileID, g.Geometry, g.BounceCoeff, g.FrictionCoeff, g.FragmentationK, g.RockType).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert geology: %w", err)
	}
	return id, nil
}

func InsertRoute(ctx context.Context, pool *pgxpool.Pool, r models.Route) (int, error) {
	var id int
	err := pool.QueryRow(ctx,
		"INSERT INTO routes (name, geometry, source) VALUES ($1, ST_GeomFromText($2, 4326), $3) RETURNING id",
		r.Name, r.Geometry, r.Source).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to insert route: %w", err)
	}
	return id, nil
}
