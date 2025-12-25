package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var cachedConfig *ProvidersConfig

// getProviderConfig loads and caches the provider configuration
func getProviderConfig() (*ProvidersConfig, error) {
	if cachedConfig != nil {
		return cachedConfig, nil
	}

	configPath, err := findConfigPath()
	if err != nil {
		return nil, err
	}

	baseConfig, err := loadConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	userConfigPath := userConfigFilePath()
	if userConfigPath != "" && fileExists(userConfigPath) {
		userConfig, err := loadConfigFile(userConfigPath)
		if err == nil {
			baseConfig = mergeProviderConfig(baseConfig, userConfig)
		}
	}

	normalized := normalizeProviderConfig(baseConfig)
	cachedConfig = &normalized
	return cachedConfig, nil
}

// clearConfigCache clears the cached configuration
func clearConfigCache() {
	cachedConfig = nil
}

// findConfigPath locates the providers.json file
func findConfigPath() (string, error) {
	// Try current directory first
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	localPath := pathJoin(cwd, "providers.json")
	if fileExists(localPath) {
		return localPath, nil
	}

	// Try executable directory
	exePath, err := os.Executable()
	if err == nil {
		exeDir := pathDirname(exePath)
		repoPath := pathJoin(exeDir, "providers.json")
		if fileExists(repoPath) {
			return repoPath, nil
		}
	}

	return "", fmt.Errorf("providers.json not found")
}

// loadConfigFile loads a configuration file
func loadConfigFile(path string) (ProvidersConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ProvidersConfig{}, err
	}

	var config ProvidersConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return ProvidersConfig{}, err
	}

	return config, nil
}

// userConfigFilePath returns the user's config file path
func userConfigFilePath() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home := getHomeDir()
		if home == "" {
			return ""
		}
		xdgConfigHome = pathJoin(home, ".config")
	}
	return pathJoin(xdgConfigHome, "agents", "providers.json")
}

// mergeProviderConfig merges two provider configurations
func mergeProviderConfig(base, override ProvidersConfig) ProvidersConfig {
	merged := ProvidersConfig{
		Sources: SourcesConfig{
			Guidelines: base.Sources.Guidelines,
			Commands:   base.Sources.Commands,
			Skills:     base.Sources.Skills,
		},
		GlobalGuidelines: append([]string{}, base.GlobalGuidelines...),
		Providers:        make(map[string]ProviderConfig),
	}

	// Override sources
	if override.Sources.Guidelines != "" {
		merged.Sources.Guidelines = override.Sources.Guidelines
	}
	if override.Sources.Commands != "" {
		merged.Sources.Commands = override.Sources.Commands
	}
	if override.Sources.Skills != "" {
		merged.Sources.Skills = override.Sources.Skills
	}

	// Merge global guidelines
	if len(override.GlobalGuidelines) > 0 {
		merged.GlobalGuidelines = append(merged.GlobalGuidelines, override.GlobalGuidelines...)
	}

	// Copy base providers
	for name, provider := range base.Providers {
		merged.Providers[name] = provider
	}

	// Merge override providers
	for name, overrideProvider := range override.Providers {
		baseProvider, exists := merged.Providers[name]
		if !exists {
			merged.Providers[name] = overrideProvider
			continue
		}
		merged.Providers[name] = mergeProvider(baseProvider, overrideProvider)
	}

	return merged
}

// mergeProvider merges two provider configurations
func mergeProvider(base, override ProviderConfig) ProviderConfig {
	merged := base

	if override.Name != "" {
		merged.Name = override.Name
	}
	if override.Guidelines != nil {
		merged.Guidelines = mergeFileSpec(base.Guidelines, override.Guidelines)
	}
	if override.Commands != nil {
		merged.Commands = mergeFileSpec(base.Commands, override.Commands)
	}
	if override.Skills != nil {
		merged.Skills = mergeFileSpec(base.Skills, override.Skills)
	}

	return merged
}

// mergeFileSpec merges two file specifications
func mergeFileSpec(base, override *FileSpec) *FileSpec {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	result := &FileSpec{
		File: base.File,
		Dir:  base.Dir,
	}

	if override.File != "" {
		result.File = override.File
	}
	if override.Dir != "" {
		result.Dir = override.Dir
	}

	return result
}

// normalizeProviderConfig normalizes provider names
func normalizeProviderConfig(cfg ProvidersConfig) ProvidersConfig {
	normalized := cfg
	normalized.Providers = make(map[string]ProviderConfig)

	for name, provider := range cfg.Providers {
		if provider.Name == "" {
			provider.Name = name
		}
		normalized.Providers[name] = provider
	}

	if normalized.GlobalGuidelines == nil {
		normalized.GlobalGuidelines = []string{}
	}

	return normalized
}
