package main

import (
	"time"

	eria "github.com/project-eria/eria-core"
	"github.com/project-eria/go-wot/consumer"
	"github.com/rs/zerolog/log"
)

func connectTemperature(eriaClient *eria.EriaClient) *consumer.ConsumedThing {
	if config.Temperature.URL != "" {
		log.Info().Str("url", config.Temperature.URL).Msg("[main] Connecting Temperature thing device...")
		if t, err := eriaClient.ConnectThing(config.Temperature.URL); err == nil {
			if config.Temperature.Rate > 0 {
				ticker := time.NewTicker(time.Duration(config.Temperature.Rate) * time.Second)
				value, err := t.ReadProperty(config.Temperature.Property)
				updateTemperature(value, err)
				go func() {
					for {
						<-ticker.C
						value, err := t.ReadProperty(config.Temperature.Property)
						updateTemperature(value, err)
					}
				}()
			}
			return t
		} else {
			log.Error().Str("url", config.Temperature.URL).Err(err).Msg("[main] Can't connect Temperature thing device")
		}
	} else {
		log.Warn().Msg("[main] No configured Temperature thing device")
	}
	return nil
}

func updateTemperature(data interface{}, err error) {
	if err != nil {
		log.Error().Err(err).Msg("[main]")
	} else {
		log.Info().Interface("data", data).Msg("[main] Temperature Value updated")

		value := data.(float64)
		_poolThing.SetPropertyValue("temperature", value)
	}
}
