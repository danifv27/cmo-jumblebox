package logger

import (
	"fmt"
	"net/url"
	"os"

	"github.com/speijnik/go-errortree"

	alogger "fry.org/qft/jumble/internal/application/logger"
	ilogger "fry.org/qft/jumble/internal/infrastructure/logger/logrus"
	ivoid "fry.org/qft/jumble/internal/infrastructure/logger/void"
)

// URI "logger:logrus?level=<logrus_level>ḉ&output=[plain|json]"
// URI "printer:void"
func Parse(URI string) (alogger.Logger, alogger.Printer, alogger.Infoer, error) {
	var l alogger.Logger
	var rcerror error

	u, err := url.Parse(URI)
	if err != nil {
		return nil, nil, nil, errortree.Add(rcerror, "Parse", err)
	}
	if u.Scheme != "logger" || u.Scheme != "printer" || u.Scheme != "infoer" {

	}
	switch u.Scheme {
	case "logger":
		switch u.Opaque {
		case "logrus":
			l = ilogger.NewLogger(os.Stdout)
		default:
			return nil, nil, nil, errortree.Add(rcerror, "logger.Parse", fmt.Errorf("unsupported logger implementation %q", u.Opaque))
		}
	case "printer":
		switch u.Opaque {
		case "void":
			return nil, ivoid.NewPrinter(), nil, nil
		default:
			return nil, nil, nil, errortree.Add(rcerror, "logger.Parse", fmt.Errorf("unsupported printer implementation %q", u.Opaque))
		}
	case "infoer":
		switch u.Opaque {
		default:
			return nil, nil, nil, errortree.Add(rcerror, "logger.Parse", fmt.Errorf("unsupported infoer implementation %q", u.Opaque))
		}
	}

	return nil, nil, nil, errortree.Add(rcerror, "Parse", fmt.Errorf("invalid scheme %s", URI))
}
