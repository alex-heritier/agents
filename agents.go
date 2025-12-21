package main

// AgentConfig defines how a guideline file should be created for an agent
// This is now loaded from providers.yaml
type AgentConfig struct {
	Name string // agent identifier (claude, cursor, etc.)
	File string // filename to create
	Dir  string // subdirectory within project dir (empty string = root)
}

// SupportedAgents returns all available agent types and their configurations
// Loaded from providers.yaml and user overrides
var SupportedAgents map[string]AgentConfig

// initSupportedAgents initializes the SupportedAgents map from config
func initSupportedAgents() {
	if SupportedAgents != nil {
		return
	}

	config, err := LoadConfig()
	if err != nil {
		// Fall back to empty map if config fails to load
		SupportedAgents = make(map[string]AgentConfig)
		return
	}

	SupportedAgents = make(map[string]AgentConfig, len(config.Providers))
	for name, provider := range config.Providers {
		SupportedAgents[name] = AgentConfig{
			Name: provider.Name,
			File: provider.Guideline.File,
			Dir:  provider.Guideline.Dir,
		}
	}
}

// GetAgentNames returns a list of all supported agent names
func GetAgentNames() []string {
	initSupportedAgents()
	names := make([]string, 0, len(SupportedAgents))
	for name := range SupportedAgents {
		names = append(names, name)
	}
	return names
}
