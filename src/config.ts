import { existsSync, readFileSync } from "node:fs";
import { homedir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import type { FileSpec, ProviderConfig, ProvidersConfig } from "./types";

let cachedConfig: ProvidersConfig | null = null;

export function getProviderConfig(): ProvidersConfig {
  if (cachedConfig) {
    return cachedConfig;
  }
  const baseConfig = loadConfigFile(findConfigPath());
  const userConfigPath = userConfigFilePath();
  const mergedConfig = userConfigPath && existsSync(userConfigPath)
    ? mergeProviderConfig(baseConfig, loadConfigFile(userConfigPath))
    : baseConfig;
  cachedConfig = normalizeProviderConfig(mergedConfig);
  return cachedConfig;
}

function findConfigPath(): string {
  const localPath = join(process.cwd(), "providers.json");
  if (existsSync(localPath)) {
    return localPath;
  }
  const moduleDir = dirname(fileURLToPath(import.meta.url));
  const repoPath = join(moduleDir, "..", "providers.json");
  if (existsSync(repoPath)) {
    return repoPath;
  }
  throw new Error("providers.json not found");
}

function loadConfigFile(path: string): ProvidersConfig {
  const raw = readFileSync(path, "utf-8");
  return JSON.parse(raw) as ProvidersConfig;
}

function userConfigFilePath(): string | null {
  const xdgConfigHome = process.env.XDG_CONFIG_HOME
    ? process.env.XDG_CONFIG_HOME
    : join(homedir(), ".config");
  return join(xdgConfigHome, "agents", "providers.json");
}

function mergeProviderConfig(base: ProvidersConfig, override: ProvidersConfig): ProvidersConfig {
  const merged: ProvidersConfig = {
    sources: {
      guidelines: override.sources?.guidelines || base.sources.guidelines,
      commands: override.sources?.commands || base.sources.commands,
    },
    globalGuidelines: base.globalGuidelines.slice(),
    providers: { ...base.providers },
  };

  if (override.globalGuidelines && override.globalGuidelines.length > 0) {
    merged.globalGuidelines = merged.globalGuidelines.concat(override.globalGuidelines);
  }

  if (override.providers) {
    for (const [name, overrideProvider] of Object.entries(override.providers)) {
      const baseProvider = merged.providers[name];
      if (!baseProvider) {
        merged.providers[name] = overrideProvider;
        continue;
      }
      merged.providers[name] = mergeProvider(baseProvider, overrideProvider);
    }
  }

  return merged;
}

function mergeProvider(base: ProviderConfig, override: ProviderConfig): ProviderConfig {
  const merged: ProviderConfig = { ...base };
  if (override.name) {
    merged.name = override.name;
  }
  if (override.guidelines) {
    merged.guidelines = mergeFileSpec(base.guidelines, override.guidelines);
  }
  if (override.commands) {
    merged.commands = mergeFileSpec(base.commands, override.commands);
  }
  return merged;
}

function mergeFileSpec(base: FileSpec | undefined, override: FileSpec): FileSpec {
  return {
    file: override.file || base?.file || "",
    dir: override.dir || base?.dir || "",
  };
}

function normalizeProviderConfig(cfg: ProvidersConfig): ProvidersConfig {
  const providers: Record<string, ProviderConfig> = { ...cfg.providers };
  for (const [name, provider] of Object.entries(providers)) {
    providers[name] = { ...provider, name: provider.name || name };
  }
  return {
    sources: cfg.sources,
    globalGuidelines: cfg.globalGuidelines ?? [],
    providers,
  };
}
