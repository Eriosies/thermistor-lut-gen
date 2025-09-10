package models

const KelvinToCelsius float64 = 273.15
const ResistanceMax float64 = 1e9

type ThermistorPoint struct {
	Temp       float64
	Resistance float64
}

type Config struct {
	InputFile      string
	OutputDir      string
	BaseName       string
	LUTSize        uint
	ADCResolution  uint
	VoltageRef     float64
	RS             float64
	RP             float64
	UpperLimitTemp float64
	LowerLimitTemp float64
	NameFlag       string
	FixedPoint     uint
}

type DeviationTable struct {
	Resistance      float64
	TemperatureCSV  float64
	TemperatureCalc float64
	Deviation       float64
}

func DetermineBaseName(cfg Config, metadata [][2]string) string {
	if cfg.NameFlag != "" {
		return cfg.NameFlag
	}

	for _, m := range metadata {
		key, val := m[0], m[1]
		if val != "" && (key == "Name" || key == "Model" || key == "Thermistor") {
			return val
		}
	}
	return "thermistor"
}
