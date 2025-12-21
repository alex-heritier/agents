package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed providers.yaml
var defaultConfigFS embed.FS

// GuidelineConfig defines how guideline files work for an agent
type GuidelineConfig struct {
	File        string   `yaml:"file"`
	Dir         string   `yaml:"dir"`
	Source      string   `yaml:"source"`
	GlobalPaths []string `yaml:"global_paths"`
}

// CommandsConfig defines how slash commands work for an agent
type CommandsConfig struct {
	Dir         string   `yaml:"dir"`
	Extension   string   `yaml:"extension"`
	SourceDir   string   `yaml:"source_dir"`
	GlobalPaths []string `yaml:"global_paths"`
}

// Provider defines configuration for a specific AI agent
type Provider struct {
	Name        string          `yaml:"name"`
	DisplayName string          `yaml:"display_name"`
	Guideline   GuidelineConfig `yaml:"guideline"`
	Commands    CommandsConfig  `yaml:"commands"`
}

// Config represents the entire providers configuration
type Config struct {
	Version   string              `yaml:"version"`
	Providers map[string]Provider `yaml:"providers"`
}

var loadedConfig *Config

// LoadConfig loads the providers configuration from embedded file and user overrides
func LoadConfig() (*Config, error) {
	if loadedConfig != nil {
		return loadedConfig, nil
	}

	// Load default embedded config
	defaultData, err := defaultConfigFS.ReadFile("providers.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(defaultData, config); err != nil {
		return nil, fmt.Errorf("failed to parse default config: %w", err)
	}

	// Try to load user config from XDG config directory
	userConfig, err := loadUserConfig()
	if err == nil && userConfig != nil {
		// Merge user config with default config
		config = mergeConfigs(config, userConfig)
	}

	loadedConfig = config
	return config, nil
}

// loadUserConfig loads the user's custom provider configuration
func loadUserConfig() (*Config, error) {
	configPath := getUserConfigPath()
	if configPath == "" {
		return nil, nil
	}

	// Check if user config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read user config: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse user config: %w", err)
	}

	return config, nil
}

// getUserConfigPath returns the path to the user's config file
// Following XDG Base Directory specification
func getUserConfigPath() string {
	// Check XDG_CONFIG_HOME first
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "agents", "providers.yaml")
	}

	// Fall back to ~/.config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(homeDir, ".config", "agents", "providers.yaml")
}

// mergeConfigs merges user config into default config
// User config values override default values
func mergeConfigs(base, override *Config) *Config {
	result := &Config{
		Version:   base.Version,
		Providers: make(map[string]Provider),
	}

	// Copy base providers
	for name, provider := range base.Providers {
		result.Providers[name] = provider
	}

	// Override with user providers
	for name, userProvider := range override.Providers {
		baseProvider, exists := result.Providers[name]
		if exists {
			// Merge provider fields
			merged := mergeProvider(baseProvider, userProvider)
			result.Providers[name] = merged
		} else {
			// Add new provider
			result.Providers[name] = userProvider
		}
	}

	return result
}

// mergeProvider merges two provider configurations
func mergeProvider(base, override Provider) Provider {
	result := base

	if override.DisplayName != "" {
		result.DisplayName = override.DisplayName
	}

	// Merge guideline config
	if override.Guideline.File != "" {
		result.Guideline.File = override.Guideline.File
	}
	if override.Guideline.Dir != "" {
		result.Guideline.Dir = override.Guideline.Dir
	}
	if override.Guideline.Source != "" {
		result.Guideline.Source = override.Guideline.Source
	}
	if len(override.Guideline.GlobalPaths) > 0 {
		result.Guideline.GlobalPaths = override.Guideline.GlobalPaths
	}

	// Merge commands config
	if override.Commands.Dir != "" {
		result.Commands.Dir = override.Commands.Dir
	}
	if override.Commands.Extension != "" {
		result.Commands.Extension = override.Commands.Extension
	}
	if override.Commands.SourceDir != "" {
		result.Commands.SourceDir = override.Commands.SourceDir
	}
	if len(override.Commands.GlobalPaths) > 0 {
		result.Commands.GlobalPaths = override.Commands.GlobalPaths
	}

	return result
}

// GetProviderNames returns a sorted list of all provider names
func GetProviderNames() []string {
	config, err := LoadConfig()
	if err != nil {
		return []string{}
	}

	names := make([]string, 0, len(config.Providers))
	for name := range config.Providers {
		names = append(names, name)
	}
	return names
}

// GetProvider returns a provider by name
func GetProvider(name string) (Provider, bool) {
	config, err := LoadConfig()
	if err != nil {
		return Provider{}, false
	}

	provider, ok := config.Providers[name]
	return provider, ok
}

// expandPath expands ~ to home directory in paths
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return homeDir
	}

	return filepath.Join(homeDir, path[2:])
}

// ExpandProviderPaths expands all ~ paths in a provider's configuration
func ExpandProviderPaths(provider Provider) Provider {
	result := provider

	// Expand guideline global paths
	expandedGuidelines := make([]string, len(provider.Guideline.GlobalPaths))
	for i, path := range provider.Guideline.GlobalPaths {
		expandedGuidelines[i] = expandPath(path)
	}
	result.Guideline.GlobalPaths = expandedGuidelines

	// Expand commands global paths
	expandedCommands := make([]string, len(provider.Commands.GlobalPaths))
	for i, path := range provider.Commands.GlobalPaths {
		expandedCommands[i] = expandPath(path)
	}
	result.Commands.GlobalPaths = expandedCommands

	return result
}
