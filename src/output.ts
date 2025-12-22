import { homedir } from "node:os";
import type { ManagedFile } from "./types";
import { formatRelativeDir } from "./discovery";

export function formatList(files: ManagedFile[], verbose: boolean, emptyMessage: string) {
  if (files.length === 0) {
    console.log(emptyMessage);
    return;
  }

  console.log(`${"Directory".padEnd(30)} ${"File".padEnd(15)} ${"Tokens".padEnd(10)}`);
  console.log("-".repeat(55));

  const cwd = process.cwd();
  const homeDir = homedir();

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
