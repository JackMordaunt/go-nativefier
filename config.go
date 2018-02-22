package nativefier

import (
	"encoding/json"

	"github.com/spf13/afero"
)

type config struct {
	Title string
	URL   string
	Debug bool
}

func loadConfig(path string) (*config, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, err
	}
	cfg := &config{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
