package thermistor

import (
	"testing"
)

func TestLeastSquares(t *testing.T) {
	testX := [][]float64{
		{1, 1, 1},
		{1, 2, 4},
		{1, 3, 9},
	}
	testY := []float64{9, 18, 31}

	expected := []float64{4, 3, 2}

	result, err := leastSquares(testX, testY)
	if err != nil {
		t.Fatalf("leastSquares returned error: %v", err)
	}

	for i := range expected {
		if !floatAlmostEqual(result[i], expected[i], 0.01) {
			t.Errorf("result[%d] = %f; want %f", i, result[i], expected[i])
		}
	}
}

func TestSteinhartCalculation(t *testing.T) {

	for _, row := range testPoints {
		result := SteinhartCalculation(row.Resistance, testSteinhartCoeff) - KelvinToCelsius
		if !floatAlmostEqual(result, row.Temp, 0.1) {
			t.Errorf("result = %f; want %f", result, row.Temp)
		}
	}

}

func TestFindSteinhartCoefficients(t *testing.T) {

	results, err := FindSteinhartCoefficients(testPoints)

	if err != nil {
		t.Fatalf("FindSteinhartcoefficients returned error: %v", err)
	}

	for i := 0; i < len(testSteinhartCoeff); i++ {
		if !floatAlmostEqualPercentage(results[i], testSteinhartCoeff[i], 0.01) {
			t.Errorf("result = %f; want %f", results[i], testSteinhartCoeff[i])
		}
	}

}

func TestCheckDeviation(t *testing.T) {
	var expectedMaxDev float64 = 0.0
	var expectedAvgDev float64 = 0.0

	fullTable, maxDev, avgDev := CheckDeviation(testPoints, testSteinhartCoeff)

	if !floatAlmostEqual(expectedAvgDev, avgDev, 0.01) {
		t.Errorf("Average deviation = %.3g, expected = %.3g", expectedAvgDev, avgDev)
	}

	if !floatAlmostEqual(expectedMaxDev, maxDev, 0.01) {
		t.Errorf("Max deviation = %.3g, expected = %.3g", expectedMaxDev, maxDev)
	}

	for _, row := range fullTable {
		if !floatAlmostEqual(row.Deviation, 0, 0.01) {
			t.Errorf("Table deviation = %.3g, expected = %.3g", row.Deviation, 0.0)
		}
	}

}
