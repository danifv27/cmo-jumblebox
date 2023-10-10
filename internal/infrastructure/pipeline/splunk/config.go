package splunk

import (
	"errors"
	"fmt"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"

	"github.com/speijnik/go-errortree"
)

type splunkConfig struct {
	name       string
	buffersize int
	outputs    int
	workers    int
}

func NewSplunkConfig(configs ...aconfigurable.ConfigurablerFn) aconfigurable.Configurabler {

	cfg := splunkConfig{}
	cfg.Configure(configs...)

	return &cfg
}

func (config *splunkConfig) Configure(configs ...aconfigurable.ConfigurablerFn) error {
	var rcerror error

	// Set default value
	config.buffersize = 0
	config.outputs = 1
	config.workers = 1

	// Loop through each option
	for _, c := range configs {
		if err := c(config); err != nil {
			return errortree.Add(rcerror, "Configure", err)
		}
	}
	// spew.Dump(config)

	return nil
}

func WithName(name string) aconfigurable.ConfigurablerFn {

	return func(cfg interface{}) error {
		var rcerror error

		if c, ok := cfg.(*splunkConfig); ok {
			c.name = name
			return nil
		}

		return errortree.Add(rcerror, "WithName", errors.New("type mismatch, *splunkConfig expected"))
	}
}

func WithBufferSize(size int) aconfigurable.ConfigurablerFn {

	return func(cfg interface{}) error {
		var rcerror error

		if c, ok := cfg.(*splunkConfig); ok {
			c.buffersize = size
			return nil
		}

		return errortree.Add(rcerror, "WithBufferSize", errors.New("type mismatch, *splunkConfig expected"))
	}
}

func (config *splunkConfig) GetInt(key string) (int, error) {
	var rcerror error

	switch key {
	case "buffersize":
		return config.buffersize, nil
	case "outputs":
		return config.outputs, nil
	case "workers":
		return config.workers, nil
	}

	return 0, errortree.Add(rcerror, "GetInt", fmt.Errorf("key %s not found", key))
}

func (config *splunkConfig) GetString(key string) (string, error) {
	var rcerror error

	switch key {
	case "name":
		return config.name, nil
	}

	return "", errortree.Add(rcerror, "GetInt", fmt.Errorf("key %s not found", key))
}
