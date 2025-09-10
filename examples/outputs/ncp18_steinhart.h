/**
	******************************************************************************
	* @file ncp18_steinhart.h
	* @date 2025-9-11
	* Generated using https://github.com/Eriosies/thermistor-lut-gen
	*
	******************************************************************************
	* Thermistor CSV metadata
	*	Name - NCP18XH
	*	Part_Number - NCP18XH103F03RB
	*	Type - NTC
	*	Manufacturer - Murata
	*	Nominal_Resistance - 10000
	*	B_Constant - 3380
	*	Temperature Format - Celsius
	*
	******************************************************************************
	* Configuration of generation
	*	Input File - Murata_NCP18XH103F03RB_10k_3380k_0603.csv
	*	LUT Size - 256
	*	ADC Resolution - 12
	*	Reference voltage - 3.30
	*	Series Resistor - 10000
	*	Parallel Resistor - 0
	*	Fixed Point - 0dp
	*	Upper temperature limit - 125.0
	*	Lower temperature limit - -40.0
	*
	******************************************************************************
	*/

#ifndef NCP18_STEINHART_H
#define NCP18_STEINHART_H

#include "stdint.h"
#include "math.h"

#define NCP18_STEINHART_USE_PARALLEL 0

#define NCP18_STEINHART_COEFF_A 8.574782e-04f
#define NCP18_STEINHART_COEFF_B 2.568106e-04f
#define NCP18_STEINHART_COEFF_C 1.688598e-07f

#define KELVIN_TO_CELSIUS 273.15f

#define NCP18_STEINHART_VREF 3.300000f

#define NCP18_STEINHART_ADC_RESOLUTION 12
#define NCP18_STEINHART_ADC_MAX ((1 << NCP18_STEINHART_ADC_RESOLUTION) - 1)

#define NCP18_STEINHART_RSERIES 10000.000000f

#define NCP18_STEINHART_RMAX 1.000E+09f
#define NCP18_STEINHART_RMIN 1.000E-01f

__attribute__((always_inline)) static inline float ncp18_steinhart_get_resistance(uint16_t adcValue)
{
	float r, v;

	if(adcValue == 0)
		return NCP18_STEINHART_RMIN;
	if(adcValue == NCP18_STEINHART_ADC_MAX)
		return NCP18_STEINHART_RMAX;

	v = NCP18_STEINHART_VREF * (float) adcValue / NCP18_STEINHART_ADC_MAX;
	r = NCP18_STEINHART_RSERIES * v / (NCP18_STEINHART_VREF - v);

#if NCP18_STEINHART_USE_PARALLEL
	r = 1/((1/r)-(1/NCP18_STEINHART_USE_PARALLEL));
#endif

	return r;
}

__attribute__((always_inline)) static inline float ncp18_steinhart_get_temp(uint16_t adcValue)
{
	float lnR = logf(ncp18_steinhart_get_resistance(adcValue));
	return 1 / (NCP18_STEINHART_COEFF_A + NCP18_STEINHART_COEFF_B * lnR + NCP18_STEINHART_COEFF_C * lnR * lnR * lnR) - KELVIN_TO_CELSIUS;
}

#endif