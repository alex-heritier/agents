package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var cachedConfig *ToolsConfig

func getToolConfig() (*ToolsConfig, error) {
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

func findConfigPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	localPath := pathJoin(cwd, "tools.json")
	if fileExists(localPath) {
		return localPath, nil
	}

	exePath, err := os.Executable()
	if err == nil {
		exeDir := pathDirname(exePath)
		repoPath := pathJoin(exeDir, "tools.json")
		if fileExists(repoPath) {
			return repoPath, nil
		}
	}

	return "", fmt.Errorf("tools.json not found")
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
		Sources: SourcesConfig{
			Guidelines: base.Sources.Guidelines,
			Commands:   base.Sources.Commands,
			Skills:     base.Sources.Skills,
		},
		GlobalGuidelines: append([]string{}, base.GlobalGuidelines...),
		Tools:            make(map[string]ToolConfig),
	}

	if override.Sources.Guidelines != "" {
		merged.Sources.Guidelines = override.Sources.Guidelines
	}
	if override.Sources.Commands != "" {
		merged.Sources.Commands = override.Sources.Commands
	}
	if override.Sources.Skills != "" {
		merged.Sources.Skills = override.Sources.Skills
	}

	if len(override.GlobalGuidelines) > 0 {
		merged.GlobalGuidelines = append(merged.GlobalGuidelines, override.GlobalGuidelines...)
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

func normalizeToolConfig(cfg ToolsConfig) ToolsConfig {
	normalized := cfg
	normalized.Tools = make(map[string]ToolConfig)

	for name, tool := range cfg.Tools {
		if tool.Name == "" {
			tool.Name = name
		}
		normalized.Tools[name] = tool
	}

	if normalized.GlobalGuidelines == nil {
		normalized.GlobalGuidelines = []string{}
	}

	return normalized
}
