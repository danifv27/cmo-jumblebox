package logger

const (
	LoggerLevelTrace loggerLevel = iota //0
	LoggerLevelDebug                    //1
	LoggerLevelInfo                     //2
)

type loggerLevel int

func (l loggerLevel) LoggerLevel() loggerLevel {

	return l
}

type LoggerLeveler interface {
	LoggerLevel() loggerLevel
}

// Logger is a generic logging interface
type Logger interface {
	// SetLevel sets the logger level.
	SetLevel(level LoggerLeveler) error
	// GetLevel returns the logger level.
	GetLevel() LoggerLeveler
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	// Log inserts a log entry.  Arguments may be handled in the manner
	// of fmt.Print, but the underlying logger may also decide to handle
	// them differently.
	Log(v ...interface{})
	// Logf insets a log entry.  Arguments are handled in the manner of fmt.Printf.
	Logf(format string, v ...interface{})
}

// Printer is an interface for Print and Printf.
type Printer interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

// Infoer is an interface for Info and Infof.
type Infoer interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})
}
