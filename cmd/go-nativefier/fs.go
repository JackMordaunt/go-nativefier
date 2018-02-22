package main

import "github.com/spf13/afero"

var fs afero.Fs

func init() {
	fs = afero.NewOsFs()
}

// SetFs to your desired filesystem backend.
// Defaults to standard OS filesystem.
func SetFs(filesystem afero.Fs) {
	fs = filesystem
}
