package main

import (
	"math"
	"time"

	eria "github.com/project-eria/eria-core"
	"github.com/project-eria/go-wot/consumer"
	"github.com/rs/zerolog/log"
)

/* Calibration
 * Reference: 468mV 1,995V
 * Pour 1,995: vToORP = 486,98mV
 * Diff = 468 - 486,98 = - 18,98mV
 */

func vToORP(volts float64) float64 {
	/* ORP electrodes give a typical range of -2V to 2V,
	 * where the positive values are for oxidizers
	 * and the negative values are for reducers
	 */
	return math.Round(((2.5 - volts) / 1.037) * 1000)
}

func getORP(volts float64) float64 {
	mv := vToORP(volts)
	cmv := getCorrectedValue(config.ORP.Calibration, mv)
	return cmv
}

func connectORP(eriaClient *eria.EriaClient) *consumer.ConsumedThing {
	if config.ORP.URL != "" {
		log.Info().Str("url", config.ORP.URL).Msg("[main] Connecting ORP thing device...")

		if t, err := eriaClient.ConnectThing(config.ORP.URL); err == nil {
			if config.ORP.Rate > 0 {
				ticker := time.NewTicker(time.Duration(config.ORP.Rate) * time.Second)
				value, err := t.ReadProperty(config.ORP.Property)
				updateORP(value, err)
				go func() {
					for {
						<-ticker.C
						value, err := t.ReadProperty(config.ORP.Property)
						updateORP(value, err)
					}
				}()
			}
			return t
		} else {
			log.Error().Str("url", config.ORP.URL).Err(err).Msg("[main] Can't connect ORP thing device")
		}
	} else {
		log.Warn().Msg("[main] No configured ORP thing device")
	}
	return nil
}

func updateORP(data interface{}, err error) {
	if err != nil {
		log.Error().Err(err).Msg("[main]")
	} else {
		log.Info().Interface("data", data).Msg("[main] ORP sensor raw value received")
		value := getORP(data.(float64))
		_poolThing.SetPropertyValue("orp", int(value))
	}
}

func configureORPActions() {
	if _orpThing != nil {
		_poolThing.SetActionHandler("calibrateORP", func(ref interface{}) (interface{}, error) {
			log.Info().Interface("mV", ref.(float64)).Msg("[main] ORP Calibration requested")
			if config.ORP.Calibration == nil {
				config.ORP.Calibration = map[float64]float64{}
			}
			// TODO wait for stabilised value
			// Get current value
			value, _ := _orpThing.ReadProperty(config.ORP.Property)
			raw := vToORP(value.(float64))
			config.ORP.Calibration[ref.(float64)] = raw
			_configManager.Save()
			return nil, nil
		})
	}
}
