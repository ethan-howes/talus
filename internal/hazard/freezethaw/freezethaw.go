package freezethaw

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type FreezeThawPrediction struct {
	OvernightLowC    float64
	SunExposureHour  float64
	FreezeThawActive bool
	RiskLevel        string
}

type openMeteoResponse struct {
	Elevation float64         `json:"elevation"`
	Hourly    openMeteoHourly `json:"hourly"`
}

type openMeteoHourly struct {
	Time          []string  `json:"time"`
	Temperature2m []float64 `json:"temperature_2m"`
}

func ComputeSunExposureTime(aspectDeg float64) float64 {
	sunHour := 12.0 - (aspectDeg-180.0)/15.0

	// clamp to 24 hrs
	for sunHour < 0 {
		sunHour += 24
	}
	for sunHour >= 24 {
		sunHour -= 24
	}
	return sunHour
}

// PredictFreezeThaw fetches the overnight temperature forecast
// for a location and determines if a freeze-thaw cycle is active
func PredictFreezeThaw(ctx context.Context, minLon, maxLon, minLat, maxLat float64, elevationM float64, aspectDeg float64) (*FreezeThawPrediction, error) {

	lat := (minLat + maxLat) / 2
	lon := (minLon + maxLon) / 2

	// collecting data from external api
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m&forecast_days=2&timezone=auto", lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get forecast: %w", err)
	}
	defer resp.Body.Close()

	var forecast openMeteoResponse
	err = json.NewDecoder(resp.Body).Decode(&forecast)
	if err != nil {
		return nil, fmt.Errorf("failed to parse forecast: %w", err)
	}

	overnightLow := forecast.Hourly.Temperature2m[22]
	for i := 22; i < 30 && i < len(forecast.Hourly.Temperature2m); i++ {
		if forecast.Hourly.Temperature2m[i] < overnightLow {
			overnightLow = forecast.Hourly.Temperature2m[i]
		}
	}

	tempAtElevation := overnightLow - (elevationM * 6.5 / 1000.0)

	sunHour := ComputeSunExposureTime(aspectDeg)
	freezeThawActive := tempAtElevation < 0

	var riskLevel string
	switch {
	case tempAtElevation < -10:
		riskLevel = "extreme"
	case tempAtElevation < -5:
		riskLevel = "high"
	case tempAtElevation < 0:
		riskLevel = "moderate"
	default:
		riskLevel = "low"
	}

	return &FreezeThawPrediction{
		OvernightLowC:    tempAtElevation,
		SunExposureHour:  sunHour,
		FreezeThawActive: freezeThawActive,
		RiskLevel:        riskLevel,
	}, nil
}
