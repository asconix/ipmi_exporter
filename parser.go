package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type splittedOutput struct {
	value [][]string
}

func convertOutput(result [][]string, replacer *strings.Replacer) (metrics []metric, err error) {
	for _, first := range result {
		var value float64
		var currentMetric metric

		for n := range first {
			first[n] = strings.TrimSpace(first[n])
		}
		first[0] = strings.ToLower(first[0])
		first[0] = replacer.Replace(first[0])

		value, err = convertValue(first[1], first[2])
		if err != nil {
			log.Printf("could not parse ipmi output: %s", err)
		}

		currentMetric.value = value
		currentMetric.unit = first[2]
		currentMetric.metricsname = first[0]

		metrics = append(metrics, currentMetric)
	}
	return metrics, err
}

func splitAoutput(output string) ([][]string, error) {
	r := csv.NewReader(strings.NewReader(output))
	r.Comma = '|'
	r.Comment = '#'
	result, err := r.ReadAll()
	if err != nil {
		log.Printf("could not parse ipmi output: %v", err)
	}
	return result, err
}

// convertValue converts a parsed value to float64 and returns it.
// Values could be in the format:
// - 0x1 (hex) if the unit is "discrete"
// - 1.527 (float) for every other unit
// - na if the sensor is not available
// returns the converted float64 Value
func convertValue(value string, unit string) (retvalue float64, err error) {
	if value != "na" {
		if unit == "discrete" {
			if strings.HasPrefix(value, "0x") {
				value = value[2:]
			}
			parsedValue, err := strconv.ParseUint(value, 16, 32)
			if err != nil {
				return 0.0, fmt.Errorf("could not translate hex: %v, %v", value, err)
			}
			retvalue = float64(parsedValue)
		} else {
			retvalue, err = strconv.ParseFloat(value, 64)
		}
	}
	return retvalue, err
}
