import { lstatSync, readdirSync, statSync } from "node:fs";
import { homedir } from "node:os";
import { basename, dirname, join, relative } from "node:path";
import type { FileSpec, ManagedFile, ProviderConfig, ProvidersConfig } from "./types";

const ignoreDir = new Set(["node_modules", ".git", "dist", "build", ".cursor"]);

export function discoverSources(sourceName: string): string[] {
  const cwd = process.cwd();
  const sources: string[] = [];
  walk(cwd, (path, entry) => {
    if (entry.isDirectory() && ignoreDir.has(entry.name)) {
      return "skip";
    }
    if (entry.isFile() && entry.name === sourceName) {
      sources.push(path);
    }
    return "continue";
  });
  return sources;
}

export function discoverAll(
  cfg: ProvidersConfig,
  sourceName: string,
  specSelector: (provider: ProviderConfig) => FileSpec | undefined,
): ManagedFile[] {
  const cwd = process.cwd();
  const allowedDirs = allowedProviderDirs(cfg, specSelector);
  const files: ManagedFile[] = [];

  walk(cwd, (path, entry) => {
    if (entry.isDirectory()) {
      if (ignoreDir.has(entry.name) && !allowedDirs.has(entry.name)) {
        return "skip";
      }
      return "continue";
    }

    const dir = dirname(path);
    const filename = entry.name;

    let agent = "";
    if (filename === sourceName) {
      agent = sourceName.toUpperCase().replace(/\.MD$/, "");
    } else {
      for (const [agentName, provider] of Object.entries(cfg.providers)) {
        const spec = specSelector(provider);
        if (!spec) {
          continue;
        }
        if (filename === spec.file) {
          if (!spec.dir) {
            agent = agentName.toUpperCase();
            break;
          }
          if (path.includes(spec.dir)) {
            agent = agentName.toUpperCase();
            break;
          }
        }
      }
      if (!agent) {
        return "continue";
      }
    }

    const stat = lstatSync(path);
    files.push({
      path,
      dir,
      agent,
      file: filename,
      isSymlink: stat.isSymbolicLink(),
      size: stat.size,
    });
    return "continue";
  });

  return files;
}

export function discoverGlobalOnly(cfg: ProvidersConfig): ManagedFile[] {
  const files: ManagedFile[] = [];
  for (const location of globalGuidelinePaths(cfg)) {
    if (!fileExists(location)) {
      continue;
    }
    const info = statSync(location);
    const filename = basename(location);
    const dir = dirname(location);
    const agent = inferProviderFromFilename(cfg, filename);
    if (!agent) {
      continue;
    }
    const isSymlink = lstatSync(location).isSymbolicLink();
    files.push({
      path: location,
      dir,
      agent,
      file: filename,
      isSymlink,
      size: info.size,
    });
  }
  return files;
}

export function inferProviderFromFilename(cfg: ProvidersConfig, filename: string): string {
  if (filename === cfg.sources.guidelines) {
    return filename.toUpperCase().replace(/\.MD$/, "");
  }
  for (const [agentName, provider] of Object.entries(cfg.providers)) {
    const spec = provider.guidelines;
    if (spec && filename === spec.file) {
      return agentName.toUpperCase();
    }
  }
  return "";
}

export function expandHomePath(path: string): string {
  if (!path.startsWith("~")) {
    return path;
  }
  return join(homedir(), path.replace(/^~\/?/, ""));
}

export function fileExists(path: string): boolean {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
}

export function isSymlink(path: string): boolean {
  try {
    return lstatSync(path).isSymbolicLink();
  } catch {
    return false;
  }
}

export function formatRelativeDir(targetDir: string, cwd: string, homeDir: string, showRelativeToHome: boolean): string {
  if (showRelativeToHome && homeDir && targetDir.startsWith(homeDir)) {
    const relDir = relative(homeDir, targetDir);
    return relDir === "" ? "~/" : `~/${relDir}`;
  }
  const relDir = relative(cwd, targetDir);
  if (!relDir) {
    return "./";
  }
  if (!relDir.startsWith(".")) {
    return `./${relDir}`;
  }
  return `${relDir}/`;
}

function walk(
  root: string,
  visitor: (path: string, entry: { name: string; isDirectory: () => boolean; isFile: () => boolean }) => "skip" | "continue",
) {
  const entries = readdirSync(root, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = join(root, entry.name);
    const decision = visitor(fullPath, entry);
    if (decision === "skip") {
      continue;
    }
    if (entry.isDirectory()) {
      walk(fullPath, visitor);
    }
  }
}

function allowedProviderDirs(
  cfg: ProvidersConfig,
  specSelector: (provider: ProviderConfig) => FileSpec | undefined,
): Set<string> {
  const allowed = new Set<string>();
  for (const provider of Object.values(cfg.providers)) {
    const spec = specSelector(provider);
    if (!spec || !spec.dir) {
      continue;
    }
    const parts = spec.dir.split("/");
    if (parts.length > 0) {
      allowed.add(parts[0]);
    }
  }
  return allowed;
}

function globalGuidelinePaths(cfg: ProvidersConfig): string[] {
  return (cfg.globalGuidelines ?? []).map((path) => expandHomePath(path));
}
