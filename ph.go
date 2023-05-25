package main

import (
	"math"
	"time"

	"github.com/project-eria/eria-core"
	"github.com/project-eria/go-wot/consumer"
	"github.com/rs/zerolog/log"
)

func vToPH(volts float64) float64 {
	var value float64
	if _tempThing != nil {
		temperature := _poolThing.GetPropertyValue("temperature").(float64)
		value = 7 - ((2.5 - volts) / (0.257179 + (0.000941468 * temperature)))
	} else {
		value = ((3.56 * volts) - 1.889)
	}

	// Round to 1 decimal
	value = math.Round(value*10) / 10

	// We fix pH between 0 and 14
	if value < 0 {
		log.Warn().Float64("pH", value).Msg("[main] pH value lower than 0")
		return 0
	}
	if value > 14 {
		log.Warn().Float64("pH", value).Msg("[main] pH value higher than 14")
		return 14
	}
	return value
}

func connectPH(eriaClient *eria.EriaClient) *consumer.ConsumedThing {
	if config.PH.URL != "" {
		log.Info().Str("url", config.PH.URL).Msg("[main] Connecting pH thing device...")
		if t, err := eriaClient.ConnectThing(config.PH.URL); err == nil {
			if config.PH.Rate > 0 {
				ticker := time.NewTicker(time.Duration(config.PH.Rate) * time.Second)
				value, err := t.ReadProperty(config.PH.Property)
				updatePH(value, err)
				go func() {
					for {
						<-ticker.C
						value, err := t.ReadProperty(config.PH.Property)
						updatePH(value, err)
					}
				}()
			}
			return t
		} else {
			log.Error().Str("url", config.PH.URL).Err(err).Msg("[main] Can't connect pH thing device")
		}
	} else {
		log.Warn().Msg("[main] No configured pH thing device")
	}
	return nil
}

func updatePH(data interface{}, err error) {
	if err != nil {
		log.Error().Err(err).Msg("[main]")
	} else {
		log.Info().Interface("data", data).Msg("[main] pH sensor raw value received")
		value := vToPH(data.(float64))
		_poolThing.SetPropertyValue("ph", value)
	}
}

func configurePHActions() {
	if _phThing != nil {
		_poolThing.SetActionHandler("calibratePH", func(ref interface{}) (interface{}, error) {
			log.Info().Interface("mV", ref.(float64)).Msg("[main] pH Calibration requested")
			if config.PH.Calibration == nil {
				config.PH.Calibration = map[float64]float64{}
			}
			// TODO wait for stabilised value
			// Get current value
			value, _ := _orpThing.ReadProperty(config.PH.Property)
			raw := vToPH(value.(float64))
			config.PH.Calibration[ref.(float64)] = raw
			_configManager.Save()
			return nil, nil
		})
	}
}
