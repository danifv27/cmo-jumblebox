package logrus

import (
	"os"

	"github.com/sirupsen/logrus"

	applogger "fry.org/qft/jumble/internal/application/logger"
)

type logger struct {
	log *logrus.Logger
}

func init() {

}

func (l *logger) Log(v ...interface{}) {

	l.log.Log(logrus.InfoLevel, v...)
}

func (l *logger) Logf(format string, v ...interface{}) {

	l.log.Logf(logrus.InfoLevel, format, v...)
}

// WithFields creates a new logger based on logrus.StandardLogger().
func NewLogger(f *os.File) applogger.Logger {

	gLogger := new(logger)
	gLogger.log = logrus.New()
	if f != nil {
		gLogger.log.SetOutput(f) // Set output to stdout; set to stderr by default
	}
	gLogger.log.SetLevel(logrus.TraceLevel)
	// Setup logger defaults
	formatter := logrus.TextFormatter{
		TimestampFormat:  "02-01-2006 15:04:05",
		FullTimestamp:    true,
		DisableTimestamp: false,
		DisableColors:    false,
	}
	gLogger.log.SetFormatter(&formatter)

	return gLogger
}
