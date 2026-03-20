package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultFormat = "table"
	configFile    = "config.yaml"
)

type Config struct {
	Version       int                `yaml:"version"`
	ActiveProfile string             `yaml:"active_profile"`
	Defaults      Defaults           `yaml:"defaults"`
	Profiles      map[string]Profile `yaml:"profiles"`
}

type Defaults struct {
	Format string `yaml:"format"`
}

type Profile struct {
	ChannelID string `yaml:"channel_id"`
	TokenType string `yaml:"token_type"` // longterm | stateless | v2
}

func DefaultConfigDir() string {
	if d := os.Getenv(EnvConfigDir); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lm")
}

func Load() (*Config, error) {
	dir := DefaultConfigDir()
	path := filepath.Join(dir, configFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	dir := DefaultConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, configFile), data, 0600)
}

func (c *Config) ActiveProfileConfig() Profile {
	name := c.ActiveProfile
	if name == "" {
		name = "default"
	}
	if p, ok := c.Profiles[name]; ok {
		return p
	}
	return Profile{TokenType: "longterm"}
}

func defaultConfig() *Config {
	return &Config{
		Version:       1,
		ActiveProfile: "default",
		Defaults:      Defaults{Format: DefaultFormat},
		Profiles:      map[string]Profile{},
	}
}
