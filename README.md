# Thermistor LUT and Steinhart-Hart Generator

This tool generates ready-to-use C header files containing:

- **Steinhart-Hart function** - temperature calculation from raw ADC values using fitted coefficients.  
- **Lookup table (LUT)** - optional direct ADC-to-temperature mapping for fast runtime performance.  




The generator is designed specifically for **NTC thermistors**. It produces functions and tables that work **directly with raw ADC readings**, eliminating the need to manually convert ADC values to resistance in firmware.


The program takes a CSV file of thermistor data (resistance vs. temperature, from a datasheet or measurements) and produces C headers and CSV outputs.  

Supported network: standard voltage divider with series resistor (and optional parallel resistor). The ADC reference voltage must match the supply voltage used in the divider.  

**Always verify generated results against known values to ensure correctness.**


## Index

1. [Installation](#installation)
2. [Building](#building)
3. [Usage](#usage)
    - [Flags](#minimal-required-flag)
    - [Input CSV formatting](#input-csv)
    - [Examples](#examples)
4. [Performance](#performance)
5. [Licence](#licence)

## Installation

You can either download a prebuilt executable from the [releases page](https://github.com/Eriosies/thermistor-lut-gen/releases) or bulid it yourself:

## Building

Clone the repository and build the generator:

```
git clone https://github.com/Eriosies/thermistor-lut-gen.git
cd thermistor-lut-gen
go build -o thermistor-gen ./cmd/thermistor-gen
```

## Usage

Generate Steinhart-Hart coefficients from a thermistor CSV:


    thermistor-gen -i thermistor.csv

#### Minimal Required Flag

- `-i` : Path to the input CSV file containing temperature vs. resistance points.

#### Optional Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-o` | Output directory for generated files | `./output` |
| `-n` | Base name for generated files | from CSV metadata name field or "thermistor" |
| `-lut` | LUT size (power of 2, 0 = Steinhart only) | 0 |
| `-a` | ADC resolution in bits | 12 |
| `-v` | ADC reference voltage (V) | 3.3 |
| `-rs` | Series resistor (kΩ) | 10.0 |
| `-rp` | Parallel resistor (kΩ), 0 = none | 0.0 |
| `-tu` | Upper temperature limit (°C) | 125 |
| `-tl` | Lower temperature limit (°C) | -40 |
| `-fp` | Decimal points for fixed point int LUT | 0 |

#### Voltage Divider Schematic 

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
    GND


### Input CSV

See [thermistor csv examples](./examples/thermistor_tables/) for examples of csv formatting. Input `.csv` file is limited to 10MB.

#### CSV Structure

**Metadata**  
- All rows **before** the header row `Temperature, Resistance` are considered metadata.  
- First column = descriptor, second column = value.  
- Metadata is printed into the generated `.h` files for reference.  
- Maximum of 20 metadata entries.

**Temperature Data**  
- Begins at the row with the header `Temperature, Resistance`.  
- Columns expected in **°C** and **Ω**.  
- Minimum of 3 points required for Steinhart-Hart calculation.  

| Column 1 | Column 2 |
|----------|----------|
| Name | NCP18 |
| Model Number | NCP18XH |
| ...  | ... |
|Temperature | Resistance |
| -40 | 195652 | 
| -35 | 148171 |
| ... | ...|

### Examples

Example generations can be seen in [Examples Folder](./examples/outputs/)


#### Example generation

`thermistor-gen -i thermistor.csv -o ./output -lut 256 -n x`

This generates a LUT of size 256, with the name "x" using defaults:
- Vref = 3.3V
- ADC resolution = 12-bit
- Series resistor = 10 kΩ
- no parallel resistor

This produces:
- `x_steinhart.h` - header with calculated coefficients and function for getting temperature from raw ADC values
- `x_lut.h` - Header with float and int LUT tables mapped from raw ADC values
- `x_LUT.csv` and `x_Variance.csv` - reference data for verification

#### Example usage of header files
```c
#include "x_steinhart.h"
#include "x_lut.h"

uint16_t adcValue;
float temperatureLUT;
float temperatureSH;

// Replace with your ADC read function
adcValue = readADC();

// LUT (fast lookup, float or int version available)
temperatureLUT = x_lut_get_temp_float(adcValue);

// Steinhart-Hart calculation
temperatureSH = x_steinhart_get_temp(adcValue);
```



## Performance

Tested on an STM32F407xx (ARM Cortex-M4 core) using the DWT `CYCCNT` register to measure clock cycles required to compute a temperature value.

#### Test Conditions

- LUT size: 256  
- ADC resolution: 12-bit  
- Series resistor: 10 kΩ  
- Parallel resistor: none  

#### Results

Clock cycles to compute temperature from the corresponding function:

##### With FPU

| Optimization | -O0  | -O2  | -Os  |
|--------------|------|------|------|
| Steinhart-Hart | 306  | 256  | 262  |
| LUT           | 33   | 4    | 13   |

##### Without FPU

| Optimization | -O0   | -O2   | -Os   |
|--------------|-------|-------|-------|
| Steinhart-Hart | 2788  | 2745  | 2726  |
| LUT           | 30    | 4     | 14    |