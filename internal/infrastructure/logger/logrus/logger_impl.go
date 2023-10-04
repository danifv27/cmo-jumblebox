package logrus

import (
	"os"

	"github.com/sirupsen/logrus"

	applogger "fry.org/qft/jumble/internal/application/logger"
)

type logger struct {
	log   *logrus.Logger
	level applogger.LoggerLeveler
}

func (l *logger) Debug(v ...interface{}) {

	switch l.level {
	case applogger.LoggerLevelInfo:
		return
	default:
		l.log.Log(logrus.DebugLevel, v...)
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {

	switch l.level {
	case applogger.LoggerLevelInfo:
		return
	default:
		l.log.Logf(logrus.DebugLevel, format, v...)
	}
}

func (l *logger) Log(v ...interface{}) {

	l.log.Log(logrus.InfoLevel, v...)
}

func (l *logger) Logf(format string, v ...interface{}) {

	l.log.Logf(logrus.InfoLevel, format, v...)
}

func (l *logger) Print(v ...interface{}) {

	l.log.Print(v...)
}

func (l *logger) Printf(format string, v ...interface{}) {

	l.log.Printf(format, v...)
}

// SetLevel sets the logger level.
func (l *logger) SetLevel(lvl applogger.LoggerLeveler) error {

	switch lvl {
	case applogger.LoggerLevelTrace:
	case applogger.LoggerLevelDebug:
		l.level = lvl
	default:
		l.level = applogger.LoggerLevelInfo
	}

	return nil
}

// GetLevel returns the logger level.
func (l *logger) GetLevel() applogger.LoggerLeveler {

	return l.level
}

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
	gLogger.SetLevel(applogger.LoggerLevelInfo)

	return gLogger
}
