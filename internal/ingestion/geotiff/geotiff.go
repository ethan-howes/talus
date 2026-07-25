package geotiff

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type GeoTIFFInfo struct {
	Filename    string
	CRS         string
	ResolutionM float64
	Rows        int
	Cols        int
	NoDataValue float64
	MinLon      float64
	MaxLon      float64
	MinLat      float64
	MaxLat      float64
}

type gdalInfo struct {
	Size              []int       `json:"size"`
	GeoTransform      []float64   `json:"geoTransform"`
	CornerCoordinates gdalCorners `json:"cornerCoordinates"`
	Bands             []gdalBand  `json:"bands"`
	Stac              gdalStac    `json:"stac"`
}

type gdalCorners struct {
	UpperLeft  []float64 `json:"upperLeft"`
	LowerLeft  []float64 `json:"lowerLeft"`
	UpperRight []float64 `json:"upperRight"`
	LowerRight []float64 `json:"lowerRight"`
}

type gdalBand struct {
	NoDataValue float64 `json:"noDataValue"`
}

type gdalStac struct {
	EPSG int `json:"proj:epsg"`
}

func Inspect(filePath string) (*GeoTIFFInfo, error) {

	// pull all metadata from file
	out, err := exec.Command("gdalinfo", "-json", filePath).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to pull filename metadata: %w", err)
	}

	var info gdalInfo
	err = json.Unmarshal(out, &info)
	if err != nil {
		return nil, fmt.Errorf("failed to parse gdalinfo output: %w", err)
	}

	return &GeoTIFFInfo{
		Filename:    filePath,
		CRS:         fmt.Sprintf("EPSG:%d", info.Stac.EPSG),
		ResolutionM: info.GeoTransform[1] * 111320.0,
		Rows:        info.Size[1],
		Cols:        info.Size[0],
		NoDataValue: info.Bands[0].NoDataValue,
		MinLon:      info.CornerCoordinates.LowerLeft[0],
		MaxLon:      info.CornerCoordinates.UpperRight[0],
		MinLat:      info.CornerCoordinates.LowerLeft[1],
		MaxLat:      info.CornerCoordinates.UpperLeft[1],
	}, nil

}
