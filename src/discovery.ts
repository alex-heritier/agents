import { readdir, stat, lstat, readlink } from "node:fs/promises";
import { join, dirname, basename } from "node:path";
import { homedir } from "node:os";
import { SupportedAgents } from "./agents.ts";
import type { GuidelineFile } from "./types.ts";

const ignoreDir = new Set(["node_modules", ".git", "dist", "build", ".cursor"]);

// Check if a file exists and is not a directory
async function fileExists(path: string): Promise<boolean> {
  try {
    const info = await stat(path);
    return !info.isDirectory();
  } catch {
    return false;
  }
}

// Check if a path is a symbolic link
export async function isSymlink(path: string): Promise<boolean> {
  try {
    const info = await lstat(path);
    return info.isSymbolicLink();
  } catch {
    return false;
  }
}

// Recursively walk a directory
async function* walkDir(dir: string, skipCursor: boolean = true): AsyncGenerator<string> {
  try {
    const entries = await readdir(dir, { withFileTypes: true });
    for (const entry of entries) {
      const fullPath = join(dir, entry.name);
      if (entry.isDirectory()) {
        if (ignoreDir.has(entry.name)) {
          // Skip .cursor only when skipCursor is true
          if (entry.name !== ".cursor" || skipCursor) {
            continue;
          }
        }
        yield* walkDir(fullPath, skipCursor);
      } else {
        yield fullPath;
      }
    }
  } catch {
    // Ignore permission errors, etc.
  }
}

// discoverAgents finds all AGENTS.md files recursively from current directory
export async function discoverAgents(): Promise<string[]> {
  const agents: string[] = [];
  const cwd = process.cwd();

  for await (const path of walkDir(cwd)) {
    if (basename(path) === "AGENTS.md") {
      agents.push(path);
    }
  }

  return agents;
}

// Infer agent type from filename
function inferAgentFromFilename(filename: string): string {
  if (filename === "AGENTS.md") {
    return "AGENTS";
  }

  for (const [agentName, cfg] of Object.entries(SupportedAgents)) {
    if (filename === cfg.file) {
      return agentName.toUpperCase();
    }
  }

  return "";
}

// discoverAll finds all guideline files (AGENTS.md, CLAUDE.md, .cursor/rules/*)
export async function discoverAll(): Promise<GuidelineFile[]> {
  const files: GuidelineFile[] = [];
  const cwd = process.cwd();

  for await (const path of walkDir(cwd, false)) {
    const dir = dirname(path);
    const filename = basename(path);

    // Determine agent type
    let agent = "";
    if (filename === "AGENTS.md") {
      agent = "AGENTS";
    } else {
      // Check if this matches any agent configuration
      for (const [agentName, cfg] of Object.entries(SupportedAgents)) {
        if (filename === cfg.file) {
          // Check if it's in the right directory (if specified)
          if (cfg.dir === "") {
            agent = agentName.toUpperCase();
            break;
          } else if (path.includes(cfg.dir)) {
            agent = agentName.toUpperCase();
            break;
          }
        }
      }
      if (!agent) {
        continue;
      }
    }

    const isLink = await isSymlink(path);
    let size = 0;
    try {
      const info = await stat(path);
      size = info.size;
    } catch {
      // Ignore errors
    }

    files.push({
      path,
      dir,
      agent,
      file: filename,
      isSymlink: isLink,
      size,
    });
  }

  return files;
}

// globalGuidelinePaths returns the standard locations for global agent guideline files
function globalGuidelinePaths(): string[] {
  const home = homedir();
  if (!home) {
    return [];
  }

  return [
    join(home, ".claude", "CLAUDE.md"),
    join(home, ".codex", "AGENTS.md"),
    join(home, ".gemini", "GEMINI.md"),
    join(home, ".config", "opencode", "AGENTS.md"),
    join(home, ".config", "amp", "AGENTS.md"),
    join(home, ".config", "AGENTS.md"),
    join(home, "AGENTS.md"),
  ];
}

// discoverGlobalOnly finds only user/system-wide agent guideline files
export async function discoverGlobalOnly(): Promise<GuidelineFile[]> {
  const files: GuidelineFile[] = [];
  const globalLocations = globalGuidelinePaths();

  if (globalLocations.length === 0) {
    return files;
  }

  for (const location of globalLocations) {
    if (await fileExists(location)) {
      let size = 0;
      try {
        const info = await stat(location);
        size = info.size;
      } catch {
        continue;
      }

      const filename = basename(location);
      const dir = dirname(location);

      const agent = inferAgentFromFilename(filename);
      if (!agent) {
        continue;
      }

      const isLink = await isSymlink(location);

      files.push({
        path: location,
        dir,
        agent,
        file: filename,
        isSymlink: isLink,
        size,
      });
    }
  }

  return files;
}
