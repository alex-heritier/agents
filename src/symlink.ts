import { readFile, unlink, symlink, mkdir, readlink, lstat } from 'fs/promises';
import { join, relative, dirname } from 'path';
import { createInterface } from 'readline';
import { getSupportedAgents } from './agents.js';
import { discoverAll } from './discovery.js';

/**
 * Sync symlinks for specified agents from AGENTS.md files
 */
export async function syncSymlinks(
  agentsPaths: string[],
  selectedAgents: string[],
  dryRun: boolean,
  verbose: boolean
): Promise<void> {
  let created = 0;
  let skipped = 0;
  const operations: string[] = [];

  const supportedAgents = await getSupportedAgents();

  for (const agentPath of agentsPaths) {
    const dir = dirname(agentPath);
    const filename = agentPath.split('/').pop()!;

    for (const agentName of selectedAgents) {
      const cfg = supportedAgents[agentName];

      let result;
      if (cfg.dir === '') {
        // Create symlink in same directory
        result = await createSymlink(dir, filename, cfg.file, dryRun, verbose);
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

  console.log(`AGENTS.md files found: ${agentsPaths.length}`);
  console.log(`Symlinks created: ${created}`);
  console.log(`Symlinks skipped: ${skipped}`);

  if (verbose && operations.length > 0) {
    console.log('\nOperations:');
    for (const op of operations) {
      console.log(op);
    }
  }
}

/**
 * Create a symlink in the same directory
 */
async function createSymlink(
  dir: string,
  source: string,
  target: string,
  dryRun: boolean,
  verbose: boolean
): Promise<{ created: number; skipped: number; operation: string }> {
  const targetPath = join(dir, target);

  try {
    const stats = await lstat(targetPath);
    // File exists, check if it's the right symlink
    if (await shouldSkipOrOverwrite(targetPath, source, stats, join(dir, source), dryRun, verbose)) {
      return { created: 0, skipped: 1, operation: verbose ? `skipped: ${targetPath} (already correct)` : '' };
    }

    // Remove existing file/symlink
    if (!dryRun) {
      await unlink(targetPath);
    }
  } catch {
    // File doesn't exist, proceed with creation
  }

  if (dryRun) {
    return {
      created: 1,
      skipped: 0,
      operation: verbose ? `would create: ${targetPath} -> ${source}` : '',
    };
  }

  try {
    await symlink(source, targetPath);
    return { created: 1, skipped: 0, operation: verbose ? `created: ${targetPath} -> ${source}` : '' };
  } catch (err) {
    return { created: 0, skipped: 1, operation: verbose ? `error: ${targetPath} (${err})` : '' };
  }
}

/**
 * Create a symlink in a subdirectory
 */
async function createSymlinkInDir(
  dir: string,
  source: string,
  subdir: string,
  target: string,
  dryRun: boolean,
  verbose: boolean
): Promise<{ created: number; skipped: number; operation: string }> {
  const subdirPath = join(dir, subdir);

  // Create subdirectory if needed
  if (!dryRun) {
    await mkdir(subdirPath, { recursive: true });
  }

  const targetPath = join(subdirPath, target);
  const symTarget = relative(subdirPath, join(dir, source));

  try {
    const stats = await lstat(targetPath);
    if (await shouldSkipOrOverwrite(targetPath, symTarget, stats, join(dir, source), dryRun, verbose)) {
      return { created: 0, skipped: 1, operation: verbose ? `skipped: ${targetPath} (already correct)` : '' };
    }

    if (!dryRun) {
      await unlink(targetPath);
    }
  } catch {
    // File doesn't exist
  }

  if (dryRun) {
    return {
      created: 1,
      skipped: 0,
      operation: verbose ? `would create: ${targetPath} -> ${symTarget}` : '',
    };
  }

  try {
    await symlink(symTarget, targetPath);
    return { created: 1, skipped: 0, operation: verbose ? `created: ${targetPath} -> ${symTarget}` : '' };
  } catch (err) {
    return { created: 0, skipped: 1, operation: verbose ? `error: ${targetPath} (${err})` : '' };
  }
}

/**
 * Check if we should skip or overwrite an existing file
 */
async function shouldSkipOrOverwrite(
  targetPath: string,
  expectedTarget: string,
  stats: any,
  sourcePath: string,
  dryRun: boolean,
  verbose: boolean
): Promise<boolean> {
  if (stats.isSymbolicLink()) {
    try {
      const link = await readlink(targetPath);
      if (link === expectedTarget) {
        return true; // Symlink is correct
      }
    } catch {
      // Can't read symlink
    }

    // Symlink points to wrong place
    if (!dryRun && !(await askForConfirmation(targetPath, expectedTarget))) {
      return true; // User said no
    }
    return false;
  }

  // It's a regular file, compare content
  try {
    const sourceContent = await readFile(sourcePath, 'utf-8');
    const targetContent = await readFile(targetPath, 'utf-8');

    if (sourceContent === targetContent) {
      return true; // Content matches
    }
  } catch {
    // Can't read files
  }

  // Content differs, ask user
  if (!dryRun && !(await askForConfirmation(targetPath, 'different version'))) {
    return true;
  }
  return false;
}

/**
 * Ask user for confirmation
 */
async function askForConfirmation(targetPath: string, reason: string): Promise<boolean> {
  return new Promise((resolve) => {
    const rl = createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    console.log(`\nFile already exists: ${targetPath} (${reason})`);
    rl.question('Overwrite? (y/n): ', (answer) => {
      rl.close();
      const response = answer.trim().toLowerCase();
      resolve(response === 'y' || response === 'yes');
    });
  });
}

/**
 * Delete guideline files for specified agents
 */
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
    const shouldDelete = selectedAgents.some((agentName) => file.agent.toLowerCase() === agentName.toLowerCase());

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

  console.log(`Files deleted: ${deleted}`);
  if (notFound > 0) {
    console.log(`Errors: ${notFound}`);
  }

  if (verbose && operations.length > 0) {
    console.log('Operations:');
    for (const op of operations) {
      console.log(op);
    }
  }
}
