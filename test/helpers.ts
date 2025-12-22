import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

export function createTempDir(prefix = "agents-test-") {
  const dir = mkdtempSync(join(tmpdir(), prefix));
  return dir;
}

export function cleanupTempDir(dir: string) {
  rmSync(dir, { recursive: true, force: true });
}

export function getRepoRoot(): string {
  return join(import.meta.dir, "..");
}
