package config

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// App represents a single piece of software to be installed via Scoop.
type App struct {
	Name            string `yaml:"name"`
	Bucket          string `yaml:"bucket"`
	DesktopShortcut bool   `yaml:"desktop_shortcut"`
}

// Config represents the application configuration.
type Config struct {
	SetupCommands     []string `yaml:"setup_commands"`
	PostSetupCommands []string `yaml:"post_setup_commands"`
	Apps              []App    `yaml:"apps"`
}

// LoadBytes loads the configuration directly from byte slice.
// Unknown fields are rejected to prevent silent misconfiguration.
func LoadBytes(data []byte) (*Config, error) {
	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
