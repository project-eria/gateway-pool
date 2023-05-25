package main

import (
	"sort"

	"github.com/rs/zerolog/log"
)

func getCorrectedValue(calibration map[float64]float64, rawValue float64) float64 {
	if calibration != nil {
		nbPoints := len(calibration)
		if nbPoints > 0 {
			// Sort keys
			var keys []float64
			for k := range calibration {
				keys = append(keys, k)
			}
			sort.Float64s(keys)
			if nbPoints == 1 {
				return singlePointCalibration(keys, calibration, rawValue)
			} else if nbPoints == 2 {
				return twoPointsCalibration(keys, calibration, rawValue)
			} else {
				log.Error().Int("points", nbPoints).Msg("[main] Too many calibration points")
			}
		}
	}

	return rawValue
}

/* Single point calibration
 * https://learn.adafruit.com/calibrating-sensors/single-point-calibration
 */
func singlePointCalibration(keys []float64, calibration map[float64]float64, rawValue float64) float64 {
	ref := keys[0]
	raw := calibration[ref]
	offset := ref - raw
	return rawValue + offset
}

/* 2 points calibration
 * https://learn.adafruit.com/calibrating-sensors/two-point-calibration
 */
func twoPointsCalibration(keys []float64, calibration map[float64]float64, rawValue float64) float64 {
	refLow := keys[0]
	refHigh := keys[1]
	rawLow := calibration[refLow]
	rawHigh := calibration[refHigh]
	rawRange := rawHigh - rawLow
	refRange := refHigh - refLow

	return (((rawValue - rawLow) * refRange) / rawRange) + refLow
}
