// Rock coefficients are estimates based on the following:
// Dorren, L.K.A. (2016). Rockyfor3D user manual, ecorisQ.
// Guzzetti et al. (2002). STONE: a computer program for 3D simulation of rock-falls.

package geology

import (
	"strings"
)

type RockParams struct {
	BounceCoeff    float64
	FrictionCoeff  float64
	FragmentationK float64
}

// Value sources are estimated from the articles cited above
var rockParams = map[string]RockParams{
	"granite":   {BounceCoeff: 0.35, FrictionCoeff: 0.62, FragmentationK: 2.1},
	"limestone": {BounceCoeff: 0.30, FrictionCoeff: 0.58, FragmentationK: 1.8},
	"sandstone": {BounceCoeff: 0.25, FrictionCoeff: 0.55, FragmentationK: 1.5},
	"shale":     {BounceCoeff: 0.20, FrictionCoeff: 0.50, FragmentationK: 1.3},
	"quartzite": {BounceCoeff: 0.38, FrictionCoeff: 0.65, FragmentationK: 2.3},
}

func Lookup(rockType string) (RockParams, bool) {
	params, ok := rockParams[strings.ToLower(rockType)]
	return params, ok
}
