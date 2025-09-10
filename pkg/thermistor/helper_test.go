package thermistor

import (
	"math"

	"github.com/Eriosies/thermistor-lut-gen/models"
)


// testPoints & testSteinhartCoeff & coeffs calculated using https://www.thinksrs.com/downloads/programs/therm%20calc/ntccalibrator/ntccalculator.html
var testPoints = []models.ThermistorPoint{
	{Temp: 0, Resistance: 27219},
	{Temp: 25, Resistance: 10000},
	{Temp: 50, Resistance: 4161},
}
var testSteinhartCoeff = [3]float64{
	9.032678970e-4,
	2.487719619e-4,
	2.041094451e-7,
}

// Compare floats with tolerance
func floatAlmostEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

// Compare floats with tolerance as a percentage (tol = 0.05 -> 5%)
func floatAlmostEqualPercentage(a, b, tol float64) bool {
	if a == 0 && b == 0 {
		return true
	}
	if b == 0 {
		return false
	}

	relDiff := math.Abs(a-b) / math.Max(math.Abs(a), math.Abs(b))

	return relDiff <= tol

}
