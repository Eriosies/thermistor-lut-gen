package thermistor

import (
	"fmt"

	"github.com/Eriosies/thermistor-lut-gen/models"
)

func getResistanceFromADCValue(vRef float64, adcValue uint, adcBits uint, rSeries float64, rParallel float64) float64 {
	var resistance float64
	var adcMax uint = (1 << adcBits) - 1

	if adcValue == 0 {
		return 0.0
	}
	if adcValue == adcMax {
		return models.ResistanceMax
	}

	vADC := vRef * float64(adcValue) / float64(adcMax+1)

	rTemp := rSeries * (vADC / (vRef - vADC))

	if rParallel == 0 {
		resistance = rTemp
	} else {
		resistance = 1 / ((1 / rTemp) - (1 / rParallel))
	}

	return resistance
}

func clampTemperature(temperature float64, tempUpperLimit float64, tempLowerLimit float64) float64 {
	ret := temperature
	if temperature > tempUpperLimit {
		ret = tempUpperLimit
	} else if temperature < tempLowerLimit {
		ret = tempLowerLimit
	}

	return ret
}

func GenerateLUT(cfg models.Config, coeff [3]float64) ([]float64, []float64, []uint, error) {
	adcValues := make([]uint, cfg.LUTSize)
	resistanceValues := make([]float64, cfg.LUTSize)
	tempValues := make([]float64, cfg.LUTSize)

	adcMax := uint((1 << cfg.ADCResolution) - 1)

	if cfg.LUTSize > uint(adcMax+1) {
		err := fmt.Errorf("error: LUT size cannot exceed ADC max.\nLUT = %d, ADC_MAX = %d", cfg.LUTSize, adcMax)
		return nil, nil, nil, err
	}

	stepSize := (adcMax + 1) / cfg.LUTSize

	for i := uint(1); i < cfg.LUTSize-1; i++ {
		adcValues[i] = i * stepSize
		resistanceValues[i] = getResistanceFromADCValue(cfg.VoltageRef, adcValues[i], cfg.ADCResolution, cfg.RS*1000, cfg.RP*1000)
		rawTemp := SteinhartCalculation(resistanceValues[i], coeff) - models.KelvinToCelsius
		tempValues[i] = clampTemperature(rawTemp, cfg.UpperLimitTemp, cfg.LowerLimitTemp)
	}

	tempValues[0] = cfg.UpperLimitTemp
	tempValues[cfg.LUTSize-1] = cfg.LowerLimitTemp

	return tempValues, resistanceValues, adcValues, nil
}
