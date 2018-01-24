package main

// Logger is a simple interface to handle logging output.
type Logger func(format string, values ...interface{})

var log = struct {
	Errorf Logger
	Debugf Logger
}{
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
	log.Debugf = logger
}

// SetErrorLogger assigns error output to the given logger.
func SetErrorLogger(logger Logger) {
	log.Errorf = logger
}
