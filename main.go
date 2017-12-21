package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/zserge/webview"
)

func fatalf(f string, v ...interface{}) {
	fmt.Printf(f+"\n", v...)
	os.Exit(1)
}

func main() {
	var (
		appMode = detectMode()
		title   = pflag.String("title", "", "Title of app.")
		address = pflag.String("address", "", "Address to fetch resources from.")
		dest    = pflag.String("output", "./dist", "Directory to put package in.")
	)
	pflag.Parse()
	executable, err := os.Executable()
	if err != nil {
		fatalf("Could not resolve executable path: %s", err)
	}
	if appMode {
		cfgPath := filepath.Join(filepath.Dir(executable), "config.json")
		cfgData, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			fatalf("Reading config: %v", err)
		}
		cfg := &struct {
			Title   string
			Address string
		}{}
		if err := json.Unmarshal(cfgData, cfg); err != nil {
			fatalf("Reading config: %v", err)
		}
		fmt.Printf("Title: %s\n", cfg.Title)
		fmt.Printf("Address: %s\n", cfg.Address)
		if err := webview.Open(cfg.Title, cfg.Address, 800, 600, true); err != nil {
			fatalf("Webview error: %s", err)
		}
		return
	}
	fmt.Printf("Bundle created: %s\n", filepath.Join(*dest, fmt.Sprintf("%s.app", *title)))
	b := Bundler{
		Target:  executable,
		Title:   *title,
		Address: *address,
	}
	if err := b.Bundle(*dest); err != nil {
		fatalf("Bundle error: %s", err)
	}
}

// appMode is true if config file is present.
func detectMode() bool {
	executable, err := os.Executable()
	if err != nil {
		fatalf("Could not get executale path: %v", err)
	}
	configPath := filepath.Join(filepath.Dir(executable), "config.json")
	fi, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		fatalf("Error while stat'ing config file: %v", err)
	}
	if fi.IsDir() {
		fatalf("Config 'file' is a directory.")
	}
	return true
}
