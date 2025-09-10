package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Eriosies/thermistor-lut-gen/internal/ccode"
	"github.com/Eriosies/thermistor-lut-gen/internal/csvparser"
	"github.com/Eriosies/thermistor-lut-gen/models"
	"github.com/Eriosies/thermistor-lut-gen/pkg/thermistor"
)

func parseFlags() models.Config {
	cfg := models.Config{}

	flag.StringVar(&cfg.InputFile, "i", "", "Input CSV file path")
	flag.StringVar(&cfg.OutputDir, "o", "./output", "Output directory")
	flag.StringVar(&cfg.NameFlag, "n", "", "Base name for generated files (optional)")
	flag.UintVar(&cfg.LUTSize, "lut", 0, "LUT size (power of 2) or 0 for Steinhart.h only (default 0)")
	flag.UintVar(&cfg.ADCResolution, "a", 12, "ADC resolution in bits")
	flag.Float64Var(&cfg.VoltageRef, "v", 3.3, "ADC reference voltage (V)")
	flag.Float64Var(&cfg.RS, "rs", 10.0, "Series resistance (kΩ)")
	flag.Float64Var(&cfg.RP, "rp", 0.0, "Parallel resistance (kΩ), 0 = none (default 0)")
	flag.Float64Var(&cfg.UpperLimitTemp, "tu", 125.0, "Upper temperature limit (°C)")
	flag.Float64Var(&cfg.LowerLimitTemp, "tl", -40.0, "Lower temperature limit (°C)")
	flag.UintVar(&cfg.FixedPoint, "fp", 0, "Adjust the position of the fixed point in the uint LUT table (default 0)")

	flag.Usage = func() {
		fmt.Println("Thermistor LUT Generator")
		fmt.Println("Usage: thermistor-gen -i input.csv -o output_dir [-n base_name] [options]")
		fmt.Println("\nVoltage divider schematic (rs in series, thermistor with optional parallel rp):")
		fmt.Println(`
	Vref
	|
	|
	[Series Resistor]
	|
	|
	+-------------------+------ Vout
	|                   |
	[Thermistor]        [Parallel Resistor] (optional)
	|                   |
	+-------------------+
	| 
	GND`)
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.InputFile == "" {
		log.Fatal("Input file is required. Use -help for more information.")
	}

	info, err := os.Stat(cfg.OutputDir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
				log.Fatal("Failed to create output directory:", err)
			}
		} else {
			log.Fatal(err)
		}
	} else if !info.IsDir() {
		log.Fatalf("Output path exists but is not a directory: %s", cfg.OutputDir)
	}

	if cfg.LUTSize != 0 && (cfg.LUTSize&(cfg.LUTSize-1)) != 0 {
		log.Fatal("LUT size must be a power of 2, or 0 for no LUT generation. e.g. 256, 512, 1024...")
	}

	if cfg.LUTSize > (1 << cfg.ADCResolution) {
		log.Fatalf("LUT size (%d) cannot exceed ADC maximum (%d).", cfg.LUTSize, 1<<cfg.ADCResolution)
	}

	return cfg
}

func main() {
	var err error
	var tempLUT, resistanceLUT []float64
	var adcLUT []uint

	cfg := parseFlags()

	fmt.Println("\nThermistor C Code LUT Generator")
	fmt.Println("------------------------------------")
	fmt.Printf(
		"Input File: %s\nOutput Directory: %s\nLUT Size: %d\nADC Resolution: %d bit\nADC Reference Voltage: %.2fV\nResistor Series: %.2fk\nResistor Parallel: %.2fk\n",
		cfg.InputFile, cfg.OutputDir, cfg.LUTSize, cfg.ADCResolution, cfg.VoltageRef, cfg.RS, cfg.RP,
	)

	points, metadata, warnings, err := csvparser.ReadCSV(cfg.InputFile)
	if err != nil {
		log.Fatal(err)
	}

	for _, w := range warnings {
		log.Printf("Warning: %s", w)
	}

	fmt.Println("\nCSV Metadata:")
	for _, m := range metadata {
		fmt.Printf("%s - %s\n", m[0], m[1])
	}
	fmt.Printf("\nNumber of points = %d\n\n", len(points))

	coeff, err := thermistor.FindSteinhartCoefficients(points)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Steinhart-Hart coefficients:\na = %.3g\nb = %.3g\nc = %.3g\n\n", coeff[0], coeff[1], coeff[2])

	fullTable, maxDev, avgDev := thermistor.CheckDeviation(points, coeff)

	fmt.Printf("Steinhart-Hart deviation from csv\n")
	fmt.Printf("Max Deviation: %.3g K, Avg Deviation: %.3g K\n", maxDev, avgDev)

	baseName := models.DetermineBaseName(cfg, metadata)

	if cfg.LUTSize != 0 {
		tempLUT, resistanceLUT, adcLUT, err = thermistor.GenerateLUT(cfg, coeff)
	}

	if err != nil {
		log.Fatal(err)
	}

	files, err := ccode.GenerateOutputs(cfg, baseName, coeff, tempLUT, resistanceLUT, adcLUT, fullTable, metadata)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nC Headers and CSV files:")
	for key, path := range files {
		fmt.Printf("  %s: %s\n", key, path)
	}
	fmt.Println()


}
