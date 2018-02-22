package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/jackmordaunt/icns"

	"github.com/spf13/afero"

	"github.com/jackmordaunt/pageicon"
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

// Darwin bundles an executable into a native package with, passing the default
// arguments to it.
type Darwin struct {
	Title     string
	URL       string
	InferIcon bool

	Target string
	Icon   string

	icon IconInferrer
	fs   afero.Fs
}

// NewBundler instantiates a Bundler.
// target: the executable file to bundle.
// title: name of the `.app` folder.
// url: website to point to.
// inferIcon: try to guess the icon.
// inferrer: responsible for doing the inference, defaults to `pageicon.Infer`.
func NewBundler(
	target string,
	title string,
	url string,
	inferIcon bool,

	inferrer IconInferrer,

) *Darwin {
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
	b := &Darwin{
		Target:    target,
		Title:     title,
		URL:       url,
		InferIcon: inferIcon,
		icon:      inferrer,
		fs:        fs,
	}
	return b
}

// Bundle into dest, which must be a valid file system path.
func (b *Darwin) Bundle(dest string) error {
	name := fmt.Sprintf("%s.app", b.Title)
	app := filepath.Join(dest, name, "Contents")
	macos := filepath.Join(app, "MacOS")
	resources := filepath.Join(app, "Resources")
	if err := b.prepare(app, macos, resources); err != nil {
		return err
	}
	if err := b.CreateExecutable(macos); err != nil {
		return errors.Wrap(err, "creating executable")
	}
	if err := b.CreateConfig(macos); err != nil {
		return err
	}
	if b.InferIcon {
		if err := b.FetchIcon(resources); err != nil {
		}
	}
	return b.CreatePlist(app)
}

func (b Darwin) prepare(paths ...string) error {
	for _, p := range paths {
		if err := b.fs.MkdirAll(p, 0755); err != nil {
			return err
		}
	}
	return nil
}

// CreateConfig creates the config file relative to dest.
func (b *Darwin) CreateConfig(dest string) error {
	config, err := json.Marshal(map[string]interface{}{
		"Title": b.Title,
		"URL":   b.URL,
	})
	if err != nil {
		return err
	}
	path := filepath.Join(dest, "config.json")
	return afero.WriteFile(b.fs, path, config, 0755)
}

// CreateExecutable creates the main program executable relative to dest.
func (b *Darwin) CreateExecutable(dest string) error {
	target, err := b.fs.Open(b.Target)
	if err != nil {
		return err
	}
	defer target.Close()
	destFile, err := b.fs.OpenFile(
		filepath.Join(dest, filepath.Base(b.Target)),
		os.O_CREATE|os.O_RDWR, 0755,
	)
	if err != nil {
		return err
	}
	defer destFile.Close()
	if _, err := io.Copy(destFile, target); err != nil {
		return err
	}
	return nil
}

// FetchIcon creates and icon file relative to dest.
func (b *Darwin) FetchIcon(dest string) error {
	if b.icon == nil {
		return errNoInferrer
	}
	icon, err := b.icon.Infer(b.URL, []string{"png", "jpg", "ico"})
	if err != nil {
		return errors.Wrap(err, "inferring icon")
	}
	if icon == nil {
		return errors.New("could not infer icon")
	}
	converted, err := b.convertIcon(icon)
	if err != nil {
		return errors.Wrap(err, "converting icon")
	}
	path := filepath.Join(dest, "icon.icns")
	return writeFile(b.fs, path, converted.Data)
}

// CreatePlist creates an Info.plist file, relative to dest.
func (b *Darwin) CreatePlist(dest string) error {
	data := map[string]string{
		"ExecutableName": b.Title,
		"Identifier":     strings.TrimSpace(b.Title),
		"BundleName":     b.Title,
		"IconName":       b.Icon,
	}
	t, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}
	path := filepath.Join(dest, "Info.plist")
	infoPlist, err := b.fs.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer infoPlist.Close()
	return t.Execute(infoPlist, data)
}

// convertIcon from png to icns.
func (b *Darwin) convertIcon(icon *Icon) (*Icon, error) {
	buf := bytes.NewBuffer(nil)
	img, _, err := image.Decode(icon.Data)
	if err != nil {
		return nil, err
	}
	if err := icns.Encode(buf, img); err != nil {
		return nil, err
	}
	return &Icon{
		Data: buf,
		Mime: "image/icns",
		Ext:  "icns",
	}, nil
}

func writeFile(fs afero.Fs, path string, r io.Reader) error {
	f, err := fs.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	return nil
}

var plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Application-Group</key>
	<array>
		<string>dot-mac</string>
	</array>
	<key>CFBundleDevelopmentRegion</key>
	<string>English</string>
	<key>CFBundleExecutable</key>
	<string>{{.ExecutableName}}</string>
	
	<key>CFBundleIdentifier</key>
	<string>com.web.{{.Identifier}}</string>
	<key>CFBundleName</key>
	<string>{{.BundleName}}</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleSupportedPlatforms</key>
	<array>
		<string>MacOSX</string>
	</array>
	<key>NSSupportsSeamlessOpening</key>
	<true/>
	<key>NSSupportsSuddenTermination</key>
	<true/>
	<key>NSHighResolutionCapable</key>
	<true/>
	<key>CFBundleIconFile</key>
	<string>{{.IconName}}</string>
</dict>
</plist>`
