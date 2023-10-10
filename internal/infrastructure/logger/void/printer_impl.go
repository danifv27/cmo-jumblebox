package logrus

import (
	applogger "fry.org/qft/jumble/internal/application/logger"
)

type printer struct {
}

func (l printer) Log(v ...interface{}) {}

func (l printer) Logf(format string, v ...interface{}) {}

func (l printer) Print(v ...interface{}) {}

func (l printer) Printf(format string, v ...interface{}) {}

func NewPrinter() applogger.Printer {

	return printer{}
}
