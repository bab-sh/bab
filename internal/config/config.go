package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/bab-sh/bab/internal/paths"
	"gopkg.in/yaml.v3"
)

const configFileName = "config.yaml"

type Config struct {
	Telemetry TelemetryConfig `yaml:"telemetry,omitempty"`
}

type TelemetryConfig struct {
	Consent *bool `yaml:"consent,omitempty"`
}

func Load() (*Config, error) {
	path, err := paths.ConfigFile(configFileName)
	if err != nil {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := paths.ConfigFile(configFileName)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	success := false
	defer func() {
		if !success {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}

	success = true
	return nil
}

func Update(fn func(*Config) error) error {
	cfg, err := Load()
	if err != nil {
		return err
	}
	if err := fn(cfg); err != nil {
		return err
	}
	return Save(cfg)
}

func Path() string {
	path, _ := paths.ConfigFile(configFileName)
	return path
}

func BoolPtr(v bool) *bool {
	return &v
}
