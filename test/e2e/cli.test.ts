import { afterEach, describe, expect, it } from "bun:test";
import { existsSync, readlinkSync, symlinkSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { cleanupTempDir, createTempDir, getRepoRoot } from "../helpers";

const repoRoot = getRepoRoot();
const cliPath = join(repoRoot, "src", "index.ts");

const tempDirs: string[] = [];

afterEach(() => {
  while (tempDirs.length > 0) {
    const dir = tempDirs.pop();
    if (dir) {
      cleanupTempDir(dir);
    }
  }
});

describe("agents CLI e2e", () => {
  it("lists empty guideline files", () => {
    const dir = createTempDir();
    tempDirs.push(dir);

    const result = Bun.spawnSync(["bun", cliPath, "list"], { cwd: dir });

    expect(result.exitCode).toBe(0);
    expect(result.stdout.toString()).toContain("No guideline files found.");
  });

  it("syncs guideline files and creates symlink", () => {
    const dir = createTempDir();
    tempDirs.push(dir);

    const agentsPath = join(dir, "AGENTS.md");
    writeFileSync(agentsPath, "# Guidelines\n");

    const result = Bun.spawnSync(["bun", cliPath, "sync", "--claude"], { cwd: dir });

    expect(result.exitCode).toBe(0);

    const linkPath = join(dir, "CLAUDE.md");
    const linkTarget = readlinkSync(linkPath);
    expect(linkTarget).toBe("AGENTS.md");
  });

  it("syncs command files and creates command symlink", () => {
    const dir = createTempDir();
    tempDirs.push(dir);

    const commandsPath = join(dir, "COMMANDS.md");
    writeFileSync(commandsPath, "/hello\n");

    const result = Bun.spawnSync(["bun", cliPath, "sync-commands", "--claude"], { cwd: dir });

    expect(result.exitCode).toBe(0);

    const linkPath = join(dir, ".claude", "commands", "commands.md");
    const linkTarget = readlinkSync(linkPath);
    expect(linkTarget).toBe(join("..", "..", "COMMANDS.md"));
  });

  it("rejects unknown flags", () => {
    const dir = createTempDir();
    tempDirs.push(dir);

    const result = Bun.spawnSync(["bun", cliPath, "list", "--nope"], { cwd: dir });

    expect(result.exitCode).toBe(1);
    expect(result.stderr.toString()).toContain("Unknown flags for list: --nope");
  });

  it("lists command files when present", () => {
    const dir = createTempDir();
    tempDirs.push(dir);

    const commandsPath = join(dir, "COMMANDS.md");
    writeFileSync(commandsPath, "/hello\n");

    const result = Bun.spawnSync(["bun", cliPath, "list-commands"], { cwd: dir });

    expect(result.exitCode).toBe(0);
    expect(result.stdout.toString()).toContain("COMMANDS.md");
  });

  it("removes synced guideline files", () => {
    const dir = createTempDir();
    tempDirs.push(dir);

    const agentsPath = join(dir, "AGENTS.md");
    writeFileSync(agentsPath, "# Guidelines\n");
    symlinkSync("AGENTS.md", join(dir, "CLAUDE.md"));

    const result = Bun.spawnSync(["bun", cliPath, "rm", "--claude"], { cwd: dir });

    expect(result.exitCode).toBe(0);
    expect(existsSync(join(dir, "CLAUDE.md"))).toBe(false);
  });

  it("shows help with --help", () => {
    const dir = createTempDir();
    tempDirs.push(dir);

    const result = Bun.spawnSync(["bun", cliPath, "list", "--help"], { cwd: dir });

    expect(result.exitCode).toBe(0);
    expect(result.stdout.toString()).toContain("Usage: agents <command> [flags]");
  });
});
