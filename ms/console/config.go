package console

import (
	"github.com/jansemmelink/utils2/errors"
	"github.com/jansemmelink/utils2/ms"
)

func init() {
	ms.RegisterServer("console", &Config{Msisdn: "27821234567"})
}

type Config struct {
	Msisdn string `json:"msisdn"`
	//Data map[string]interface{} `json:"data"`
}

func (c *Config) Validate() error {
	if c.Msisdn == "" {
		return errors.Errorf("missing msisdn")
	}
	return nil
}

func (c Config) Create() (ms.Server, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid config")
	}
	s := server{
		config: c,
	}
	return s, nil
}
