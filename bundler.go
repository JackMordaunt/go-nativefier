package main

import (
	"github.com/pkg/errors"
)

var (
	errNoInferrer = errors.New("no icon inferer available")
)

// Packager creates an OS specific executable package.
// For OSX this is typically a ".app", windows is ".exe" and linux is an elf
// binary.
type Packager interface {
	Pack(dest string) error
}
