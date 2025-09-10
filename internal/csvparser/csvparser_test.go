package csvparser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Eriosies/thermistor-lut-gen/internal/csvparser"
	"github.com/Eriosies/thermistor-lut-gen/models"
)

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "test.csv")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp CSV: %v", err)
	}
	return tmpFile
}

func TestParseThermistorCSV(t *testing.T) {
	csv := `Manufacturer,TestCorp
Location,Lab1
Temperature,Resistance
25,10000
50,5000
75,2500
`
	file := writeTempCSV(t, csv)

	metadata, points, err := csvparser.ParseThermistorCSV(file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(metadata) != 2 {
		t.Errorf("expected 2 metadata rows, got %d", len(metadata))
	}
	if len(points) != 3 {
		t.Errorf("expected 3 data points, got %d", len(points))
	}

	expected := []models.ThermistorPoint{
		{Temp: 25, Resistance: 10000},
		{Temp: 50, Resistance: 5000},
		{Temp: 75, Resistance: 2500},
	}
	for i, pt := range points {
		if pt != expected[i] {
			t.Errorf("point %d mismatch: got %+v, want %+v", i, pt, expected[i])
		}
	}
}

func TestParseThermistorCSV_Errors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"missing table", "Metadata1,Value1"},
		{"bad data", "Temperature,Resistance\n25,abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := writeTempCSV(t, tt.content)
			_, _, err := csvparser.ParseThermistorCSV(file)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}

	// nonexistent file
	_, _, err := csvparser.ParseThermistorCSV("nonexistent.csv")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestReadCSV_Warnings(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectErr     bool
		expectWarning bool
	}{
		{"too few points", "Temperature,Resistance\n25,10000\n50,5000", true, false},
		{"minimum points", "Temperature,Resistance\n25,10000\n50,5000\n75,2500", false, true},
		{"low point count warning", "Temperature,Resistance\n25,10000\n50,5000\n75,2500\n100,1250\n150,625", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := writeTempCSV(t, tt.content)
			points, metadata, warnings, err := csvparser.ReadCSV(file)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error=%v, got %v", tt.expectErr, err)
			}
			if (len(warnings) > 0) != tt.expectWarning {
				t.Errorf("expected warning=%v, got %v", tt.expectWarning, warnings)
			}
			_ = points
			_ = metadata
		})
	}
}
