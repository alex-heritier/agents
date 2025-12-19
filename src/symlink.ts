import { readlink, symlink, unlink, mkdir, readFile, lstat, stat } from "node:fs/promises";
import { join, dirname } from "node:path";
import { createInterface } from "node:readline";
import { SupportedAgents } from "./agents.ts";
import { discoverAll, isSymlink } from "./discovery.ts";
import { formatSyncSummary, formatRmSummary } from "./output.ts";

// Ask for user confirmation
async function askForConfirmation(targetPath: string, reason: string): Promise<boolean> {
  console.log(`\nFile already exists: ${targetPath} (${reason})`);

  const rl = createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  return new Promise((resolve) => {
    rl.question("Overwrite? (y/n): ", (answer) => {
      rl.close();
      const response = answer.trim().toLowerCase();
      resolve(response === "y" || response === "yes");
    });
  });
}

// Check if we should skip or overwrite an existing file
async function shouldSkipOrOverwrite(
  targetPath: string,
  expectedTarget: string,
  sourcePath: string,
  dryRun: boolean
): Promise<boolean> {
  const info = await lstat(targetPath);

  // Check if it's a symlink
  if (info.isSymbolicLink()) {
    try {
      const link = await readlink(targetPath);
      if (link === expectedTarget) {
        return true; // Symlink is correct
      }
    } catch {
      // Can't read link
    }
    // Symlink points to wrong place, ask to overwrite
    if (!dryRun && !(await askForConfirmation(targetPath, expectedTarget))) {
      return true; // User said no, skip it
    }
    return false; // Overwrite
  }

  // It's a regular file, compare content with source
  try {
    const sourceContent = await readFile(sourcePath, "utf-8");
    const targetContent = await readFile(targetPath, "utf-8");

    if (sourceContent === targetContent) {
      return true; // Content matches, skip
    }
  } catch {
    // Can't read files, ask user
    if (!dryRun && !(await askForConfirmation(targetPath, "source"))) {
      return true;
    }
    return false;
  }

  // Content differs, ask user
  if (!dryRun && !(await askForConfirmation(targetPath, "different version"))) {
    return true;
  }
  return false;
}

async function createSymlinkFile(
  dir: string,
  source: string,
  target: string,
  dryRun: boolean,
  verbose: boolean
): Promise<{ created: number; skipped: number; operation: string }> {
  const sourcePath = join(dir, source);
  const targetPath = join(dir, target);

  // Check if target already exists
  try {
    await lstat(targetPath);
    // File exists, check if it's the right symlink or if we should overwrite
    if (await shouldSkipOrOverwrite(targetPath, source, sourcePath, dryRun)) {
      if (verbose) {
        return { created: 0, skipped: 1, operation: `skipped: ${targetPath} (already correct)` };
      }
      return { created: 0, skipped: 1, operation: "" };
    }

    // User wants to overwrite
    if (!dryRun) {
      await unlink(targetPath);
    }
  } catch {
    // File doesn't exist, proceed
  }

  if (dryRun) {
    if (verbose) {
      return { created: 1, skipped: 0, operation: `would create: ${targetPath} -> ${source}` };
    }
    return { created: 1, skipped: 0, operation: "" };
  }

  // Create relative symlink
  try {
    await symlink(source, targetPath);
    if (verbose) {
      return { created: 1, skipped: 0, operation: `created: ${targetPath} -> ${source}` };
    }
    return { created: 1, skipped: 0, operation: "" };
  } catch (err) {
    if (verbose) {
      return { created: 0, skipped: 1, operation: `error: ${targetPath} (${err})` };
    }
    return { created: 0, skipped: 1, operation: "" };
  }
}

async function createSymlinkInDir(
  dir: string,
  source: string,
  subdir: string,
  target: string,
  dryRun: boolean,
  verbose: boolean
): Promise<{ created: number; skipped: number; operation: string }> {
  // Create subdirectory if needed
  const subdirPath = join(dir, subdir);
  try {
    await stat(subdirPath);
  } catch {
    if (!dryRun) {
      await mkdir(subdirPath, { recursive: true });
    }
  }

  const sourcePath = join(dir, source);
  const targetPath = join(subdirPath, target);
  const symTarget = join("..", "..", source);

  // Check if target already exists
  try {
    await lstat(targetPath);
    // File exists, check if it's the right symlink or if we should overwrite
    if (await shouldSkipOrOverwrite(targetPath, symTarget, sourcePath, dryRun)) {
      if (verbose) {
        return { created: 0, skipped: 1, operation: `skipped: ${targetPath} (already correct)` };
      }
      return { created: 0, skipped: 1, operation: "" };
    }

    // User wants to overwrite
    if (!dryRun) {
      await unlink(targetPath);
    }
  } catch {
    // File doesn't exist, proceed
  }

  if (dryRun) {
    if (verbose) {
      return { created: 1, skipped: 0, operation: `would create: ${targetPath} -> ${symTarget}` };
    }
    return { created: 1, skipped: 0, operation: "" };
  }

  // Create relative symlink back to AGENTS.md
  try {
    await symlink(symTarget, targetPath);
    if (verbose) {
      return { created: 1, skipped: 0, operation: `created: ${targetPath} -> ${symTarget}` };
    }
    return { created: 1, skipped: 0, operation: "" };
  } catch (err) {
    if (verbose) {
      return { created: 0, skipped: 1, operation: `error: ${targetPath} (${err})` };
    }
    return { created: 0, skipped: 1, operation: "" };
  }
}

export async function syncSymlinks(
  agents: string[],
  selectedAgents: string[],
  dryRun: boolean,
  verbose: boolean
): Promise<void> {
  let created = 0;
  let skipped = 0;
  const operations: string[] = [];

  for (const agentPath of agents) {
    const dir = dirname(agentPath);
    const filename = "AGENTS.md";

    for (const agentName of selectedAgents) {
      const cfg = SupportedAgents[agentName];
      if (!cfg) continue;

      let result: { created: number; skipped: number; operation: string };

      if (cfg.dir === "") {
        // Create symlink in same directory
        result = await createSymlinkFile(dir, filename, cfg.file, dryRun, verbose);
      } else {
        // Create symlink in subdirectory
        result = await createSymlinkInDir(dir, filename, cfg.dir, cfg.file, dryRun, verbose);
      }

      created += result.created;
      skipped += result.skipped;
      if (result.operation) {
        operations.push(result.operation);
      }
    }
  }

  formatSyncSummary(agents.length, created, skipped, verbose, operations);
}

export async function deleteGuidelineFiles(
  selectedAgents: string[],
  dryRun: boolean,
  verbose: boolean
): Promise<void> {
  let deleted = 0;
  let notFound = 0;
  const operations: string[] = [];

  const allFiles = await discoverAll();

  for (const file of allFiles) {
    // Check if this file matches any of the selected agents
    const shouldDelete = selectedAgents.some(
      (agentName) => file.agent.toLowerCase() === agentName.toLowerCase()
    );

    if (!shouldDelete) {
      continue;
    }

    if (dryRun) {
      deleted++;
      if (verbose) {
        operations.push(`would delete: ${file.path}`);
      }
    } else {
      try {
        await unlink(file.path);
        deleted++;
        if (verbose) {
          operations.push(`deleted: ${file.path}`);
        }
      } catch (err) {
        notFound++;
        if (verbose) {
          operations.push(`error: ${file.path} (${err})`);
        }
      }
    }
  }

  formatRmSummary(deleted, notFound, verbose, operations);
}
