package ccode_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Eriosies/thermistor-lut-gen/internal/ccode"
	"github.com/Eriosies/thermistor-lut-gen/models"
)

func TestGenerateSteinhartCcode(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_steinhart.h")

	metadata := [][2]string{{"Manufacturer", "TestCorp"}}
	cfg := models.Config{
		InputFile:      "test.csv",
		ADCResolution:  12,
		RS:             10,
		RP:             5,
		VoltageRef:     3.3,
		FixedPoint:     2,
		UpperLimitTemp: 100,
		LowerLimitTemp: 0,
	}
	coeff := [3]float64{0.001, 0.0001, 0.00001}

	err := ccode.GenerateSteinhartCcode(filePath, coeff, metadata, cfg)
	if err != nil {
		t.Fatalf("GenerateSteinhartCcode returned error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "COEFF_A") || !strings.Contains(content, "COEFF_B") {
		t.Errorf("generated file missing coefficients")
	}

	if !strings.Contains(content, "TEST_STEINHART_H") {
		t.Errorf("generated header guard not found")
	}
}

func TestGenerateLUTCcode(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_lut.h")

	metadata := [][2]string{{"Manufacturer", "TestCorp"}}
	cfg := models.Config{
		LUTSize:       4,
		InputFile:     "test.csv",
		ADCResolution: 10,
		FixedPoint:    2,
	}
	lutTemp := []float64{0, 25, 50, 75}

	err := ccode.GenerateLUTCcode(filePath, lutTemp, metadata, cfg)
	if err != nil {
		t.Fatalf("GenerateLUTCcode returned error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "static const float test_lut_float") {
		t.Errorf("generated LUT float array not found")
	}
	if !strings.Contains(content, "#define TEST_LUT_SIZE") {
		t.Errorf("LUT size define not found")
	}

}

func TestGenerateOutputs(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := models.Config{
		OutputDir:      tmpDir,
		LUTSize:        2,
		InputFile:      "test.csv",
		ADCResolution:  12,
		RS:             10,
		RP:             5,
		VoltageRef:     3.3,
		FixedPoint:     2,
		UpperLimitTemp: 100,
		LowerLimitTemp: 0,
	}
	baseName := "test"
	coeff := [3]float64{0.001, 0.0001, 0.00001}
	tempLUT := []float64{0, 50}
	resistanceLUT := []float64{10000, 5000}
	adcLUT := []uint{0, 4095}
	fullTable := []models.DeviationTable{
		{Resistance: 10000, TemperatureCSV: 0, TemperatureCalc: 0, Deviation: 0},
	}
	metadata := [][2]string{{"Manufacturer", "TestCorp"}}

	files, err := ccode.GenerateOutputs(cfg, baseName, coeff, tempLUT, resistanceLUT, adcLUT, fullTable, metadata)
	if err != nil {
		t.Fatalf("GenerateOutputs returned error: %v", err)
	}

	for _, path := range files {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected file %s to exist, got error: %v", path, err)
		}
	}
}

func extractFloatArray(content, arrayName string) ([]float64, error) {
	// Find the start of the array.
	start := strings.Index(content, "static const float "+arrayName+"_float[")
	if start == -1 {
		return nil, nil
	}
	braceStart := strings.Index(content[start:], "{")
	braceEnd := strings.Index(content[start:], "};")
	if braceStart == -1 || braceEnd == -1 {
		return nil, nil
	}
	arrayStr := content[start+braceStart+1 : start+braceEnd]
	arrayStr = strings.ReplaceAll(arrayStr, "\n", "")
	arrayStr = strings.ReplaceAll(arrayStr, "\t", "")
	elements := strings.Split(arrayStr, ",")
	var result []float64
	for _, e := range elements {
		e = strings.TrimSpace(e)
		e = strings.TrimSpace(strings.TrimSuffix(e, "f"))
		if e == "" {
			continue
		}
		v, err := strconv.ParseFloat(e, 64)
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func TestLUTValues(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_lut.h")

	lutTemp := []float64{0, 25, 50, 75}
	cfg := models.Config{
		LUTSize:       uint(len(lutTemp)),
		InputFile:     "test.csv",
		ADCResolution: 10,
		FixedPoint:    2,
	}

	err := ccode.GenerateLUTCcode(filePath, lutTemp, [][2]string{}, cfg)
	if err != nil {
		t.Fatalf("GenerateLUTCcode returned error: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read generated file: %v", err)
	}

	content := string(data)
	array, err := extractFloatArray(content, "test_lut")
	if err != nil {
		t.Fatalf("failed to parse LUT array: %v", err)
	}

	if len(array) != len(lutTemp) {
		t.Fatalf("expected %d LUT values, got %d", len(lutTemp), len(array))
	}

	for i, v := range lutTemp {
		if array[i] != v {
			t.Errorf("value mismatch at index %d: got %f, want %f", i, array[i], v)
		}
	}
}
