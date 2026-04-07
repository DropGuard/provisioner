package config

import (
	"gopkg.in/yaml.v3"
)

// App represents a single piece of software to be installed via Scoop.
type App struct {
	Name   string `yaml:"name"`
	Bucket string `yaml:"bucket"`
	Global bool   `yaml:"global"`
}

// Config represents the application configuration.
type Config struct {
	SetupCommands          []string `yaml:"setup_commands"`
	CreateDesktopShortcuts bool     `yaml:"create_desktop_shortcuts"`
	Apps                   []App    `yaml:"apps"`
}

// LoadBytes loads the configuration directly from byte slice.
func LoadBytes(data []byte) (*Config, error) {
	var cfg Config
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

