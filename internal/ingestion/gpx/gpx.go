package gpx

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

type gpxFile struct {
	XMLName xml.Name   `xml:"gpx"`
	Tracks  []gpxTrack `xml:"trk"`
	Routes  []gpxRoute `xml:"rte"`
}

type gpxTrack struct {
	Segments []gpxSegment `xml:"trkseg"`
}

type gpxSegment struct {
	Points []gpxPoint `xml:"trkpt"`
}

type gpxRoute struct {
	Points []gpxPoint `xml:"rtept"`
}

type gpxPoint struct {
	Lat float64 `xml:"lat,attr"`
	Lon float64 `xml:"lon,attr"`
}

func Parse(filePath string) (string, error) {

	// file read
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read gpx file: %w", err)
	}

	// unmarshal to struct
	var gpx gpxFile
	err = xml.Unmarshal(data, &gpx)
	if err != nil {
		return "", fmt.Errorf("failed to parse gpx file: %w", err)
	}

	// collect all coordinate points
	var points []gpxPoint
	for _, track := range gpx.Tracks {
		for _, segment := range track.Segments {
			points = append(points, segment.Points...)
		}
	}

	for _, route := range gpx.Routes {
		points = append(points, route.Points...)
	}

	// creating postGIS linestring
	if len(points) < 2 {
		return "", fmt.Errorf("gpx file has fewer than 2 points")
	}

	parts := make([]string, len(points))
	for i, p := range points {
		parts[i] = fmt.Sprintf("%f %f", p.Lon, p.Lat)
	}

	linestring := "LINESTRING(" + strings.Join(parts, ", ") + ")"

	return linestring, nil
}
