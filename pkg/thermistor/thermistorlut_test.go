package thermistor

import (
	"testing"

	"github.com/Eriosies/thermistor-lut-gen/models"
)

type adcResistTable struct {
	adcVal     uint
	resistance float64
}

func TestGetResistanceFromADCValue(t *testing.T) {
	var temp float64
	cfg5vRp := models.Config{
		ADCResolution: 12,
		VoltageRef:    5.0,
		RS:            10000,
		RP:            20000,
	}

	cfg3v3Rs := models.Config{
		ADCResolution: 12,
		VoltageRef:    3.3,
		RS:            10000,
		RP:            0,
	}

	vrTableRp := []adcResistTable{
		{adcVal: 410, resistance: 1176}, //0.5V
		{adcVal: 819, resistance: 2857},
		{adcVal: 1229, resistance: 5455},
		{adcVal: 1638, resistance: 10000},
		{adcVal: 2048, resistance: 20001},
		{adcVal: 2457, resistance: 60004}, //3.0V
	}

	vrTableRs := []adcResistTable{
		{adcVal: 621, resistance: 1785.7}, //0.5V
		{adcVal: 1241, resistance: 4347.9},
		{adcVal: 1862, resistance: 8333.5},
		{adcVal: 2482, resistance: 15385.6},
		{adcVal: 3103, resistance: 31253.1},
		{adcVal: 3724, resistance: 100000}, //3.0V
	}

	for _, row := range vrTableRp {
		rTest := getResistanceFromADCValue(cfg5vRp.VoltageRef, row.adcVal, cfg5vRp.ADCResolution, cfg5vRp.RS, cfg5vRp.RP)
		if !floatAlmostEqualPercentage(rTest, row.resistance, 0.005) {
			t.Errorf("Expected %f, got %f", row.resistance, rTest)
		}
	}

	for _, row := range vrTableRs {
		rTest := getResistanceFromADCValue(cfg3v3Rs.VoltageRef, row.adcVal, cfg3v3Rs.ADCResolution, cfg3v3Rs.RS, cfg3v3Rs.RP)
		if !floatAlmostEqualPercentage(rTest, row.resistance, 0.005) {
			t.Errorf("Expected %f, got %f", row.resistance, rTest)
		}
	}
	temp = getResistanceFromADCValue(cfg3v3Rs.VoltageRef, 0, cfg3v3Rs.ADCResolution, cfg3v3Rs.RS, cfg3v3Rs.RP)
	if temp != 0 {
		t.Errorf("Input of ADC max, expected 0, actual = %.3g", temp)
	}
	temp = getResistanceFromADCValue(cfg3v3Rs.VoltageRef, 4095, cfg3v3Rs.ADCResolution, cfg3v3Rs.RS, cfg3v3Rs.RP)
	if temp < models.ResistanceMax {
		t.Errorf("Input of ADC val = 0, expected >= %.3g, actual = %.3g", models.ResistanceMax, temp)
	}

}

func TestClampTemperature(t *testing.T) {
	tests := []struct {
		temp, upper, lower, want float64
	}{
		{25, 30, 20, 25}, // inside range
		{35, 30, 20, 30}, // above upper limit
		{15, 30, 20, 20}, // below lower limit
		{30, 30, 20, 30}, // exactly at upper limit
		{20, 30, 20, 20}, // exactly at lower limit
	}

	for _, tt := range tests {
		got := clampTemperature(tt.temp, tt.upper, tt.lower)
		if got != tt.want {
			t.Errorf("clampTemperature(%f,%f,%f) = %f; want %f", tt.temp, tt.upper, tt.lower, got, tt.want)
		}
	}
}

func TestGenerateLUT(t *testing.T) {
	cfg := models.Config{
		LUTSize:        16,
		ADCResolution:  12,
		VoltageRef:     3.3,
		RS:             10,  // kΩ
		RP:             100, // kΩ
		UpperLimitTemp: 100.0,
		LowerLimitTemp: 0.0,
	}

	temps, resistances, adcs, err := GenerateLUT(cfg, testSteinhartCoeff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(temps) != int(cfg.LUTSize) {
		t.Errorf("temps length = %d, want %d", len(temps), cfg.LUTSize)
	}
	if len(resistances) != int(cfg.LUTSize) {
		t.Errorf("resistances length = %d, want %d", len(resistances), cfg.LUTSize)
	}
	if len(adcs) != int(cfg.LUTSize) {
		t.Errorf("adcs length = %d, want %d", len(adcs), cfg.LUTSize)
	}

	// Check first and last values are clamped correctly
	if temps[0] != cfg.UpperLimitTemp {
		t.Errorf("first temp = %f, want %f", temps[0], cfg.UpperLimitTemp)
	}
	if temps[len(temps)-1] != cfg.LowerLimitTemp {
		t.Errorf("last temp = %f, want %f", temps[len(temps)-1], cfg.LowerLimitTemp)
	}

	// Optional: check monotonicity or general correctness of values
	for i := 1; i < len(adcs)-1; i++ {
		if adcs[i] <= adcs[i-1] {
			t.Errorf("adc values not increasing at index %d: %d <= %d", i, adcs[i], adcs[i-1])
		}
	}
}

// Test error condition
func TestGenerateLUT_Error(t *testing.T) {
	cfg := models.Config{
		LUTSize:       5000, // exceeds 12-bit ADC max (4095)
		ADCResolution: 12,
		VoltageRef:    3.3,
		RS:            10,
		RP:            100,
	}
	testSteinhartCoeff := [3]float64{0.001, 0.0002, 0.000003}

	_, _, _, err := GenerateLUT(cfg, testSteinhartCoeff)
	if err == nil {
		t.Fatalf("expected error for LUT size exceeding ADC max, got nil")
	}
}
