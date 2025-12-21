import { readdir, stat, lstat, symlink, unlink, readlink, mkdir } from 'fs/promises';
import { join, dirname, basename, relative, extname } from 'path';
import { loadConfig, expandProviderPaths } from './config.js';
import { dirExists, isSymlinkFile } from './discovery.js';
import type { CommandFile } from './types.js';

const ignoreDir = new Set(['node_modules', '.git', 'dist', 'build', '.cursor']);

/**
 * Discover all COMMANDS directories recursively
 */
export async function discoverCommands(): Promise<string[]> {
  const commandDirs: string[] = [];
  const cwd = process.cwd();
  await walkDirectory(cwd, async (path, stats) => {
    if (stats.isDirectory() && basename(path) === 'COMMANDS') {
      commandDirs.push(path);
    }
  });
  return commandDirs;
}

/**
 * Discover all slash command files for all providers
 */
export async function discoverAllCommands(): Promise<CommandFile[]> {
  const files: CommandFile[] = [];
  const cwd = process.cwd();
  const config = await loadConfig();

  await walkDirectory(cwd, async (path, stats) => {
    if (stats.isFile()) {
      const dir = dirname(path);
      const filename = basename(path);

      // Check if this file matches any provider's command configuration
      for (const provider of Object.values(config.providers)) {
        // Check if file is in the provider's commands directory
        if (!path.includes(provider.commands.dir)) {
          continue;
        }

        // Check if file has the right extension
        if (provider.commands.extension && !filename.endsWith(provider.commands.extension)) {
          continue;
        }

        // Extract command name (filename without extension)
        let cmdName = filename;
        if (provider.commands.extension) {
          cmdName = filename.slice(0, -provider.commands.extension.length);
        }

        const isSymlink = await isSymlinkFile(path);

        files.push({
          path,
          dir,
          provider: provider.name.toUpperCase(),
          name: cmdName,
          file: filename,
          isSymlink,
          size: stats.size,
        });
        break;
      }
    }
  });

  return files;
}

/**
 * Discover only user/system-wide slash commands
 */
export async function discoverGlobalCommands(): Promise<CommandFile[]> {
  const files: CommandFile[] = [];
  const config = await loadConfig();

  for (const provider of Object.values(config.providers)) {
    const expandedProvider = expandProviderPaths(provider);
    for (const globalPath of expandedProvider.commands.global_paths) {
      if (!(await dirExists(globalPath))) {
        continue;
      }

      // Walk the global commands directory
      await walkDirectory(globalPath, async (path, stats) => {
        if (stats.isFile()) {
          const filename = basename(path);

          // Check if file has the right extension
          if (provider.commands.extension && !filename.endsWith(provider.commands.extension)) {
            return;
          }

          // Extract command name
          let cmdName = filename;
          if (provider.commands.extension) {
            cmdName = filename.slice(0, -provider.commands.extension.length);
          }

          const isSymlink = await isSymlinkFile(path);

          files.push({
            path,
            dir: dirname(path),
            provider: provider.name.toUpperCase(),
            name: cmdName,
            file: filename,
            isSymlink,
            size: stats.size,
          });
        }
      });
    }
  }

  return files;
}

/**
 * Sync command files for specified providers from COMMANDS directory
 */
export async function syncCommandFiles(
  commandDirs: string[],
  selectedProviders: string[],
  dryRun: boolean,
  verbose: boolean
): Promise<void> {
  let created = 0;
  let skipped = 0;
  const operations: string[] = [];

  const config = await loadConfig();

  for (const commandDir of commandDirs) {
    const parentDir = dirname(commandDir);

    // Read all files in COMMANDS directory
    let entries;
    try {
      entries = await readdir(commandDir, { withFileTypes: true });
    } catch (err) {
      if (verbose) {
        operations.push(`error reading ${commandDir}: ${err}`);
      }
      continue;
    }

    for (const providerName of selectedProviders) {
      const provider = config.providers[providerName];
      if (!provider) {
        continue;
      }

      // Create provider's commands directory
      const targetDir = join(parentDir, provider.commands.dir);
      if (!dryRun) {
        await mkdir(targetDir, { recursive: true });
      }

      // Sync each command file
      for (const entry of entries) {
        if (entry.isDirectory()) {
          continue;
        }

        const sourceFile = entry.name;

        // Skip if file doesn't have markdown extension
        if (!sourceFile.endsWith('.md')) {
          continue;
        }

        const sourcePath = join(commandDir, sourceFile);

        // Determine target filename
        let targetFile = sourceFile;
        if (provider.commands.extension && provider.commands.extension !== extname(sourceFile)) {
          // Change extension if needed
          const baseName = sourceFile.slice(0, -extname(sourceFile).length);
          targetFile = baseName + provider.commands.extension;
        }

        const targetPath = join(targetDir, targetFile);

        // Calculate relative path from target to source
        const relPath = relative(targetDir, sourcePath);

        // Create symlink
        const result = await createCommandSymlink(targetPath, relPath, dryRun, verbose);
        created += result.created;
        skipped += result.skipped;
        if (result.operation) {
          operations.push(result.operation);
        }
      }
    }
  }

  console.log(`COMMANDS directories found: ${commandDirs.length}`);
  console.log(`Command symlinks created: ${created}`);
  console.log(`Command symlinks skipped: ${skipped}`);

  if (verbose && operations.length > 0) {
    console.log('\nOperations:');
    for (const op of operations) {
      console.log(op);
    }
  }
}

/**
 * Create a symlink for a command file
 */
async function createCommandSymlink(
  targetPath: string,
  sourcePath: string,
  dryRun: boolean,
  verbose: boolean
): Promise<{ created: number; skipped: number; operation: string }> {
  // Check if target already exists
  try {
    const stats = await lstat(targetPath);
    if (stats.isSymbolicLink()) {
      const link = await readlink(targetPath);
      if (link === sourcePath) {
        // Symlink is already correct
        return {
          created: 0,
          skipped: 1,
          operation: verbose ? `skipped: ${targetPath} (already correct)` : '',
        };
      }
    }

    // File exists but is not the right symlink
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
      operation: verbose ? `would create: ${targetPath} -> ${sourcePath}` : '',
    };
  }

  // Create symlink
  try {
    await symlink(sourcePath, targetPath);
    return { created: 1, skipped: 0, operation: verbose ? `created: ${targetPath} -> ${sourcePath}` : '' };
  } catch (err) {
    return { created: 0, skipped: 1, operation: verbose ? `error: ${targetPath} (${err})` : '' };
  }
}

/**
 * Delete command files for specified providers
 */
export async function deleteCommandFiles(
  selectedProviders: string[],
  dryRun: boolean,
  verbose: boolean
): Promise<void> {
  let deleted = 0;
  let notFound = 0;
  const operations: string[] = [];

  const allFiles = await discoverAllCommands();

  for (const file of allFiles) {
    const shouldDelete = selectedProviders.some(
      (providerName) => file.provider.toLowerCase() === providerName.toLowerCase()
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

/**
 * Walk a directory tree recursively
 */
async function walkDirectory(
  dir: string,
  callback: (path: string, stats: any) => Promise<void>
): Promise<void> {
  try {
    const entries = await readdir(dir, { withFileTypes: true });

    for (const entry of entries) {
      const fullPath = join(dir, entry.name);

      if (entry.isDirectory()) {
        // Skip ignored directories
        if (ignoreDir.has(entry.name)) {
          continue;
        }
        await walkDirectory(fullPath, callback);
      } else {
        const stats = await stat(fullPath);
        await callback(fullPath, stats);
      }
    }
  } catch {
    // Silently skip directories we can't read
  }
}
