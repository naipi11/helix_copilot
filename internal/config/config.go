package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const DefaultModel = "gpt-5.4-mini"
const DefaultLanguageServerPackage = "@github/copilot-language-server"

type Config struct {
	Model                 string `json:"model"`
	LanguageServerPackage string `json:"languageServerPackage"`
	// LanguageServerJSPath, when non-empty, pins an absolute path to the
	// Copilot language server entry script (typically
	// .../@github/copilot-language-server/dist/language-server.js).
	// When set, the launcher invokes it directly via node and skips
	// auto-discovery / npx fallback. Empty means "auto-detect".
	LanguageServerJSPath string `json:"languageServerJSPath,omitempty"`
}

func Defaults() Config {
	return Config{Model: DefaultModel, LanguageServerPackage: DefaultLanguageServerPackage}
}

func DefaultPath() string {
	if runtime.GOOS == "windows" {
		// On Windows, prefer %APPDATA% (which Helix itself uses) so the
		// helix-copilot config sits next to the Helix config tree. This
		// matches defaultHelixConfigDir() in cmd/helix-copilot/main.go.
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "helix-copilot", "config.json")
		}
	}
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
