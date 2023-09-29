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

	return nil
}

// WithDone configures a pipeline stage to cancel the returned context when all goroutines started by the stage
// have been stopped.
// This is appropriate for termination detection for ANY stages in a pipeline.
// To await termination of ALL stages in a pipeline, use WithWaitGroup.
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
