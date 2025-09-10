package csvparser

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Eriosies/thermistor-lut-gen/models"
)

const maxCSVSize int64 = 10 * 1024 * 1024 //10mb
const maxMetadataSize int = 20

func ParseThermistorCSV(path string) (metadata [][2]string, points []models.ThermistorPoint, err error) {
	info, err := os.Stat(path)

	if err != nil {
		return nil, nil, err
	}

	if info.Size() > maxCSVSize {
		return nil, nil, fmt.Errorf("file too large: %d mb -- should be under 10mb", info.Size()/(1024*1024))
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()

	if err != nil {
		return nil, nil, err
	}

	inTable := false
	for i, row := range rows {
		if len(row) != 2 || (row[0] == "" && row[1] == "") {
			continue
		}

		if strings.EqualFold(strings.TrimSpace(row[0]), "Temperature") && strings.EqualFold(strings.TrimSpace(row[1]), "Resistance") {
			//Start of data
			inTable = true
			continue
		}

		if inTable {
			temp, err1 := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
			resistance, err2 := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)

			if err1 != nil || err2 != nil {
				return nil, nil, fmt.Errorf("failed to parse row %d: %v, %v, row content: %v", i+1, err1, err2, row)
			}
			points = append(points, models.ThermistorPoint{Temp: temp, Resistance: resistance})

		} else {
			metadata = append(metadata, [2]string{strings.TrimSpace(row[0]), strings.TrimSpace(row[1])})
		}

	}

	if !inTable || len(points) == 0 {
		return nil, nil, fmt.Errorf("failed to find Temperature/Resistance data in %s", path)
	}

	return metadata, points, nil
}

func ReadCSV(file string) ([]models.ThermistorPoint, [][2]string, []string, error) {
	metadata, points, err := ParseThermistorCSV(file)
	if err != nil {
		return nil, nil, nil, err
	}

	var warnings []string
	if len(metadata) > maxMetadataSize {
		warnings = append(warnings, fmt.Sprintf("Too much metadata (%d entries, max %d). Check CSV file.", len(metadata), maxMetadataSize))
	}

	if len(points) < 3 {
		return nil, nil, warnings, fmt.Errorf("not enough points to calculate Steinhart coefficients. Minimum = 3")
	} else if len(points) < 6 {
		warnings = append(warnings, fmt.Sprintf("Low point count (%d). Results may be unstable.", len(points)))
	} else if len(points) > 1000 {
		warnings = append(warnings, fmt.Sprintf("Large number of points (%d) - calculation may take a while.", len(points)))
	}

	return points, metadata, warnings, nil
}

func WriteCSV(filePath string, header string, rows [][]string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, header)
	for _, row := range rows {
		fmt.Fprintln(f, strings.Join(row, ","))
	}
	return nil
}
