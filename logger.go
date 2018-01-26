package main

// Logger is a simple interface to handle logging output.
type Logger func(format string, values ...interface{})

type logger struct {
	Errorf Logger
	Debugf Logger
}

var defaultLogger = logger{
	Errorf: emptyLogger(),
	Debugf: emptyLogger(),
}

func emptyLogger() Logger {
	return func(string, ...interface{}) {
		return
	}
}

// SetDebugLogger assigns debug output to the given logger.
func SetDebugLogger(logger Logger) {
	defaultLogger.Debugf = logger
}

// SetErrorLogger assigns error output to the given logger.
func SetErrorLogger(logger Logger) {
	defaultLogger.Errorf = logger
}
