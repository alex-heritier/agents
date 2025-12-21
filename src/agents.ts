import { loadConfig } from './config.js';
import type { AgentConfig } from './types.js';

let supportedAgents: Record<string, AgentConfig> | null = null;

/**
 * Initialize the SupportedAgents map from config
 */
async function initSupportedAgents(): Promise<Record<string, AgentConfig>> {
  if (supportedAgents) {
    return supportedAgents;
  }

  const config = await loadConfig();
  supportedAgents = {};

  for (const [name, provider] of Object.entries(config.providers)) {
    supportedAgents[name] = {
      name: provider.name,
      file: provider.guideline.file,
      dir: provider.guideline.dir,
    };
  }

  return supportedAgents;
}

/**
 * Get all supported agent configurations
 */
export async function getSupportedAgents(): Promise<Record<string, AgentConfig>> {
  return await initSupportedAgents();
}

/**
 * Get a list of all supported agent names
 */
export async function getAgentNames(): Promise<string[]> {
  const agents = await initSupportedAgents();
  return Object.keys(agents);
}
