import { readdir, stat, lstat } from 'fs/promises';
import { join, dirname, basename } from 'path';
import { getSupportedAgents } from './agents.js';
import { loadConfig } from './config.js';
import type { GuidelineFile } from './types.js';

const ignoreDir = new Set(['node_modules', '.git', 'dist', 'build', '.cursor']);

/**
 * Discover all AGENTS.md files recursively from current directory
 */
export async function discoverAgents(): Promise<string[]> {
  const agents: string[] = [];
  const cwd = process.cwd();
  await walkDirectory(cwd, async (path, stats) => {
    if (stats.isFile() && basename(path) === 'AGENTS.md') {
      agents.push(path);
    }
  });
  return agents;
}

/**
 * Discover all guideline files (AGENTS.md, CLAUDE.md, .cursor/rules/*)
 */
export async function discoverAll(): Promise<GuidelineFile[]> {
  const files: GuidelineFile[] = [];
  const cwd = process.cwd();
  const supportedAgents = await getSupportedAgents();

  await walkDirectory(
    cwd,
    async (path, stats) => {
      if (stats.isFile()) {
        const dir = dirname(path);
        const filename = basename(path);

        // Determine agent type
        let agent = '';
        if (filename === 'AGENTS.md') {
          agent = 'AGENTS';
        } else {
          // Check if this matches any agent configuration
          for (const [agentName, cfg] of Object.entries(supportedAgents)) {
            if (filename === cfg.file) {
              // Check if it's in the right directory (if specified)
              if (cfg.dir === '' || path.includes(cfg.dir)) {
                agent = agentName.toUpperCase();
                break;
              }
            }
          }
        }

        if (agent) {
          const isSymlink = await isSymlinkFile(path);
          files.push({
            path,
            dir,
            agent,
            file: filename,
            isSymlink,
            size: stats.size,
          });
        }
      }
    },
    true // Allow .cursor directory
  );

  return files;
}

/**
 * Discover only user/system-wide agent guideline files
 */
export async function discoverGlobalOnly(): Promise<GuidelineFile[]> {
  const files: GuidelineFile[] = [];
  const globalLocations = globalGuidelinePaths();

  for (const location of globalLocations) {
    if (await fileExists(location)) {
      try {
        const stats = await stat(location);
        const filename = basename(location);
        const dir = dirname(location);
        const agent = inferAgentFromFilename(filename);

        if (agent) {
          const isSymlink = await isSymlinkFile(location);
          files.push({
            path: location,
            dir,
            agent,
            file: filename,
            isSymlink,
            size: stats.size,
          });
        }
      } catch {
        // Skip files that can't be read
      }
    }
  }

  return files;
}

/**
 * Get standard locations for global agent guideline files
 */
function globalGuidelinePaths(): string[] {
  const homeDir = process.env.HOME;
  if (!homeDir) {
    return [];
  }

  return [
    join(homeDir, '.claude', 'CLAUDE.md'),
    join(homeDir, '.codex', 'AGENTS.md'),
    join(homeDir, '.gemini', 'GEMINI.md'),
    join(homeDir, '.config', 'opencode', 'AGENTS.md'),
    join(homeDir, '.config', 'amp', 'AGENTS.md'),
    join(homeDir, '.config', 'AGENTS.md'),
    join(homeDir, 'AGENTS.md'),
  ];
}

/**
 * Infer agent type from filename
 */
async function inferAgentFromFilename(filename: string): Promise<string> {
  if (filename === 'AGENTS.md') {
    return 'AGENTS';
  }

  const supportedAgents = await getSupportedAgents();
  for (const [agentName, cfg] of Object.entries(supportedAgents)) {
    if (filename === cfg.file) {
      return agentName.toUpperCase();
    }
  }

  return '';
}

/**
 * Check if a file exists
 */
export async function fileExists(path: string): Promise<boolean> {
  try {
    const stats = await stat(path);
    return stats.isFile();
  } catch {
    return false;
  }
}

/**
 * Check if a directory exists
 */
export async function dirExists(path: string): Promise<boolean> {
  try {
    const stats = await stat(path);
    return stats.isDirectory();
  } catch {
    return false;
  }
}

/**
 * Check if a path is a symlink
 */
export async function isSymlinkFile(path: string): Promise<boolean> {
  try {
    const stats = await lstat(path);
    return stats.isSymbolicLink();
  } catch {
    return false;
  }
}

/**
 * Walk a directory tree recursively
 */
async function walkDirectory(
  dir: string,
  callback: (path: string, stats: any) => Promise<void>,
  allowCursor = false
): Promise<void> {
  try {
    const entries = await readdir(dir, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = join(dir, entry.name);

      // Skip ignored directories
      if (entry.isDirectory()) {
        if (ignoreDir.has(entry.name) && !(allowCursor && entry.name === '.cursor')) {
          continue;
        }
        await walkDirectory(fullPath, callback, allowCursor);
      } else {
        const stats = await stat(fullPath);
        await callback(fullPath, stats);
      }
    }
  } catch {
    // Silently skip directories we can't read
  }
}
