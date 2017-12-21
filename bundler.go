package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Bundler bundles an executable into a native package with, passing the default
// arguments to it.
type Bundler struct {
	Target  string
	Title   string
	Address string
}

// Bundle into dest, which must be a valid file system path.
func (b *Bundler) Bundle(dest string) error {
	app := filepath.Join(dest, fmt.Sprintf("%s.app", b.Title), "Contents")
	macos := filepath.Join(app, "MacOS")
	if err := os.MkdirAll(macos, 0777); err != nil {
		return err
	}
	target, err := os.Open(b.Target)
	if err != nil {
		return err
	}
	defer target.Close()
	destFile, err := os.OpenFile(filepath.Join(macos, b.Title), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer destFile.Close()
	if _, err := io.Copy(destFile, target); err != nil {
		return err
	}
	if err := b.CreatePlist(app); err != nil {
		return err
	}
	config, err := json.Marshal(map[string]string{
		"Title":   b.Title,
		"Address": b.Address,
	})
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(macos, "config.json"), config, 0644)
}

// CreatePlist creates and Info.plist file, relative to dest.
func (b *Bundler) CreatePlist(dest string) error {
	data := map[string]string{
		"ExecutableName": b.Title,
		"Identifier":     strings.TrimSpace(b.Title),
		"BundleName":     b.Title,
	}
	t, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}
	path := filepath.Join(dest, "Info.plist")
	infoPlist, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer infoPlist.Close()
	return t.Execute(infoPlist, data)
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
	<key>QuartzGLEnable</key>
	<false/>
</dict>
</plist>`

// Markup required for icons
// <key>CFBundleIconFile</key>
// <string>{{.IconName}}</string>
