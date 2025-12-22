import { homedir } from "node:os";
import { basename, dirname, join, relative } from "node:path";

export function pathBasename(path: string): string {
  return basename(path);
}

export function pathDirname(path: string): string {
  return dirname(path);
}

export function pathJoin(...parts: string[]): string {
  return join(...parts);
}

export function pathRelative(from: string, to: string): string {
  return relative(from, to);
}

export function getHomeDir(): string {
  return homedir();
}

export function expandHomePath(path: string): string {
  if (!path.startsWith("~")) {
    return path;
  }
  return pathJoin(getHomeDir(), path.replace(/^~\/?/, ""));
}

export function formatRelativeDir(
  targetDir: string,
  cwd: string,
  homeDir: string,
  showRelativeToHome: boolean,
): string {
  if (showRelativeToHome && homeDir && targetDir.startsWith(homeDir)) {
    const relDir = pathRelative(homeDir, targetDir);
    return relDir === "" ? "~/" : `~/${relDir}`;
  }
  const relDir = pathRelative(cwd, targetDir);
  if (!relDir) {
    return "./";
  }
  if (!relDir.startsWith(".")) {
    return `./${relDir}`;
  }
  return `${relDir}/`;
}
