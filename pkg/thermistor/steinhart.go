package thermistor

import (
	"math"

	"github.com/Eriosies/thermistor-lut-gen/models"
	"gonum.org/v1/gonum/mat"
)

const KelvinToCelsius float64 = 273.15

func leastSquares(X [][]float64, Y []float64) ([]float64, error) {

	rowsX := len(X)    //point count
	colsX := len(X[0]) //coeff count

	coeffs := make([]float64, colsX)

	flatX := make([]float64, 0, rowsX*colsX)
	for _, row := range X {
		flatX = append(flatX, row...)
	}

	Xmat := mat.NewDense(rowsX, colsX, flatX)

	Yvec := mat.NewVecDense(len(Y), Y)

	var qr mat.QR
	qr.Factorize(Xmat)

	var beta mat.VecDense
	err := qr.SolveVecTo(&beta, false, Yvec)

	if err != nil {
		return nil, err
	}

	for i := 0; i < colsX; i++ {
		coeffs[i] = beta.AtVec(i)
	}

	return coeffs, nil

}

func SteinhartCalculation(resistance float64, coeff [3]float64) float64 {

	lnR := math.Log(resistance)
	lnR3 := lnR * lnR * lnR
	return 1 / (coeff[0] + coeff[1]*lnR + coeff[2]*lnR3)
}

func FindSteinhartCoefficients(points []models.ThermistorPoint) ([3]float64, error) {
	n := len(points)
	X := make([]float64, n*3)
	Y := make([]float64, n)

	for i, p := range points {
		lnR := math.Log(p.Resistance)
		X[i*3+0] = 1.0
		X[i*3+1] = lnR
		X[i*3+2] = lnR * lnR * lnR

		Y[i] = 1.0 / (p.Temp + KelvinToCelsius)
	}

	X2D := make([][]float64, n)
	for i := range X2D {
		X2D[i] = X[i*3 : i*3+3]
	}

	coeffs, err := leastSquares(X2D, Y)

	var result [3]float64
	copy(result[:], coeffs)
	return result, err
}

func CheckDeviation(points []models.ThermistorPoint, coeff [3]float64) ([]models.DeviationTable, float64, float64) {
	var fullTable []models.DeviationTable
	var maxDev, avgDev float64 = 0, 0

	for _, p := range points {
		tTemp := SteinhartCalculation(p.Resistance, coeff) - models.KelvinToCelsius
		deviation := p.Temp - tTemp
		fullTable = append(fullTable, models.DeviationTable{	
			Resistance: p.Resistance,
			TemperatureCSV: p.Temp,
			TemperatureCalc: tTemp,
			Deviation: deviation})

		if math.Abs(deviation) > maxDev {
			maxDev = math.Abs(deviation)
		}
		avgDev += math.Abs(deviation)
	}

	avgDev /= float64(len(points))
	return fullTable, maxDev, avgDev
}
