package main

import (
	"fmt"
	"os"
)

// Logger provides a simple logging api.
type Logger interface {
	Info(format string, values ...interface{})
	Debug(format string, values ...interface{})
}

// DefaultLogger implements Logger.
type DefaultLogger struct {
	Verbose bool
}

// Info prints to stdout.
func (l DefaultLogger) Info(format string, values ...interface{}) {
	fmt.Fprintf(
		os.Stdout,
		fmt.Sprintf("%s", format),
		values...,
	)
}

// Debug prints to stdout if 'Verbose' is true.
func (l DefaultLogger) Debug(format string, values ...interface{}) {
	if l.Verbose {
		fmt.Fprintf(
			os.Stdout,
			fmt.Sprintf("[debug] %s", format),
			values...,
		)
	}
}
