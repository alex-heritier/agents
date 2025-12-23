import { existsSync, lstatSync, readdirSync, readFileSync, statSync } from "node:fs";
import type { FileSpec, ManagedFile, ProviderConfig, ProvidersConfig, SkillFile, SkillMetadata } from "./types";
import { expandHomePath, getHomeDir, pathBasename, pathDirname, pathJoin } from "./paths";

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

    const dir = pathDirname(path);
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
    const filename = pathBasename(location);
    const dir = pathDirname(location);
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

function walk(
  root: string,
  visitor: (path: string, entry: { name: string; isDirectory: () => boolean; isFile: () => boolean }) => "skip" | "continue",
) {
  const entries = readdirSync(root, { withFileTypes: true });
  for (const entry of entries) {
    const fullPath = pathJoin(root, entry.name);
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

// Skill discovery functions
export function discoverSkills(): SkillFile[] {
  const skills: SkillFile[] = [];

  // Discover global skills (~/.claude/skills/)
  const globalSkillsDir = pathJoin(getHomeDir(), ".claude", "skills");
  if (existsSync(globalSkillsDir)) {
    const globalSkills = discoverSkillsInDir(globalSkillsDir, "global");
    skills.push(...globalSkills);
  }

  // Discover project skills (.claude/skills/)
  const projectSkillsDir = pathJoin(process.cwd(), ".claude", "skills");
  if (existsSync(projectSkillsDir)) {
    const projectSkills = discoverSkillsInDir(projectSkillsDir, "project");
    skills.push(...projectSkills);
  }

  return skills;
}

export function discoverSourceSkills(sourceDirName: string): string[] {
  const cwd = process.cwd();
  const skillDirs: string[] = [];

  walk(cwd, (path, entry) => {
    if (entry.isDirectory() && ignoreDir.has(entry.name)) {
      return "skip";
    }
    // Skip .claude directory to avoid treating .claude/skills as a source
    if (entry.isDirectory() && entry.name === ".claude") {
      return "skip";
    }
    if (entry.isDirectory() && entry.name === sourceDirName) {
      // Found a source skills directory, scan for skill subdirectories
      const subdirs = readdirSync(path, { withFileTypes: true });
      for (const subdir of subdirs) {
        if (subdir.isDirectory()) {
          const skillFile = pathJoin(path, subdir.name, "SKILL.md");
          if (existsSync(skillFile)) {
            skillDirs.push(pathJoin(path, subdir.name));
          }
        }
      }
      return "skip";
    }
    return "continue";
  });

  return skillDirs;
}

function discoverSkillsInDir(dir: string, location: "global" | "project"): SkillFile[] {
  const skills: SkillFile[] = [];

  try {
    const entries = readdirSync(dir, { withFileTypes: true });
    for (const entry of entries) {
      if (!entry.isDirectory()) {
        continue;
      }

      const skillName = entry.name;
      const skillPath = pathJoin(dir, skillName);
      const skillFile = pathJoin(skillPath, "SKILL.md");

      if (!existsSync(skillFile)) {
        continue;
      }

      const { metadata, error } = parseSkillMetadata(skillFile);

      skills.push({
        path: skillFile,
        dir: skillPath,
        skillName,
        location,
        metadata,
        error,
      });
    }
  } catch (err) {
    // Directory doesn't exist or can't be read
  }

  return skills;
}

function parseSkillMetadata(skillFile: string): { metadata?: SkillMetadata; error?: string } {
  try {
    const content = readFileSync(skillFile, "utf-8");
    const frontmatterMatch = content.match(/^---\s*\n([\s\S]*?)\n---/);

    if (!frontmatterMatch) {
      return { error: "No frontmatter found" };
    }

    const frontmatter = frontmatterMatch[1];
    const metadata: Partial<SkillMetadata> = {};

    for (const line of frontmatter.split("\n")) {
      const match = line.match(/^(\w+(?:-\w+)*):\s*(.+)$/);
      if (!match) {
        continue;
      }

      const [, key, value] = match;
      const trimmedValue = value.trim();

      if (key === "name") {
        metadata.name = trimmedValue;
      } else if (key === "description") {
        metadata.description = trimmedValue;
      } else if (key === "license") {
        metadata.license = trimmedValue;
      } else if (key === "allowed-tools") {
        metadata.allowedTools = trimmedValue;
      }
    }

    if (!metadata.name || !metadata.description) {
      return { error: "Missing required fields (name, description)" };
    }

    return { metadata: metadata as SkillMetadata };
  } catch (err) {
    return { error: `Failed to parse: ${err}` };
  }
}
