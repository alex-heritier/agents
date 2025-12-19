import { relative } from "node:path";
import { homedir } from "node:os";
import type { GuidelineFile } from "./types.ts";

function estimateTokens(size: number): number {
  return Math.floor(size / 4);
}

export function formatList(files: GuidelineFile[], verbose: boolean): void {
  if (files.length === 0) {
    console.log("No guideline files found.");
    return;
  }

  // Print header
  console.log(`${"Directory".padEnd(30)} ${"File".padEnd(15)} ${"Tokens".padEnd(10)}`);
  console.log("-".repeat(55));

  const cwd = process.cwd();
  const home = homedir() || "";

  // Determine if we should show paths relative to home directory (~)
  // We do this when all files are in standard global agent locations
  let showRelativeToHome = false;
  if (files.length > 0 && home) {
    const standardGlobalPatterns = [
      `${home}/.claude`,
      `${home}/.codex`,
      `${home}/.gemini`,
      `${home}/.config`,
    ];

    const allFilesAreGlobal = files.every((f) =>
      standardGlobalPatterns.some((pattern) => f.dir.startsWith(pattern))
    );

    if (allFilesAreGlobal) {
      showRelativeToHome = true;
    }
  }

  for (const f of files) {
    let displayDir: string;

    if (showRelativeToHome && home && f.dir.startsWith(home)) {
      // Make directory relative to home directory and prefix with ~/
      const relDir = relative(home, f.dir);
      if (!relDir || relDir === ".") {
        displayDir = "~/";
      } else {
        displayDir = "~/" + relDir;
      }
    } else {
      // Make directory relative to cwd
      const relDir = relative(cwd, f.dir);
      if (relDir === "" || relDir === ".") {
        displayDir = "./";
      } else if (!relDir.startsWith(".")) {
        displayDir = "./" + relDir;
      } else {
        displayDir = relDir + "/";
      }
    }

    let filename = f.file;
    let tokensStr: string;
    if (f.isSymlink) {
      filename = "*" + filename;
      tokensStr = "-";
    } else {
      const tokens = estimateTokens(f.size);
      tokensStr = tokens.toString();
    }

    console.log(`${displayDir.padEnd(30)} ${filename.padEnd(15)} ${tokensStr.padEnd(10)}`);
  }

  if (verbose) {
    console.log(`\nTotal: ${files.length} files found`);
  }
}

export function formatSyncSummary(
  found: number,
  created: number,
  skipped: number,
  verbose: boolean,
  operations: string[]
): void {
  console.log(`AGENTS.md files found: ${found}`);
  console.log(`Symlinks created: ${created}`);
  console.log(`Symlinks skipped: ${skipped}`);

  if (verbose && operations.length > 0) {
    console.log("\nOperations:");
    for (const op of operations) {
      console.log(op);
    }
  }
}

export function formatRmSummary(
  deleted: number,
  notFound: number,
  verbose: boolean,
  operations: string[]
): void {
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
