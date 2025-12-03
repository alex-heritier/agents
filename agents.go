package main

// AgentConfig defines how a guideline file should be created for an agent
type AgentConfig struct {
	Name string // agent identifier (claude, cursor, etc.)
	File string // filename to create
	Dir  string // subdirectory within project dir (empty string = root)
}

// SupportedAgents defines all available agent types and their configurations
var SupportedAgents = map[string]AgentConfig{
	"claude": {
		Name: "claude",
		File: "CLAUDE.md",
		Dir:  "",
	},
	"cursor": {
		Name: "cursor",
		File: "agents.md",
		Dir:  ".cursor/rules",
	},
	"copilot": {
		Name: "copilot",
		File: "COPILOT.md",
		Dir:  "",
	},
}

// GetAgentNames returns a list of all supported agent names
func GetAgentNames() []string {
	names := make([]string, 0, len(SupportedAgents))
	for name := range SupportedAgents {
		names = append(names, name)
	}
	return names
}
