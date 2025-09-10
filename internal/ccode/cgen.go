package ccode

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Eriosies/thermistor-lut-gen/internal/csvparser"
	"github.com/Eriosies/thermistor-lut-gen/models"
)

const resistanceMax float64 = 1e9
const resistanceMin float64 = 0.1
const arrayLinebreak int = 16

func trimToFileName(path string) string {
	fileName := filepath.Base(path)
	fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return fileNameNoExt
}

func printHeader(w *bufio.Writer, name string, metadata [][2]string, cfg models.Config) {
	now := time.Now()

	fmt.Fprintf(w, "/**\n")
	fmt.Fprintf(w, "\t******************************************************************************\n")
	fmt.Fprintf(w, "\t* @file %s.h\n", name)
	fmt.Fprintf(w, "\t* @date %d-%d-%d\n", now.Year(), now.Month(), now.Day())
	fmt.Fprintf(w, "\t* Generated using https://github.com/Eriosies/thermistor-lut-gen\n")
	fmt.Fprintf(w, "\t*\n")
	fmt.Fprintf(w, "\t******************************************************************************\n")
	fmt.Fprintf(w, "\t* Thermistor CSV metadata\n")
	for _, m := range metadata {
		fmt.Fprintf(w, "\t*\t%s - %s\n", m[0], m[1])
	}
	fmt.Fprintf(w, "\t*\n")
	fmt.Fprintf(w, "\t******************************************************************************\n")
	fmt.Fprintf(w, "\t* Configuration of generation\n")
	fmt.Fprintf(w, "\t*\tInput File - %s\n", trimToFileName(cfg.InputFile)+".csv")
	fmt.Fprintf(w, "\t*\tLUT Size - %d\n", cfg.LUTSize)
	fmt.Fprintf(w, "\t*\tADC Resolution - %d\n", cfg.ADCResolution)
	fmt.Fprintf(w, "\t*\tReference voltage - %.2f\n", cfg.VoltageRef)
	fmt.Fprintf(w, "\t*\tSeries Resistor - %.0f\n", cfg.RS*1000)
	fmt.Fprintf(w, "\t*\tParallel Resistor - %.0f\n", cfg.RP*1000)
	fmt.Fprintf(w, "\t*\tFixed Point - %ddp\n", cfg.FixedPoint)
	fmt.Fprintf(w, "\t*\tUpper temperature limit - %.1f\n", cfg.UpperLimitTemp)
	fmt.Fprintf(w, "\t*\tLower temperature limit - %.1f\n", cfg.LowerLimitTemp)
	fmt.Fprintf(w, "\t*\n")
	fmt.Fprintf(w, "\t******************************************************************************\n")

	fmt.Fprintf(w, "\t*/\n\n")

}

func GenerateSteinhartCcode(path string, coeff [3]float64, metadata [][2]string, cfg models.Config) error {
	var useParallel int
	var adcType string

	name := trimToFileName(path)
	nameUpper := strings.ToUpper(name)
	nameLower := strings.ToLower(name)

	switch {
	case cfg.ADCResolution > 16:
		adcType = "uint32_t"
	case cfg.ADCResolution > 8:
		adcType = "uint16_t"
	default:
		adcType = "uint8_t"
	}

	if cfg.RP != 0.0 {
		useParallel = 1
	}

	nameUseParallel := fmt.Sprintf("%s_USE_PARALLEL", nameUpper)
	nameCoeffA := fmt.Sprintf("%s_COEFF_A", nameUpper)
	nameCoeffB := fmt.Sprintf("%s_COEFF_B", nameUpper)
	nameCoeffC := fmt.Sprintf("%s_COEFF_C", nameUpper)
	nameVRef := fmt.Sprintf("%s_VREF", nameUpper)
	nameADCRes := fmt.Sprintf("%s_ADC_RESOLUTION", nameUpper)
	nameADCMax := fmt.Sprintf("%s_ADC_MAX", nameUpper)
	nameRSeries := fmt.Sprintf("%s_RSERIES", nameUpper)
	namePSeries := fmt.Sprintf("%s_PSERIES", nameUpper)
	nameRMax := fmt.Sprintf("%s_RMAX", nameUpper)
	nameRMin := fmt.Sprintf("%s_RMIN", nameUpper)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	printHeader(w, name, metadata, cfg)

	fmt.Fprintf(w, "#ifndef %s_H\n", nameUpper)
	fmt.Fprintf(w, "#define %s_H\n\n", nameUpper)
	fmt.Fprintf(w, "#include \"stdint.h\"\n#include \"math.h\"\n\n")

	fmt.Fprintf(w, "#define %s %d\n\n", nameUseParallel, useParallel)

	fmt.Fprintf(w, "#define %s %ef\n", nameCoeffA, coeff[0])
	fmt.Fprintf(w, "#define %s %ef\n", nameCoeffB, coeff[1])
	fmt.Fprintf(w, "#define %s %ef\n\n", nameCoeffC, coeff[2])

	fmt.Fprintf(w, "#define KELVIN_TO_CELSIUS 273.15f\n\n")

	fmt.Fprintf(w, "#define %s %ff\n\n", nameVRef, cfg.VoltageRef)

	fmt.Fprintf(w, "#define %s %d\n", nameADCRes, cfg.ADCResolution)
	fmt.Fprintf(w, "#define %s ((1 << %s) - 1)\n\n", nameADCMax, nameADCRes)

	fmt.Fprintf(w, "#define %s %ff\n", nameRSeries, cfg.RS*1000)
	if useParallel != 0 {
		fmt.Fprintf(w, "#define %s %ff\n", namePSeries, cfg.RP*1000)
	}

	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "#define %s %.3Ef\n", nameRMax, resistanceMax)
	fmt.Fprintf(w, "#define %s %.3Ef\n", nameRMin, resistanceMin)

	fmt.Fprintf(w, "\n__attribute__((always_inline)) static inline float %s_get_resistance(%s adcValue)\n", nameLower, adcType)
	fmt.Fprintf(w, "{\n\tfloat r, v;\n\n")
	fmt.Fprintf(w, "\tif(adcValue == 0)\n\t\treturn %s;\n", nameRMin)
	fmt.Fprintf(w, "\tif(adcValue == %s)\n\t\treturn %s;\n\n", nameADCMax, nameRMax)
	fmt.Fprintf(w, "\tv = %s * (float) adcValue / %s;\n", nameVRef, nameADCMax)
	fmt.Fprintf(w, "\tr = %s * v / (%s - v);\n\n", nameRSeries, nameVRef)

	fmt.Fprintf(w, "#if %s\n", nameUseParallel)
	fmt.Fprintf(w, "\tr = 1/((1/r)-(1/%s));\n", nameUseParallel)
	fmt.Fprintf(w, "#endif\n\n")

	fmt.Fprintf(w, "\treturn r;\n")
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "__attribute__((always_inline)) static inline float %s_get_temp(%s adcValue)\n", nameLower, adcType)
	fmt.Fprintf(w, "{\n")
	fmt.Fprintf(w, "\tfloat lnR = logf(%s_get_resistance(adcValue));\n", nameLower)
	fmt.Fprintf(w, "\treturn 1 / (%s + %s * lnR + %s * lnR * lnR * lnR) - KELVIN_TO_CELSIUS;\n", nameCoeffA, nameCoeffB, nameCoeffC)
	fmt.Fprintf(w, "}\n\n")

	fmt.Fprintf(w, "#endif")

	return w.Flush()
}

func GenerateLUTCcode(path string, lutTemp []float64, metadata [][2]string, cfg models.Config) error {
	if cfg.LUTSize == 0 {
		return fmt.Errorf("LUT size is 0; cannot generate LUT header")
	}

	name := trimToFileName(path)
	nameUpper := strings.ToUpper(name)

	lutSizeBits := uint(math.Log2(float64(cfg.LUTSize)))

	nameUseFloat := fmt.Sprintf("%s_USE_FLOAT", nameUpper)
	nameUseInt := fmt.Sprintf("%s_USE_INT", nameUpper)
	nameLUTSize := fmt.Sprintf("%s_SIZE", nameUpper)
	nameLUTSizeBits := fmt.Sprintf("%s_SIZE_BITS", nameUpper)
	nameADCRes := fmt.Sprintf("%s_ADC_RESOLUTION", nameUpper)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	printHeader(w, name, metadata, cfg)

	intTypeString := "int16_t"
	if lutTemp[cfg.LUTSize-1] < 127 && lutTemp[0] > -128 {
		intTypeString = "int8_t"
	}

	fmt.Fprintf(w, "#ifndef %s_H\n", nameUpper)
	fmt.Fprintf(w, "#define %s_H\n\n", nameUpper)
	fmt.Fprintf(w, "#include \"stdint.h\"\n\n")

	fmt.Fprintf(w, "#define %s 1\n", nameUseFloat)
	fmt.Fprintf(w, "#define %s 0\n\n", nameUseInt)

	fmt.Fprintf(w, "#define %s %dU\n", nameLUTSize, cfg.LUTSize)
	fmt.Fprintf(w, "#define %s %dU\n", nameLUTSizeBits, lutSizeBits)
	fmt.Fprintf(w, "#define %s %dU\n\n\n", nameADCRes, cfg.ADCResolution)

	fmt.Fprintf(w, "#if %s\n\n", nameUseFloat)
	fmt.Fprintf(w, "static const float %s_float[%s] = {", name, nameLUTSize)
	for i := 0; i < int(cfg.LUTSize)-1; i++ {
		if i%arrayLinebreak == 0 {
			fmt.Fprintf(w, "\n\t")
		}
		fmt.Fprintf(w, "%.2ff, ", lutTemp[i])
	}
	fmt.Fprintf(w, "%.2ff };\n\n", lutTemp[cfg.LUTSize-1])

	fmt.Fprintf(w, "__attribute__((always_inline)) static inline float %s_get_temp_float(uint32_t adcValue)\n", name)
	fmt.Fprintf(w, "{\n\tuint32_t index = adcValue >> (%s - %s);\n\treturn %s_float[index];\n}\n\n", nameADCRes, nameLUTSizeBits, name)

	fmt.Fprintf(w, "#endif\n\n")

	fmt.Fprintf(w, "#if %s\n\n", nameUseInt)
	fmt.Fprintf(w, "static const %s %s_int[%s] = {", intTypeString, name, nameLUTSize)
	for i := 0; i < int(cfg.LUTSize)-1; i++ {
		if i%arrayLinebreak == 0 {
			fmt.Fprintf(w, "\n\t")
		}
		fmt.Fprintf(w, "%d, ", int(lutTemp[i]*math.Pow(10, float64(cfg.FixedPoint))))
	}
	fmt.Fprintf(w, "%d };\n\n", int(lutTemp[cfg.LUTSize-1]))

	fmt.Fprintf(w, "__attribute__((always_inline)) static inline %s %s_get_temp_int(uint32_t adcValue)\n", intTypeString, name)
	fmt.Fprintf(w, "{\n\tuint32_t index = adcValue >> (%s - %s);\n\treturn %s_int[index];\n}\n\n", nameADCRes, nameLUTSizeBits, name)

	fmt.Fprintf(w, "#endif\n\n")
	fmt.Fprintf(w, "#endif")

	return w.Flush()
}

func GenerateOutputs(cfg models.Config, baseName string, coeff [3]float64,
	tempLUT, resistanceLUT []float64, adcLUT []uint,
	fullTable []models.DeviationTable, metadata [][2]string,
) (map[string]string, error) {

	files := make(map[string]string)

	lutCFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_lut.h", strings.ToLower(baseName)))
	steinhartCFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_steinhart.h", strings.ToLower(baseName)))
	files["steinhartC"] = steinhartCFile

	if cfg.LUTSize != 0 {
		lutCSV := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_LUT.csv", baseName))
		files["lutC"] = lutCFile
		files["lutCSV"] = lutCSV

		var lutRows [][]string
		for i := range tempLUT {
			lutRows = append(lutRows, []string{
				fmt.Sprintf("%.3f", resistanceLUT[i]),
				fmt.Sprintf("%.3f", tempLUT[i]),
				fmt.Sprintf("%d", adcLUT[i]),
			})
		}

		if err := csvparser.WriteCSV(lutCSV, "Resistance (Ω),Table Temp (°C),ADC Value", lutRows); err != nil {
			return files, err
		}

		if err := GenerateLUTCcode(lutCFile, tempLUT, metadata, cfg); err != nil {
			return files, err
		}
	}

	if err := GenerateSteinhartCcode(steinhartCFile, coeff, metadata, cfg); err != nil {
		return files, err
	}

	varianceCSV := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_Variance.csv", baseName))
	files["varianceCSV"] = varianceCSV

	var varianceRows [][]string
	for _, row := range fullTable {
		varianceRows = append(varianceRows, []string{
			fmt.Sprintf("%.3f", row.Resistance),
			fmt.Sprintf("%.3f", row.TemperatureCSV),
			fmt.Sprintf("%.3f", row.TemperatureCalc),
			fmt.Sprintf("%.3f", row.Deviation),
		})
	}
	if err := csvparser.WriteCSV(varianceCSV, "Resistance (Ω),Table Temp (°C),Fitted Temp (°C),Deviation (K)", varianceRows); err != nil {
		return files, err
	}

	return files, nil
}
