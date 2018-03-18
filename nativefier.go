package nativefier

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/jackmordaunt/pageicon"
	"github.com/spf13/afero"
)

// Packager creates an OS specific executable package.
// For OSX this is typically a ".app", windows is ".exe" and linux is an elf
// binary.
type Packager interface {
	Pack(dest string) error
}

// NewPackager instantiates a Packager.
// target: the executable file to bundle.
// title: name of the `.app` folder.
// url: website to point to.
// inferIcon: try to guess the icon.
// inferrer: responsible for doing the inference, defaults to `pageicon.Infer`.
func NewPackager(
	target string,
	title string,
	url string,
	inferIcon bool,

	inferrer IconInferrer,
	fs afero.Fs,
) (Packager, error) {
	if !strings.HasPrefix("http", url) {
		if !strings.HasPrefix("www", url) {
			url = fmt.Sprintf("https://www.%s", url)
		} else {
			url = fmt.Sprintf("https://%s", url)
		}
	}
	if inferrer == nil {
		inferrer = IconInferrerFunc(pageicon.Infer)
	}
	if fs == nil {
		fs = afero.NewOsFs()
	}
	var p Packager
	switch runtime.GOOS {
	case "darwin":
		p = &Darwin{
			Target:    target,
			Title:     title,
			URL:       url,
			InferIcon: inferIcon,
			icon:      inferrer,
			fs:        fs,
		}
	case "windows":
		// Windows{}
	case "linux":
		// Elf{}
	default:
		return nil, fmt.Errorf("no packager implemented for %s",
			runtime.GOOS)
	}
	return p, nil
}
