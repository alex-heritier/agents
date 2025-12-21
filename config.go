package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileSpec struct {
	File string `json:"file"`
	Dir  string `json:"dir"`
}

type ProviderConfig struct {
	Name       string    `json:"name,omitempty"`
	Guidelines *FileSpec `json:"guidelines,omitempty"`
	Commands   *FileSpec `json:"commands,omitempty"`
}

type SourcesConfig struct {
	Guidelines string `json:"guidelines"`
	Commands   string `json:"commands"`
}

type ProvidersConfig struct {
	Sources          SourcesConfig             `json:"sources"`
	GlobalGuidelines []string                  `json:"globalGuidelines"`
	Providers        map[string]ProviderConfig `json:"providers"`
}

var (
	configOnce sync.Once
	configData ProvidersConfig
	configErr  error
)

func getProviderConfig() (ProvidersConfig, error) {
	configOnce.Do(func() {
		configData, configErr = loadProviderConfig()
	})
	return configData, configErr
}

func loadProviderConfig() (ProvidersConfig, error) {
	configPath, err := findConfigPath()
	if err != nil {
		return ProvidersConfig{}, err
	}

	baseConfig, err := loadConfigFile(configPath)
	if err != nil {
		return ProvidersConfig{}, err
	}

	userConfigPath, ok := userConfigFilePath()
	if ok && fileExists(userConfigPath) {
		overrideConfig, err := loadConfigFile(userConfigPath)
		if err != nil {
			return ProvidersConfig{}, fmt.Errorf("failed to load user config %s: %w", userConfigPath, err)
		}
		baseConfig = mergeProviderConfig(baseConfig, overrideConfig)
	}

	return normalizeProviderConfig(baseConfig), nil
}

func findConfigPath() (string, error) {
	const configFile = "providers.json"
	if fileExists(configFile) {
		return configFile, nil
	}

	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		candidate := filepath.Join(execDir, configFile)
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	return "", errors.New("providers.json not found")
}

func loadConfigFile(path string) (ProvidersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProvidersConfig{}, err
	}

	var cfg ProvidersConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return ProvidersConfig{}, err
	}
	return cfg, nil
}

func userConfigFilePath() (string, bool) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", false
		}
		xdgConfigHome = filepath.Join(home, ".config")
	}

	return filepath.Join(xdgConfigHome, "agents", "providers.json"), true
}

func mergeProviderConfig(base, override ProvidersConfig) ProvidersConfig {
	if override.Sources.Guidelines != "" {
		base.Sources.Guidelines = override.Sources.Guidelines
	}
	if override.Sources.Commands != "" {
		base.Sources.Commands = override.Sources.Commands
	}

	if len(override.GlobalGuidelines) > 0 {
		base.GlobalGuidelines = append(base.GlobalGuidelines, override.GlobalGuidelines...)
	}

	if base.Providers == nil {
		base.Providers = map[string]ProviderConfig{}
	}

	for name, overrideProvider := range override.Providers {
		baseProvider, ok := base.Providers[name]
		if !ok {
			base.Providers[name] = overrideProvider
			continue
		}

		base.Providers[name] = mergeProvider(baseProvider, overrideProvider)
	}

	return base
}

func mergeProvider(base, override ProviderConfig) ProviderConfig {
	if override.Name != "" {
		base.Name = override.Name
	}

	if override.Guidelines != nil {
		if base.Guidelines == nil {
			base.Guidelines = &FileSpec{}
		}
		base.Guidelines = mergeFileSpec(base.Guidelines, override.Guidelines)
	}

	if override.Commands != nil {
		if base.Commands == nil {
			base.Commands = &FileSpec{}
		}
		base.Commands = mergeFileSpec(base.Commands, override.Commands)
	}

	return base
}

func mergeFileSpec(base, override *FileSpec) *FileSpec {
	if override.File != "" {
		base.File = override.File
	}
	if override.Dir != "" {
		base.Dir = override.Dir
	}
	return base
}

func normalizeProviderConfig(cfg ProvidersConfig) ProvidersConfig {
	if cfg.Providers == nil {
		cfg.Providers = map[string]ProviderConfig{}
	}

	for name, provider := range cfg.Providers {
		if provider.Name == "" {
			provider.Name = name
		}
		cfg.Providers[name] = provider
	}

	return cfg
}
