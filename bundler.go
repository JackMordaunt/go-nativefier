package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"github.com/disintegration/imaging"
	"github.com/muesli/smartcrop"
	"github.com/spf13/afero"

	"github.com/jackmordaunt/pageicon"
	"github.com/nfnt/resize"
	"github.com/pkg/errors"
)

var (
	errNoInferrer = errors.New("no icon inferer available")
)

// Bundler bundles an executable into a native package with, passing the default
// arguments to it.
type Bundler struct {
	Title     string
	URL       string
	InferIcon bool

	Target string
	Icon   string

	icon IconInferrer
	fs   afero.Fs
	log  logger
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

) *Bundler {
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
	b := &Bundler{
		Target:    target,
		Title:     title,
		URL:       url,
		InferIcon: inferIcon,
		icon:      inferrer,
		log:       defaultLogger,
		fs:        fs,
	}
	return b
}

// Bundle into dest, which must be a valid file system path.
func (b *Bundler) Bundle(dest string) error {
	name := fmt.Sprintf("%s.app", b.Title)
	app := filepath.Join(dest, name, "Contents")
	macos := filepath.Join(app, "MacOS")
	resources := filepath.Join(app, "Resources")
	if err := b.prepare(app, macos, resources); err != nil {
		return err
	}
	b.log.Debugf("Creating executable\n")
	if err := b.CreateExecutable(macos); err != nil {
		return errors.Wrap(err, "creating executable")
	}
	b.log.Debugf("Creating config\n")
	if err := b.CreateConfig(macos); err != nil {
		return err
	}
	if b.InferIcon {
		b.log.Debugf("Inferring icon\n")
		if err := b.FetchIcon(resources); err != nil {
			b.log.Debugf("[error] %s\n", err)
		}
	}
	b.log.Debugf("Creating Info.plist\n")
	return b.CreatePlist(app)
}

func (b Bundler) prepare(paths ...string) error {
	for _, p := range paths {
		if err := b.fs.MkdirAll(p, 0755); err != nil {
			return err
		}
	}
	return nil
}

// CreateConfig creates the config file relative to dest.
func (b *Bundler) CreateConfig(dest string) error {
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
func (b *Bundler) CreateExecutable(dest string) error {
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
func (b *Bundler) FetchIcon(dest string) error {
	if b.icon == nil {
		return errNoInferrer
	}
	b.log.Debugf("url is %q\n", b.URL)
	icon, err := b.icon.Infer(b.URL, []string{"png", "jpg", "ico"})
	if err != nil {
		return errors.Wrap(err, "inferring icon")
	}
	b.log.Debugf("inferred icon: %v\n", icon.Source)
	if icon == nil {
		return errors.New("could not infer icon")
	}
	converted, err := b.convertIcon(icon)
	if err != nil {
		return errors.Wrap(err, "converting icon")
	}
	b.log.Debugf("icon converted\n")
	path := filepath.Join(dest, "icon.icns")
	return writeFile(b.fs, path, converted.Data)
}

// CreatePlist creates an Info.plist file, relative to dest.
func (b *Bundler) CreatePlist(dest string) error {
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
// depends on iconutil (macos native) and the filesystem.
//
// Note: We are forced to use the native filesystem to convert the icon.
// The only way to mock the filesytem here would be to intercept the file i/o
// performed by iconutil, which I don't know how to do. Named pipes?
//
// The os dependency doesn't leak from this method.
//
// Todo(jackmordaunt) Look into writing a Go implementation of
// iconutil to avoid the dependency.
func (b *Bundler) convertIcon(icon *Icon) (*Icon, error) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(icon.Data)
	if err != nil {
		return nil, err
	}
	// Prepare image.
	original, _, err := image.Decode(bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "decoding")
	}
	// Todo: find size without hard coding 128x128
	// This code is ugly/brittle.
	cropped := imaging.Fill(original, 128, 128, imaging.Center, imaging.Lanczos)
	size, err := closestIconsetSize(cropped)
	if err != nil {
		return nil, err
	}
	scale := ""
	nameSize := size
	if size == 1024 {
		nameSize = 512
		scale = "@2x"
	}
	// Prepare files for iconutil.
	iconName := fmt.Sprintf("icon_%dx%d%s.png", nameSize, nameSize, scale)
	resizedPng := resize.Resize(uint(size), uint(size), cropped, resize.Bicubic)
	resizedPngBuf := bytes.NewBuffer(nil)
	if err := png.Encode(resizedPngBuf, resizedPng); err != nil {
		return nil, errors.Wrap(err, "encoding png")
	}
	tmpPath := filepath.Join(tmp, "icon.iconset", iconName)
	tmpDir := filepath.Dir(tmpPath)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, err
	}
	wr, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(wr, resizedPngBuf); err != nil {
		wr.Close()
		return nil, err
	}
	wr.Close()
	// Create the .icns file.
	iconPath := filepath.Join(
		filepath.Dir(tmpDir),
		"icon.icns",
	)
	cmd := exec.Command(
		"iconutil",
		"-c", "icns",
		tmpDir,
		iconPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "iconutil failed: %+v: %s",
			cmd.Args, string(output))
	}
	converted, err := pageicon.NewFromFile(iconPath)
	if err != nil {
		return nil, err
	}
	return converted, nil
}

func smartCrop(img image.Image) (image.Image, error) {
	a := smartcrop.NewAnalyzer()
	cropped, err := a.FindBestCrop(img, 512, 512)
	if err != nil {
		return nil, err
	}
	return cropped, nil
}

func closestIconsetSize(img image.Image) (int, error) {
	dimensions := img.Bounds().Size()
	biggest := dimensions.X
	if dimensions.Y > biggest {
		biggest = dimensions.Y
	}
	closest := struct {
		Set      bool
		Distance int
		Index    int
	}{}
	for ii, size := range iconsetSizes {
		distance := size - biggest
		if distance < 0 {
			distance = distance * -1
		}
		if !closest.Set {
			closest.Distance = distance
			closest.Index = ii
			closest.Set = true
		}
		if distance < closest.Distance {
			closest.Distance = distance
			closest.Index = ii
		}
	}
	return iconsetSizes[closest.Index], nil
}

func duplicate(data []byte) []byte {
	ret := make([]byte, len(data))
	copy(ret, data)
	return ret
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

var iconsetSizes = []int{
	16,
	32,
	128,
	256,
	512,
	1024,
}

// pipe returns handles for reading from  and writing to a named pipe.
// io.Writer writes to the pipe.
// io.Reader reads from the pipe.
// io.Closer closes the pipe, releasing the file handle.
func pipe(path string) (io.Writer, io.Reader, io.Closer, error) {
	if err := syscall.Mkfifo(path, 0755); err != nil {
		return nil, nil, nil, err
	}
	wr, err := os.OpenFile(path, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, nil, nil, err
	}
	r, err := os.OpenFile(path, os.O_RDONLY, os.ModeNamedPipe)
	if err != nil {
		return nil, nil, nil, err
	}
	return wr, r, wr, nil
}
