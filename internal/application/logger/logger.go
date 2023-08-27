package logger

// Logger is a generic logging interface
type Logger interface {
	// Log inserts a log entry.  Arguments may be handled in the manner
	// of fmt.Print, but the underlying logger may also decide to handle
	// them differently.
	Log(v ...interface{})
	// Logf insets a log entry.  Arguments are handled in the manner of
	// fmt.Printf.
	Logf(format string, v ...interface{})
}

// Printer is an interface for Print and Printf.
type Printer interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
}

// Infoer is an interface for Infoer and Infof.
type Infoer interface {
	Info(v ...interface{})
	Infof(format string, v ...interface{})
}
