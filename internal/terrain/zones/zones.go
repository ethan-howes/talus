package zones

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/ethan-howes/talus/internal/store/models"
)

type cluster struct {
	cells []int
}

func readRaster(path string) ([]float32, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read raster %s: %w", path, err)
	}

	// convert raw bytes to float32 slice
	floats := make([]float32, len(data)/4)
	for i := range floats {
		bits := binary.LittleEndian.Uint32(data[i*4 : i*4+4])
		floats[i] = math.Float32frombits(bits)
	}
	return floats, nil
}

func findClusters(slope, plan, tri []float32, rows, cols int, slopeThresh, triThresh float64) []cluster {

	visited := make([]bool, rows*cols)
	var clusters []cluster

	for i := 0; i < rows*cols; i++ {
		if float64(slope[i]) > slopeThresh && plan[i] > 0 && float64(tri[i]) > triThresh && !visited[i] {
			var c cluster
			queue := []int{i}
			visited[i] = true

			for len(queue) > 0 {
				j := queue[0]
				queue = queue[1:]

				c.cells = append(c.cells, j)

				row := j / cols
				col := j % cols

				var neighbors []int

				// do not include boarders in cluster creation
				if row > 0 {
					neighbors = append(neighbors, (row-1)*cols+col)
				}
				if row < rows-1 {
					neighbors = append(neighbors, (row+1)*cols+col)
				}
				if col > 0 {
					neighbors = append(neighbors, row*cols+(col-1))
				}
				if col < cols-1 {
					neighbors = append(neighbors, row*cols+(col+1))
				}

				for _, n := range neighbors {
					if n >= 0 && n < rows*cols &&
						!visited[n] &&
						float64(slope[n]) > slopeThresh &&
						plan[n] > 0 &&
						float64(tri[n]) > triThresh {
						visited[n] = true
						queue = append(queue, n)
					}

				}

			}
			clusters = append(clusters, c)

		}
	}
	return clusters
}

func DetectSourceZones(slopePath, planPath, triPath string, rows, cols int, cellSizeM float64, slopeThresh, triThresh float64, originLon, originLat float64, demTileID, geologyID int) ([]models.SourceZone, error) {

	slope, err := readRaster(slopePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read slope: %w", err)
	}

	plan, err := readRaster(planPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan: %w", err)
	}

	tri, err := readRaster(triPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tri: %w", err)
	}

	clusters := findClusters(slope, plan, tri, rows, cols, slopeThresh, triThresh)

	var zones []models.SourceZone
	for _, c := range clusters {
		// small clusters are likely noise
		if len(c.cells) < 10 {
			continue
		}

		// bounding box computation
		minRow := rows
		minCol := cols
		maxRow := 0
		maxCol := 0
		var slopeSum float64

		for _, idx := range c.cells {
			r := idx / cols
			col := idx % cols
			if r < minRow {
				minRow = r
			}
			if r > maxRow {
				maxRow = r
			}
			if col < minCol {
				minCol = col
			}
			if col > maxCol {
				maxCol = col
			}
			slopeSum += float64(slope[idx])
		}
		meanSlope := slopeSum / float64(len(c.cells))
		areaMi2 := float64(len(c.cells)) * cellSizeM * cellSizeM

		// bbox to geographic location
		// 111320 is an approx of how many meters are in 1 degree of latitude or longitude
		minLon := originLon + float64(minCol)*cellSizeM/111320.0
		maxLon := originLon + float64(maxCol)*cellSizeM/111320.0
		maxLat := originLat - float64(minRow)*cellSizeM/111320.0
		minLat := originLat - float64(maxRow)*cellSizeM/111320.0

		centLon := (minLon + maxLon) / 2
		centLat := (minLat + maxLat) / 2

		polygon := fmt.Sprintf(
			"POLYGON((%f %f, %f %f, %f %f, %f %f, %f %f))",
			minLon, minLat,
			maxLon, minLat,
			maxLon, maxLat,
			minLon, maxLat,
			minLon, minLat,
		)

		centroid := fmt.Sprintf("POINT(%f %f)", centLon, centLat)

		zones = append(zones, models.SourceZone{
			DemTileID:     demTileID,
			Geometry:      polygon,
			Centroid:      centroid,
			MeanSlopeDeg:  meanSlope,
			MeanAspectDeg: 0,
			AreaM2:        areaMi2,
			GeologyID:     geologyID,
		})
	}
	return zones, nil
}
