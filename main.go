package main

import (
	eria "github.com/project-eria/eria-core"
	configmanager "github.com/project-eria/eria-core/config-manager"
	"github.com/project-eria/go-wot/consumer"
	zlog "github.com/rs/zerolog/log"
)

var config = struct {
	Host        string `yaml:"host"`
	Port        uint   `yaml:"port" default:"80"`
	ExposedAddr string `yaml:"exposedAddr"`
	PH          struct {
		URL         string              `yaml:"url"`
		Property    string              `yaml:"property"`
		Rate        int                 `yaml:"rate"` // seconds
		Calibration map[float64]float64 `yaml:"calibration,omitempty"`
	} `yaml:"ph"`
	ORP struct {
		URL         string              `yaml:"url"`
		Property    string              `yaml:"property"`
		Rate        int                 `yaml:"rate"` // seconds
		Calibration map[float64]float64 `yaml:"calibration,omitempty"`
	} `yaml:"orp"`
	Temperature struct {
		URL         string              `yaml:"url"`
		Property    string              `yaml:"property"`
		Rate        int                 `yaml:"rate"` // seconds
		Calibration map[float64]float64 `yaml:"calibration,omitempty"`
	} `yaml:"temperature"`
}{}

var (
	_configManager *configmanager.ConfigManager
	_poolThing     *eria.EriaThing
	_orpThing      *consumer.ConsumedThing
	_phThing       *consumer.ConsumedThing
	_tempThing     *consumer.ConsumedThing
)

func main() {
	defer func() {
		zlog.Info().Msg("[main] Stopped")
	}()

	eria.Init("ERIA Pool Manager")

	// Loading config
	_configManager = eria.LoadConfig(&config)

	eriaServer := eria.NewServer(config.Host, config.Port, config.ExposedAddr, "")

	poolTD, err := eria.NewThingDescription(
		"eria:manager:pool",
		eria.AppVersion,
		"Pool",
		"Pool Manager",
		[]string{"PHSensor", "ORPSensor", "TemperatureSensor"},
	)

	if err != nil {
		zlog.Fatal().Err(err).Msg("[main] NewThingDescription Error")
	}

	_poolThing, _ = eriaServer.AddThing("", poolTD)

	eriaClient := eria.NewClient()
	_tempThing = connectTemperature(eriaClient) // Before pH, to adjust pH based on °C
	_orpThing = connectORP(eriaClient)
	_phThing = connectPH(eriaClient)
	if _orpThing != nil {
		configureORPActions()
	}
	if _phThing != nil {
		configurePHActions()
	}
	eriaServer.StartServer()
}
