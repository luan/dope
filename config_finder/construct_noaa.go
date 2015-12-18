package config_finder

import (
	"crypto/tls"
	"errors"
	"net/url"
	"os"

	"github.com/cloudfoundry/noaa"
)

type NoaaConfig struct {
	trafficControllerURL string
}

func NewNoaaConsumer() (*noaa.Consumer, error) {
	config := NoaaConfig{}
	config.PopulateFromEnv()

	err := config.Validate()
	if err != nil {
		return nil, err
	}

	noaaClient := noaa.NewConsumer(
		config.trafficControllerURL,
		&tls.Config{InsecureSkipVerify: config.IsSecure()},
		nil,
	)

	return noaaClient, nil
}

func (c *NoaaConfig) IsSecure() bool {
	u, err := url.Parse(c.trafficControllerURL)
	if err != nil {
		panic("crap")
	}
	return u.Scheme == "wss"
}

func (c *NoaaConfig) PopulateFromEnv() {
	if c.trafficControllerURL == "" {
		c.trafficControllerURL = os.Getenv("TRAFFIC_CONTROLLER_URL")
	}
}

func (c *NoaaConfig) Validate() error {
	if c.trafficControllerURL == "" {
		return errors.New("You must set TRAFFIC_CONTROLLER_URL")
	}

	return nil
}
