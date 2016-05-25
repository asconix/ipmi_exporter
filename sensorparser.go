package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

type metric struct {
	metricsname string
	value       float64
	unit        string
	addr        string
}

func convertOutput(first []string, replacer *strings.Replacer) (currentMetric metric, err error) {
	var value float64

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

	return currentMetric, err
}

func splitIpmiSensorOutput(commandOutput string, replacer *strings.Replacer) (metrics []metric, err error) {

	r := csv.NewReader(strings.NewReader(commandOutput))
	r.Comma = '|'
	r.Comment = '#'

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return metrics, fmt.Errorf("could not split csv : %v", err)
		}
		result, err := convertOutput(record, replacer)
		if err != nil {
			return metrics, fmt.Errorf("could not convert csv commandOutput : %v", err)
		}
		metrics = append(metrics, result)
	}

	return metrics, err
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
