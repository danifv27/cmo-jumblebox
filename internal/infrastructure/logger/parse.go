package logger

import (
	"fmt"
	"net/url"
	"os"

	"github.com/speijnik/go-errortree"

	alogger "fry.org/qft/jumble/internal/application/logger"
	ilogger "fry.org/qft/jumble/internal/infrastructure/logger/logrus"
)

// URI "logger:logrus?level=<logrus_level>ḉ&output=[plain|json]"
func Parse(URI string) (alogger.Logger, error) {
	var l alogger.Logger
	var rcerror error

	u, err := url.Parse(URI)
	if err != nil {
		return nil, errortree.Add(rcerror, "Parse", err)
	}
	if u.Scheme != "logger" {
		return nil, errortree.Add(rcerror, "Parse", fmt.Errorf("invalid scheme %s", URI))
	}
	switch u.Opaque {
	case "logrus":
		l = ilogger.NewLogger(os.Stdout)
	default:
		return nil, errortree.Add(rcerror, "logger.Parse", fmt.Errorf("unsupported logger implementation %q", u.Opaque))
	}

	return l, nil
}
