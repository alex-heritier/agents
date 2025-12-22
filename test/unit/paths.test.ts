import { describe, expect, it } from "bun:test";
import { expandHomePath, formatRelativeDir, pathBasename, pathDirname, pathJoin, pathRelative } from "../../src/paths";
import { homedir } from "node:os";

const home = homedir();

describe("paths", () => {
  it("joins and splits paths", () => {
    const joined = pathJoin("/tmp", "agents", "file.txt");
    expect(pathDirname(joined)).toBe("/tmp/agents");
    expect(pathBasename(joined)).toBe("file.txt");
  });

  it("expands home paths", () => {
    const expanded = expandHomePath("~/projects");
    expect(expanded).toBe(pathJoin(home, "projects"));
  });

  it("formats relative directory to home", () => {
    const target = pathJoin(home, "projects", "demo");
    const display = formatRelativeDir(target, "/tmp", home, true);
    expect(display).toBe("~/projects/demo");
  });

  it("formats relative directory to cwd", () => {
    const display = formatRelativeDir("/tmp/demo", "/tmp", home, false);
    expect(display).toBe("./demo");
  });

  it("returns relative path", () => {
    expect(pathRelative("/tmp", "/tmp/demo/file.txt")).toBe("demo/file.txt");
  });
});
