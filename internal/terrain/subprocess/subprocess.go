package subprocess

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// runs and collects outputs from the cuda scripts

// ./terrain input.bin slope.bin aspect.bin plan.bin profile.bin tri.bin 10812 10812 92.59

type TerrainResult struct {
	SlopePath            string
	AspectPath           string
	PlanPath             string
	ProfilePath          string
	TriPath              string
	GpuTimeMsSlopeAspect float64
	CpuTimeMsSlopeAspect float64
	GpuTimeMsCurvature   float64
	CpuTimeMsCurvature   float64
	GpuTimeMsTri         float64
	CpuTimeMsTri         float64
}

func RunTerrain(binaryPath string, inputPath string, outputDir string, rows int, cols int, cellSize float64) (*TerrainResult, error) {

	slopePath := filepath.Join(outputDir, "slope.bin")
	aspectPath := filepath.Join(outputDir, "aspect.bin")
	planPath := filepath.Join(outputDir, "plan.bin")
	profilePath := filepath.Join(outputDir, "profile.bin")
	triPath := filepath.Join(outputDir, "tri.bin")

	cmd := exec.Command(binaryPath, inputPath, slopePath, aspectPath, planPath, profilePath, triPath, strconv.Itoa(rows), strconv.Itoa(cols), strconv.FormatFloat(cellSize, 'f', 6, 64))

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to exec command: %w", err)
	}

	text := string(out)
	lines := strings.Split(text, "\n")

	result := &TerrainResult{
		SlopePath:   slopePath,
		AspectPath:  aspectPath,
		PlanPath:    planPath,
		ProfilePath: profilePath,
		TriPath:     triPath,
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}

		switch key {
		case "gpu_time_ms_slope_aspect":
			result.GpuTimeMsSlopeAspect = value
		case "cpu_time_ms_slope_aspect":
			result.CpuTimeMsSlopeAspect = value
		case "gpu_time_ms_curvature":
			result.GpuTimeMsCurvature = value
		case "cpu_time_ms_curvature":
			result.CpuTimeMsCurvature = value
		case "gpu_time_ms_tri":
			result.GpuTimeMsTri = value
		case "cpu_time_ms_tri":
			result.CpuTimeMsTri = value
		}
	}
	return result, nil
}
