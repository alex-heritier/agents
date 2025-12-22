import { describe, expect, it } from "bun:test";
import { parseArgs } from "../../src/args";

const allowed = new Set(["--verbose", "--dry-run", "-g", "-h"]);

describe("parseArgs", () => {
  it("parses long flags and short flags", () => {
    const result = parseArgs(["--verbose", "-g"], allowed);
    expect(result.flags.has("--verbose")).toBe(true);
    expect(result.flags.has("-g")).toBe(true);
    expect(result.unknown).toEqual([]);
    expect(result.help).toBe(false);
  });

  it("parses combined short flags", () => {
    const result = parseArgs(["-gh"], allowed);
    expect(result.flags.has("-g")).toBe(true);
    expect(result.flags.has("-h")).toBe(true);
    expect(result.help).toBe(false);
  });

  it("flags unknown options", () => {
    const result = parseArgs(["--nope", "-x"], allowed);
    expect(result.unknown).toEqual(["--nope", "-x"]);
  });

  it("sets help when --help or -h appears", () => {
    const result = parseArgs(["--help"], allowed);
    expect(result.help).toBe(true);
  });
});
