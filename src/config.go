package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed tools.json
var defaultToolsConfig []byte

var cachedConfig *ToolsConfig

func getToolConfig() (*ToolsConfig, error) {
	if cachedConfig != nil {
		return cachedConfig, nil
	}

	configPath := findConfigPath()
	var baseConfig ToolsConfig
	var err error

	if configPath != "" {
		baseConfig, err = loadConfigFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", configPath, err)
		}
	} else {
		baseConfig, err = loadDefaultConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load default embedded config: %w", err)
		}
	}

	userConfigPath := userConfigFilePath()
	if userConfigPath != "" && fileExists(userConfigPath) {
		userConfig, err := loadConfigFile(userConfigPath)
		if err == nil {
			baseConfig = mergeToolConfig(baseConfig, userConfig)
		}
	}

	normalized := normalizeToolConfig(baseConfig)
	cachedConfig = &normalized
	return cachedConfig, nil
}

// clearConfigCache clears the cached configuration
func clearConfigCache() {
	cachedConfig = nil
}

func findConfigPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	localPath := pathJoin(cwd, "tools.json")
	if fileExists(localPath) {
		return localPath
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := pathDirname(exePath)
		repoPath := pathJoin(exeDir, "tools.json")
		if fileExists(repoPath) {
			return repoPath
		}
	}

	return ""
}

func loadConfigFile(path string) (ToolsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ToolsConfig{}, err
	}

	var config ToolsConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return ToolsConfig{}, err
	}

	return config, nil
}

func loadDefaultConfig() (ToolsConfig, error) {
	var config ToolsConfig
	err := json.Unmarshal(defaultToolsConfig, &config)
	if err != nil {
		return ToolsConfig{}, err
	}
	return config, nil
}

func userConfigFilePath() string {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home := getHomeDir()
		if home == "" {
			return ""
		}
		xdgConfigHome = pathJoin(home, ".config")
	}
	return pathJoin(xdgConfigHome, "agents", "tools.json")
}

func mergeToolConfig(base, override ToolsConfig) ToolsConfig {
	merged := ToolsConfig{
		Standard: base.Standard,
		Tools:    make(map[string]ToolConfig),
	}

	if override.Standard != "" {
		merged.Standard = override.Standard
	}

	for name, tool := range base.Tools {
		merged.Tools[name] = tool
	}

	for name, overrideTool := range override.Tools {
		baseTool, exists := merged.Tools[name]
		if !exists {
			merged.Tools[name] = overrideTool
			continue
		}
		merged.Tools[name] = mergeTool(baseTool, overrideTool)
	}

	return merged
}

func mergeTool(base, override ToolConfig) ToolConfig {
	merged := base

	if override.Name != "" {
		merged.Name = override.Name
	}

	if override.Pattern != "" {
		merged.Pattern = override.Pattern
	}

	if len(override.Global) > 0 {
		merged.Global = append(merged.Global, override.Global...)
	}

	return merged
}

func normalizeToolConfig(cfg ToolsConfig) ToolsConfig {

	normalized := cfg

	normalized.Tools = make(map[string]ToolConfig)

	for name, tool := range cfg.Tools {

		if tool.Name == "" {

			tool.Name = name

		}

		if tool.Global == nil {

			tool.Global = []string{}

		}

		normalized.Tools[name] = tool

	}

	return normalized

}
