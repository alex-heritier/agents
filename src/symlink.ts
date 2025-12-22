import { lstatSync, mkdirSync, readFileSync, readlinkSync, rmSync, symlinkSync } from "node:fs";
import type { FileSpec, ProviderConfig, ProvidersConfig } from "./types";
import { discoverAll } from "./discovery";
import { pathBasename, pathDirname, pathJoin, pathRelative } from "./paths";

export function syncSymlinks(
  sources: string[],
  selectedProviders: string[],
  cfg: ProvidersConfig,
  specSelector: (provider: ProviderConfig) => FileSpec | undefined,
  dryRun: boolean,
  verbose: boolean,
): { created: number; skipped: number; operations: string[] } {
  let created = 0;
  let skipped = 0;
  const operations: string[] = [];

  for (const sourcePath of sources) {
    const dir = pathDirname(sourcePath);
    const filename = pathBasename(sourcePath);

    for (const providerName of selectedProviders) {
      const provider = cfg.providers[providerName];
      const spec = specSelector(provider);
      if (!spec) {
        continue;
      }

      let result: { created: number; skipped: number; op?: string };
      if (!spec.dir) {
        result = createSymlink(dir, filename, spec.file, dryRun, verbose);
      } else {
        result = createSymlinkInDir(dir, filename, spec.dir, spec.file, dryRun, verbose);
      }
      created += result.created;
      skipped += result.skipped;
      if (result.op) {
        operations.push(result.op);
      }
    }
  }

  return { created, skipped, operations };
}

export function deleteManagedFiles(
  selectedProviders: string[],
  cfg: ProvidersConfig,
  sourceName: string,
  specSelector: (provider: ProviderConfig) => FileSpec | undefined,
  dryRun: boolean,
  verbose: boolean,
) {
  let deleted = 0;
  let notFound = 0;
  const operations: string[] = [];

  const allFiles = discoverAll(cfg, sourceName, specSelector);

  for (const file of allFiles) {
    const shouldDelete = selectedProviders.some((provider) => file.agent.toLowerCase() === provider.toLowerCase());
    if (!shouldDelete) {
      continue;
    }

    if (dryRun) {
      deleted += 1;
      if (verbose) {
        operations.push(`would delete: ${file.path}`);
      }
      continue;
    }

    try {
      rmSync(file.path);
      deleted += 1;
      if (verbose) {
        operations.push(`deleted: ${file.path}`);
      }
    } catch (error) {
      notFound += 1;
      if (verbose) {
        operations.push(`error: ${file.path} (${error})`);
      }
    }
  }

  formatRmSummary(deleted, notFound, verbose, operations);
}

function createSymlink(dir: string, source: string, target: string, dryRun: boolean, verbose: boolean) {
  const sourcePath = pathJoin(dir, source);
  const targetPath = pathJoin(dir, target);

  if (exists(targetPath)) {
    const info = lstatSync(targetPath);
    if (shouldSkipOrOverwrite(targetPath, source, info.isSymbolicLink(), sourcePath, dryRun)) {
      return { created: 0, skipped: 1, op: verbose ? `skipped: ${targetPath} (already correct)` : undefined };
    }
    if (!dryRun) {
      rmSync(targetPath);
    }
  }

  if (dryRun) {
    return { created: 1, skipped: 0, op: verbose ? `would create: ${targetPath} -> ${source}` : undefined };
  }

  try {
    symlinkSync(source, targetPath);
    return { created: 1, skipped: 0, op: verbose ? `created: ${targetPath} -> ${source}` : undefined };
  } catch (error) {
    return { created: 0, skipped: 1, op: verbose ? `error: ${targetPath} (${error})` : undefined };
  }
}

function createSymlinkInDir(
  dir: string,
  source: string,
  subdir: string,
  target: string,
  dryRun: boolean,
  verbose: boolean,
) {
  const subdirPath = pathJoin(dir, subdir);
  if (!exists(subdirPath)) {
    if (!dryRun) {
      mkdirSync(subdirPath, { recursive: true });
    }
  }

  const sourcePath = pathJoin(dir, source);
  const targetPath = pathJoin(subdirPath, target);
  const symTarget = pathRelative(subdirPath, sourcePath);

  if (exists(targetPath)) {
    const info = lstatSync(targetPath);
    if (shouldSkipOrOverwrite(targetPath, symTarget, info.isSymbolicLink(), sourcePath, dryRun)) {
      return { created: 0, skipped: 1, op: verbose ? `skipped: ${targetPath} (already correct)` : undefined };
    }
    if (!dryRun) {
      rmSync(targetPath);
    }
  }

  if (dryRun) {
    return { created: 1, skipped: 0, op: verbose ? `would create: ${targetPath} -> ${symTarget}` : undefined };
  }

  try {
    symlinkSync(symTarget, targetPath);
    return { created: 1, skipped: 0, op: verbose ? `created: ${targetPath} -> ${symTarget}` : undefined };
  } catch (error) {
    return { created: 0, skipped: 1, op: verbose ? `error: ${targetPath} (${error})` : undefined };
  }
}

function shouldSkipOrOverwrite(
  targetPath: string,
  expectedTarget: string,
  isSymlink: boolean,
  sourcePath: string,
  dryRun: boolean,
) {
  if (!exists(targetPath)) {
    return false;
  }

  if (isSymlink) {
    try {
      const link = readlinkSync(targetPath);
      if (link === expectedTarget) {
        return true;
      }
    } catch {
      return false;
    }
    if (!dryRun && !askForConfirmation(targetPath, expectedTarget)) {
      return true;
    }
    return false;
  }

  const sourceContent = safeReadFile(sourcePath);
  const targetContent = safeReadFile(targetPath);
  if (sourceContent === null || targetContent === null) {
    if (!dryRun && !askForConfirmation(targetPath, "source")) {
      return true;
    }
    return false;
  }

  if (sourceContent === targetContent) {
    return true;
  }

  if (!dryRun && !askForConfirmation(targetPath, "different version")) {
    return true;
  }
  return false;
}

function askForConfirmation(targetPath: string, reason: string): boolean {
  try {
    process.stdout.write(`\nFile already exists: ${targetPath} (${reason})\nOverwrite? (y/n): `);
  } catch {
    return false;
  }
  const response = readLine();
  if (!response) {
    return false;
  }
  const normalized = response.trim().toLowerCase();
  return normalized === "y" || normalized === "yes";
}

function readLine(): string | null {
  try {
    const data = readFileSync(0, "utf-8");
    if (!data) {
      return null;
    }
    const line = data.split(/\r?\n/)[0];
    return line === undefined ? null : line;
  } catch {
    return null;
  }
}

function safeReadFile(path: string): string | null {
  try {
    return readFileSync(path, "utf-8");
  } catch {
    return null;
  }
}

function exists(path: string): boolean {
  try {
    lstatSync(path);
    return true;
  } catch {
    return false;
  }
}

function formatRmSummary(deleted: number, notFound: number, verbose: boolean, operations: string[]) {
  console.log(`Files deleted: ${deleted}`);
  if (notFound > 0) {
    console.log(`Errors: ${notFound}`);
  }

  if (verbose && operations.length > 0) {
    console.log("Operations:");
    for (const op of operations) {
      console.log(op);
    }
  }
}
