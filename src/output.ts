import type { ManagedFile, SkillFile } from "./types";
import { formatRelativeDir, getHomeDir } from "./paths";

export function formatList(files: ManagedFile[], verbose: boolean, emptyMessage: string) {
  if (files.length === 0) {
    console.log(emptyMessage);
    return;
  }

  console.log(`${"Directory".padEnd(30)} ${"File".padEnd(15)} ${"Tokens".padEnd(10)}`);
  console.log("-".repeat(55));

  const cwd = process.cwd();
  const homeDir = getHomeDir();

  const standardGlobalPatterns = [
    `${homeDir}/.claude`,
    `${homeDir}/.codex`,
    `${homeDir}/.gemini`,
    `${homeDir}/.config`,
  ];

  const allFilesAreGlobal = files.every((file) =>
    standardGlobalPatterns.some((pattern) => file.dir.startsWith(pattern)),
  );

  for (const file of files) {
    const displayDir = formatRelativeDir(file.dir, cwd, homeDir, allFilesAreGlobal);
    const filename = file.isSymlink ? `*${file.file}` : file.file;
    const tokensStr = file.isSymlink ? "-" : estimateTokens(file.size).toString();
    console.log(`${displayDir.padEnd(30)} ${filename.padEnd(15)} ${tokensStr.padEnd(10)}`);
  }

  if (verbose) {
    console.log(`\nTotal: ${files.length} files found`);
  }
}

export function formatSyncSummary(
  sourceName: string,
  found: number,
  created: number,
  skipped: number,
  verbose: boolean,
  operations: string[],
) {
  console.log(`${sourceName} files found: ${found}`);
  console.log(`Symlinks created: ${created}`);
  console.log(`Symlinks skipped: ${skipped}`);

  if (verbose && operations.length > 0) {
    console.log("\nOperations:");
    for (const op of operations) {
      console.log(op);
    }
  }
}

function estimateTokens(size: number): number {
  return Math.floor(size / 4);
}

export function formatSkillsList(skills: SkillFile[], verbose: boolean) {
  if (skills.length === 0) {
    console.log("No skills found.");
    return;
  }

  console.log(`${"Skill Name".padEnd(25)} ${"Location".padEnd(10)} ${"Status".padEnd(15)} ${"Description"}`);
  console.log("-".repeat(80));

  for (const skill of skills) {
    const status = skill.error ? `Error: ${skill.error}` : "Valid";
    const description = skill.metadata?.description
      ? skill.metadata.description.substring(0, 40) + (skill.metadata.description.length > 40 ? "..." : "")
      : "-";

    console.log(
      `${skill.skillName.padEnd(25)} ${skill.location.padEnd(10)} ${status.padEnd(15)} ${description}`,
    );

    if (verbose && skill.metadata) {
      console.log(`  Name: ${skill.metadata.name}`);
      if (skill.metadata.license) {
        console.log(`  License: ${skill.metadata.license}`);
      }
      if (skill.metadata.allowedTools) {
        console.log(`  Allowed Tools: ${skill.metadata.allowedTools}`);
      }
      console.log(`  Path: ${skill.path}`);
      console.log();
    }
  }

  if (verbose) {
    console.log(`\nTotal: ${skills.length} skills found`);
  }
}

export function formatSkillsSyncSummary(
  found: number,
  created: number,
  skipped: number,
  verbose: boolean,
  operations: string[],
) {
  console.log(`Source skills found: ${found}`);
  console.log(`Skills synced: ${created}`);
  console.log(`Skills skipped: ${skipped}`);

  if (verbose && operations.length > 0) {
    console.log("\nOperations:");
    for (const op of operations) {
      console.log(op);
    }
  }
}
