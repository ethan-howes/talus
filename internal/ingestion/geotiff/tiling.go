package geotiff

import (
	"math"
)

type Tile struct {
	Row          int
	Col          int
	OriginRow    int
	OriginCol    int
	Rows         int
	Cols         int
	InteriorRows int
	InteriorCols int
	HaloCells    int
	OriginLon    float64
	OriginLat    float64
}

func ComputeTiles(info *GeoTIFFInfo, maxCellsPerSide int, haloMeters float64) []Tile {

	// width of halo
	haloCells := int(math.Ceil(haloMeters / info.ResolutionM))

	// tiles needed for each direction
	nTileRows := int(math.Ceil(float64(info.Rows) / float64(maxCellsPerSide)))
	nTileCols := int(math.Ceil(float64(info.Cols) / float64(maxCellsPerSide)))

	var tiles []Tile
	for tileRow := 0; tileRow < nTileRows; tileRow++ {
		for tileCol := 0; tileCol < nTileCols; tileCol++ {

			// interior bounds (without halo)
			startRow := tileRow * maxCellsPerSide
			startCol := tileCol * maxCellsPerSide
			endRow := min(startRow+maxCellsPerSide, info.Rows)
			endCol := min(startCol+maxCellsPerSide, info.Cols)

			// halo bounds (clamp to DEM edges)
			haloStartRow := max(startRow-haloCells, 0)
			haloStartCol := max(startCol-haloCells, 0)
			haloEndRow := min(endRow+haloCells, info.Rows)
			haloEndCol := min(endCol+haloCells, info.Cols)

			originLon := info.MinLon + float64(startCol)*info.ResolutionM/111320.0
			originLat := info.MaxLat - float64(startRow)*info.ResolutionM/111320.0

			tiles = append(tiles, Tile{
				Row:          tileRow,
				Col:          tileCol,
				OriginRow:    haloStartRow,
				OriginCol:    haloStartCol,
				Rows:         haloEndRow - haloStartRow,
				Cols:         haloEndCol - haloStartCol,
				InteriorRows: endRow - startRow,
				InteriorCols: endCol - startCol,
				HaloCells:    haloCells,
				OriginLon:    originLon,
				OriginLat:    originLat,
			})
		}
	}
	return tiles
}
