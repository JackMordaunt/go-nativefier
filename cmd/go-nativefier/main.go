package main

import (
	"fmt"
	"os"
	"path/filepath"

	nativefier "github.com/jackmordaunt/go-nativefier"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

func main() {
	var (
		appMode = detectMode()
		url     = ""
		title   = pflag.String("title", "", "Title of app.")
		dest    = pflag.String("output", "./dist", "Directory to put bundle in.")
		// dev     = pflag.Bool("dev", false, "Dev mode enables web inspector for webkit. For windows you can include firebug web inspector.")
	)
	pflag.Parse()
	executable, err := os.Executable()
	if err != nil {
		fmt.Printf("Could not resolve executable path: %v", err)

	}
	if appMode {
		config := filepath.Join(filepath.Dir(executable), "config.json")
		if err := runApp(config); err != nil {
			fatalf("App failed: %v", err)
		}
		return
	}
	if len(pflag.Args()) < 1 {
		pflag.Usage()
		os.Exit(0)
	}
	url = pflag.Args()[0]
	b := nativefier.NewPackager(
		executable,
		*title,
		url,
		true,
		nil,
		nil,
	)
	if err := b.Pack(*dest); err != nil {
		fatalf("Bundle failed: %s", err)
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

func runApp(config string) error {
	cfg, err := loadConfig(config)
	if err != nil {
		return errors.Wrap(err, "reading config")
	}
	wv := nativefier.NewWebview(cfg.Title, cfg.URL, 1280, 800, true, cfg.Debug)
	signals.OnTerminate(func() {
		wv.Exit()
		fmt.Println("\nExiting")
	})
	defer wv.Exit()
	wv.Run()
	return nil
}
func fatalf(f string, v ...interface{}) {
	fmt.Printf(f+"\n", v...)
	os.Exit(1)
}
