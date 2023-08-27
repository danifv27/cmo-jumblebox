package tail

import (
	"errors"
	"time"

	aconfigurable "fry.org/qft/jumble/internal/application/configurable"
	alogger "fry.org/qft/jumble/internal/application/logger"
	ilogger "fry.org/qft/jumble/internal/infrastructure/logger"
	"github.com/speijnik/go-errortree"
)

type tailCfg struct {
	rotatedFilePathPatterns []string
	positionFile            PositionFile
	readFromHead            bool
	tailCfgFollowRotate
	printer alogger.Printer
}

type tailCfgFollowRotate struct {
	detectRotateDelay   time.Duration
	followRotate        bool
	watchRotateInterval time.Duration
}

func newTailCfg() (tailCfg, error) {
	var rcerror, err error
	var printer alogger.Printer

	cfg := tailCfg{}
	if _, printer, _, err = ilogger.Parse("printer:void"); err != nil {
		return cfg, errortree.Add(rcerror, "newTailCfg", err)
	}
	cfg.printer = printer

	return cfg, nil
}

func (cfg *tailCfg) Configure(configs ...aconfigurable.ConfigurablerFn) error {
	var rcerror error

	// Set default value
	cfg.detectRotateDelay = 5 * time.Second
	cfg.followRotate = true
	cfg.readFromHead = false
	cfg.watchRotateInterval = 100 * time.Millisecond

	// Loop through each option
	for _, c := range configs {
		if err := c(cfg); err != nil {
			return errortree.Add(rcerror, "Configure", err)
		}
	}

	return nil
}

func WithPrinter(p alogger.Printer) aconfigurable.ConfigurablerFn {

	return aconfigurable.ConfigurablerFn(func(i interface{}) error {
		var rcerror error
		var f *tailCfg
		var ok bool

		if f, ok = i.(*tailCfg); ok {
			f.printer = p
			return nil
		}

		return errortree.Add(rcerror, "option.WithLogger", errors.New("type mismatch, option expected"))
	})
}

// WithRotatedFilePathPatterns let you change rotatedFilePathPatterns
func WithRotatedFilePathPatterns(globPatterns []string) aconfigurable.ConfigurablerFn {

	return aconfigurable.ConfigurablerFn(func(i interface{}) error {
		var rcerror error
		var f *tailCfg
		var ok bool

		if f, ok = i.(*tailCfg); ok {
			f.rotatedFilePathPatterns = globPatterns
			return nil
		}

		return errortree.Add(rcerror, "option.WithLogger", errors.New("type mismatch, option expected"))
	})
}

// WithPositionFile let you change positionFile
func WithPositionFile(positionFile PositionFile) aconfigurable.ConfigurablerFn {

	return aconfigurable.ConfigurablerFn(func(i interface{}) error {
		var rcerror error
		var f *tailCfg
		var ok bool

		if f, ok = i.(*tailCfg); ok {
			f.positionFile = positionFile
			return nil
		}

		return errortree.Add(rcerror, "option.WithPositionFile", errors.New("type mismatch, option expected"))
	})
}

// WithPositionFilePath let you change positionFile
func WithPositionFilePath(path string) (aconfigurable.ConfigurablerFn, error) {
	if path == "" {
		return WithPositionFile(nil), nil
	}
	pf, err := OpenPositionFile(path)
	if err != nil {
		return nil, err
	}
	return WithPositionFile(pf), nil
}

// WithDetectRotateDelay let you change detectRotateDelay
func WithDetectRotateDelay(v time.Duration) aconfigurable.ConfigurablerFn {

	return aconfigurable.ConfigurablerFn(func(i interface{}) error {
		var rcerror error
		var f *tailCfg
		var ok bool

		if f, ok = i.(*tailCfg); ok {
			f.detectRotateDelay = v
			return nil
		}

		return errortree.Add(rcerror, "option.WithDetectRotateDelay", errors.New("type mismatch, option expected"))
	})
}

// WithFollowRotate let you change followRotate
func WithFollowRotate(follow bool) aconfigurable.ConfigurablerFn {

	return aconfigurable.ConfigurablerFn(func(i interface{}) error {
		var rcerror error
		var f *tailCfg
		var ok bool

		if f, ok = i.(*tailCfg); ok {
			f.followRotate = follow
			return nil
		}

		return errortree.Add(rcerror, "option.WithFollowRotate", errors.New("type mismatch, option expected"))
	})
}

// WithReadFromHead let you change readFromHead
func WithReadFromHead(v bool) aconfigurable.ConfigurablerFn {

	return aconfigurable.ConfigurablerFn(func(i interface{}) error {
		var rcerror error
		var f *tailCfg
		var ok bool

		if f, ok = i.(*tailCfg); ok {
			f.readFromHead = v
			return nil
		}

		return errortree.Add(rcerror, "option.WithReadFromHead", errors.New("type mismatch, option expected"))
	})
}

// WithWatchRotateInterval let you change watchRotateInterval
func WithWatchRotateInterval(v time.Duration) aconfigurable.ConfigurablerFn {

	return aconfigurable.ConfigurablerFn(func(i interface{}) error {
		var rcerror error
		var f *tailCfg
		var ok bool

		if f, ok = i.(*tailCfg); ok {
			f.watchRotateInterval = v
			return nil
		}

		return errortree.Add(rcerror, "option.WithWatchRotateInterval", errors.New("type mismatch, option expected"))
	})
}
