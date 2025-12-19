// AgentConfig defines how a guideline file should be created for an agent
export interface AgentConfig {
  name: string;  // agent identifier (claude, cursor, etc.)
  file: string;  // filename to create
  dir: string;   // subdirectory within project dir (empty string = root)
}

// SupportedAgents defines all available agent types and their configurations
export const SupportedAgents: Record<string, AgentConfig> = {
  claude: {
    name: "claude",
    file: "CLAUDE.md",
    dir: "",
  },
  cursor: {
    name: "cursor",
    file: "agents.md",
    dir: ".cursor/rules",
  },
  copilot: {
    name: "copilot",
    file: "COPILOT.md",
    dir: "",
  },
  gemini: {
    name: "gemini",
    file: "GEMINI.md",
    dir: "",
  },
  qwen: {
    name: "qwen",
    file: "QWEN.md",
    dir: "",
  },
};

// GetAgentNames returns a list of all supported agent names
export function getAgentNames(): string[] {
  return Object.keys(SupportedAgents);
}
