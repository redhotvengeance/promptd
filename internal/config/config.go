package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/redhotvengeance/promptd/internal/promptd"
)

func LoadConfig() (promptd.Config, error) {
	var config promptd.Config

	baseConfigDir, err := os.UserConfigDir()
	if err != nil {
		return config, fmt.Errorf("could not determine user config directory: %w", err)
	}

	configPath := filepath.Join(baseConfigDir, "promptd", "config.toml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := createDefaultConfig(configPath); err != nil {
				return config, fmt.Errorf("failed to create default config: %w", err)
			}

			data = []byte(defaultConfigTemplate)
		} else {
			return config, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := toml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("syntax error in config file %s: %w", configPath, err)
	}

	return config, nil
}

func createDefaultConfig(path string) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("count not create config directory: %w", err)
	}

	if err := os.WriteFile(path, []byte(defaultConfigTemplate), 0600); err != nil {
		return fmt.Errorf("could not write default config file: %w", err)
	}

	return nil
}
