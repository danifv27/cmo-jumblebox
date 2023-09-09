package splunk

import (
	"fmt"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"

	"github.com/speijnik/go-errortree"
)

type splunkConfig struct {
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

	return nil
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
