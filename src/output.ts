import { relative } from 'path';
import type { GuidelineFile, CommandFile } from './types.js';

/**
 * Estimate token count from file size (rough estimate: 4 bytes per token)
 */
function estimateTokens(size: number): number {
  return Math.floor(size / 4);
}

/**
 * Format and display a list of guideline files
 */
export function formatList(files: GuidelineFile[], verbose: boolean): void {
  if (files.length === 0) {
    console.log('No guideline files found.');
    return;
  }

  // Print header
  console.log('Directory'.padEnd(30) + 'File'.padEnd(15) + 'Tokens');
  console.log('-'.repeat(55));

  const cwd = process.cwd();
  const homeDir = process.env.HOME || '';

  // Determine if we should show paths relative to home directory
  let showRelativeToHome = false;
  if (files.length > 0 && homeDir) {
    const standardGlobalPatterns = [
      `${homeDir}/.claude`,
      `${homeDir}/.codex`,
      `${homeDir}/.gemini`,
      `${homeDir}/.config`,
    ];

    const allFilesAreGlobal = files.every((f) =>
      standardGlobalPatterns.some((pattern) => f.dir.startsWith(pattern))
    );

    showRelativeToHome = allFilesAreGlobal;
  }

  for (const f of files) {
    let displayDir: string;

    if (showRelativeToHome && homeDir && f.dir.startsWith(homeDir)) {
      // Make directory relative to home directory and prefix with ~/
      const relDir = f.dir.slice(homeDir.length);
      displayDir = relDir === '' || relDir === '/' ? '~/' : `~${relDir}`;
    } else {
      // Make directory relative to cwd
      const relDir = relative(cwd, f.dir);
      if (relDir === '') {
        displayDir = './';
      } else if (!relDir.startsWith('.')) {
        displayDir = './' + relDir;
      } else {
        displayDir = relDir + '/';
      }
    }

    let filename = f.file;
    let tokensStr: string;
    if (f.isSymlink) {
      filename = '*' + filename;
      tokensStr = '-';
    } else {
      const tokens = estimateTokens(f.size);
      tokensStr = tokens.toString();
    }

    console.log(displayDir.padEnd(30) + filename.padEnd(15) + tokensStr);
  }

  if (verbose) {
    console.log(`\nTotal: ${files.length} files found`);
  }
}

/**
 * Format and display a list of command files
 */
export function formatCommandList(files: CommandFile[], verbose: boolean): void {
  if (files.length === 0) {
    console.log('No command files found.');
    return;
  }

  // Print header
  console.log('Directory'.padEnd(30) + 'Command'.padEnd(20) + 'Provider'.padEnd(15) + 'Tokens');
  console.log('-'.repeat(75));

  const cwd = process.cwd();
  const homeDir = process.env.HOME || '';

  // Determine if we should show paths relative to home directory
  let showRelativeToHome = false;
  if (files.length > 0 && homeDir) {
    const standardGlobalPatterns = [
      `${homeDir}/.claude`,
      `${homeDir}/.cursor`,
      `${homeDir}/.config`,
    ];

    const allFilesAreGlobal = files.every((f) =>
      standardGlobalPatterns.some((pattern) => f.dir.startsWith(pattern))
    );

    showRelativeToHome = allFilesAreGlobal;
  }

  for (const f of files) {
    let displayDir: string;

    if (showRelativeToHome && homeDir && f.dir.startsWith(homeDir)) {
      const relDir = f.dir.slice(homeDir.length);
      displayDir = relDir === '' || relDir === '/' ? '~/' : `~${relDir}`;
    } else {
      const relDir = relative(cwd, f.dir);
      if (relDir === '') {
        displayDir = './';
      } else if (!relDir.startsWith('.')) {
        displayDir = './' + relDir;
      } else {
        displayDir = relDir + '/';
      }
    }

    let cmdName = f.name;
    let tokensStr: string;
    if (f.isSymlink) {
      cmdName = '*' + cmdName;
      tokensStr = '-';
    } else {
      const tokens = estimateTokens(f.size);
      tokensStr = tokens.toString();
    }

    console.log(displayDir.padEnd(30) + cmdName.padEnd(20) + f.provider.padEnd(15) + tokensStr);
  }

  if (verbose) {
    console.log(`\nTotal: ${files.length} command files found`);
  }
}
