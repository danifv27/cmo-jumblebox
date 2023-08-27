package logrus

import (
	applogger "fry.org/qft/jumble/internal/application/logger"
)

type logger struct {
}

func (l logger) Log(v ...interface{}) {}

func (l logger) Logf(format string, v ...interface{}) {}

func (l logger) Print(v ...interface{}) {}

func (l logger) Printf(format string, v ...interface{}) {}

// WithFields creates a new logger based on logrus.StandardLogger().
func NewLogger() applogger.Logger {

	return logger{}
}
