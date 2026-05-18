package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const DefaultModel = "gpt-4o-copilot"
const DefaultLanguageServerPackage = "@github/copilot-language-server"

type Config struct {
	Model                 string `json:"model"`
	LanguageServerPackage string `json:"languageServerPackage"`
}

func Defaults() Config {
	return Config{Model: DefaultModel, LanguageServerPackage: DefaultLanguageServerPackage}
}

func DefaultPath() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(base, "helix-copilot", "config.json")
}

func Load(path string) (Config, error) {
	cfg := Defaults()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, err
	}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.LanguageServerPackage == "" {
		cfg.LanguageServerPackage = DefaultLanguageServerPackage
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.LanguageServerPackage == "" {
		cfg.LanguageServerPackage = DefaultLanguageServerPackage
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}
