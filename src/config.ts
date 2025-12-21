import { readFile } from 'fs/promises';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import type { Config, Provider } from './types.js';

// Simple YAML parser for our config format
function parseYAML(content: string): any {
  const lines = content.split('\n');
  const result: any = {};
  const stack: any[] = [{ obj: result, indent: -2 }];

  for (let line of lines) {
    // Skip comments and empty lines
    if (line.trim().startsWith('#') || line.trim() === '') continue;

    const indent = line.search(/\S/);
    if (indent === -1) continue;

    const content = line.trim();

    // Handle key-value pairs
    if (content.includes(':')) {
      const [key, ...valueParts] = content.split(':');
      const value = valueParts.join(':').trim();

      // Pop stack until we find the right parent
      while (stack.length > 1 && indent <= stack[stack.length - 1].indent) {
        stack.pop();
      }

      const parent = stack[stack.length - 1].obj;

      if (value === '') {
        // This is a parent key
        const newObj = {};
        parent[key.trim()] = newObj;
        stack.push({ obj: newObj, indent });
      } else if (value.startsWith('"') || value.startsWith("'")) {
        // String value
        parent[key.trim()] = value.slice(1, -1);
      } else {
        // Simple value
        parent[key.trim()] = value;
      }
    } else if (content.startsWith('- ')) {
      // Array item
      const parent = stack[stack.length - 1].obj;
      const lastKey = Object.keys(parent).pop();
      if (lastKey && !Array.isArray(parent[lastKey])) {
        parent[lastKey] = [];
      }
      if (lastKey) {
        const value = content.slice(2).trim();
        parent[lastKey].push(value.startsWith('"') || value.startsWith("'") ? value.slice(1, -1) : value);
      }
    }
  }

  return result;
}

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

let loadedConfig: Config | null = null;

/**
 * Load the providers configuration from embedded file and user overrides
 */
export async function loadConfig(): Promise<Config> {
  if (loadedConfig) {
    return loadedConfig;
  }

  // Load default embedded config
  const defaultConfigPath = join(__dirname, '..', 'providers.yaml');
  const defaultData = await readFile(defaultConfigPath, 'utf-8');
  const config: Config = parseYAML(defaultData);

  // Try to load user config from XDG config directory
  const userConfig = await loadUserConfig();
  if (userConfig) {
    // Merge user config with default config
    mergeConfigs(config, userConfig);
  }

  loadedConfig = config;
  return config;
}

/**
 * Load the user's custom provider configuration
 */
async function loadUserConfig(): Promise<Config | null> {
  const configPath = getUserConfigPath();
  if (!configPath) {
    return null;
  }

  try {
    const data = await readFile(configPath, 'utf-8');
    const config: Config = parseYAML(data);
    return config;
  } catch (err) {
    // File doesn't exist or can't be read
    return null;
  }
}

/**
 * Get the path to the user's config file
 * Following XDG Base Directory specification
 */
function getUserConfigPath(): string | null {
  // Check XDG_CONFIG_HOME first
  const xdgConfigHome = process.env.XDG_CONFIG_HOME;
  if (xdgConfigHome) {
    return join(xdgConfigHome, 'agents', 'providers.yaml');
  }

  // Fall back to ~/.config
  const homeDir = process.env.HOME;
  if (!homeDir) {
    return null;
  }

  return join(homeDir, '.config', 'agents', 'providers.yaml');
}

/**
 * Merge user config into base config
 * User config values override default values
 */
function mergeConfigs(base: Config, override: Config): void {
  // Override with user providers
  for (const [name, userProvider] of Object.entries(override.providers)) {
    const baseProvider = base.providers[name];
    if (baseProvider) {
      // Merge provider fields
      base.providers[name] = mergeProvider(baseProvider, userProvider);
    } else {
      // Add new provider
      base.providers[name] = userProvider;
    }
  }
}

/**
 * Merge two provider configurations
 */
function mergeProvider(base: Provider, override: Provider): Provider {
  const result = { ...base };

  if (override.display_name) {
    result.display_name = override.display_name;
  }

  // Merge guideline config
  if (override.guideline.file) {
    result.guideline.file = override.guideline.file;
  }
  if (override.guideline.dir) {
    result.guideline.dir = override.guideline.dir;
  }
  if (override.guideline.source) {
    result.guideline.source = override.guideline.source;
  }
  if (override.guideline.global_paths.length > 0) {
    result.guideline.global_paths = override.guideline.global_paths;
  }

  // Merge commands config
  if (override.commands.dir) {
    result.commands.dir = override.commands.dir;
  }
  if (override.commands.extension) {
    result.commands.extension = override.commands.extension;
  }
  if (override.commands.source_dir) {
    result.commands.source_dir = override.commands.source_dir;
  }
  if (override.commands.global_paths.length > 0) {
    result.commands.global_paths = override.commands.global_paths;
  }

  return result;
}

/**
 * Get provider names from config
 */
export async function getProviderNames(): Promise<string[]> {
  const config = await loadConfig();
  return Object.keys(config.providers);
}

/**
 * Get a provider by name
 */
export async function getProvider(name: string): Promise<Provider | null> {
  const config = await loadConfig();
  return config.providers[name] || null;
}

/**
 * Expand ~ to home directory in paths
 */
export function expandPath(path: string): string {
  if (!path.startsWith('~')) {
    return path;
  }

  const homeDir = process.env.HOME;
  if (!homeDir) {
    return path;
  }

  if (path === '~') {
    return homeDir;
  }

  return join(homeDir, path.slice(2));
}

/**
 * Expand all ~ paths in a provider's configuration
 */
export function expandProviderPaths(provider: Provider): Provider {
  const result = { ...provider };

  // Expand guideline global paths
  result.guideline.global_paths = provider.guideline.global_paths.map(expandPath);

  // Expand commands global paths
  result.commands.global_paths = provider.commands.global_paths.map(expandPath);

  return result;
}
