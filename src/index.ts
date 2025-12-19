#!/usr/bin/env bun

import { SupportedAgents, getAgentNames } from "./agents.ts";
import { discoverAgents, discoverAll, discoverGlobalOnly } from "./discovery.ts";
import { formatList } from "./output.ts";
import { syncSymlinks, deleteGuidelineFiles } from "./symlink.ts";

function printHelp(): void {
  let help = `Agent Guidelines Manager CLI

Usage: agents <command> [flags]

Commands:
  list                     Discover and display all guideline files with metadata
                           Flags:
                             --verbose    Show detailed output
                             -g           Show only user/system-wide agent guideline files
                             --global     Show only user/system-wide agent guideline files
                             --<agent>    Filter by specific agent files (e.g., --claude, --cursor)

  sync                     Find all AGENTS.md files and create symlinks
                           Flags:`;

  console.log(help);

  for (const agent of getAgentNames()) {
    const cfg = SupportedAgents[agent];
    console.log(`                             --${cfg.name}       Create ${cfg.file} symlinks`);
  }

  let help2 = `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`;
  console.log(help2);

  for (const agent of getAgentNames()) {
    const cfg = SupportedAgents[agent];
    console.log(`                             --${cfg.name}       Delete ${cfg.file} files`);
  }

  const help3 = `                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  help                     Show this help message

Examples:
  agents list
  agents list --verbose
  agents list --claude
  agents list --gemini --global
  agents list --claude --cursor --verbose
  agents sync --claude --cursor
  agents sync --claude --cursor --dry-run
  agents rm --claude
  agents rm --cursor --gemini --dry-run`;

  console.log(help3);
}

interface ParsedFlags {
  verbose: boolean;
  global: boolean;
  dryRun: boolean;
  agents: string[];
}

function parseFlags(args: string[]): ParsedFlags {
  const flags: ParsedFlags = {
    verbose: false,
    global: false,
    dryRun: false,
    agents: [],
  };

  const agentNames = getAgentNames();

  for (const arg of args) {
    if (arg === "--verbose") {
      flags.verbose = true;
    } else if (arg === "-g" || arg === "--global") {
      flags.global = true;
    } else if (arg === "--dry-run") {
      flags.dryRun = true;
    } else if (arg.startsWith("--")) {
      const flagName = arg.slice(2);
      if (agentNames.includes(flagName)) {
        flags.agents.push(flagName);
      }
    }
  }

  return flags;
}

async function cmdList(args: string[]): Promise<void> {
  const flags = parseFlags(args);

  let files;
  if (flags.global) {
    files = await discoverGlobalOnly();
  } else {
    files = await discoverAll();
  }

  // Filter by specified agents if any are provided
  if (flags.agents.length > 0) {
    files = files.filter((f) =>
      flags.agents.some((agent) => agent.toUpperCase() === f.agent)
    );
  }

  formatList(files, flags.verbose);
}

async function cmdSync(args: string[]): Promise<void> {
  const flags = parseFlags(args);

  if (flags.agents.length === 0) {
    const agentNames = getAgentNames();
    console.log(`Please specify at least one agent flag: --${agentNames.join(", --")}`);
    process.exit(1);
  }

  const agentsFiles = await discoverAgents();
  await syncSymlinks(agentsFiles, flags.agents, flags.dryRun, flags.verbose);
}

async function cmdRm(args: string[]): Promise<void> {
  const flags = parseFlags(args);

  if (flags.agents.length === 0) {
    const agentNames = getAgentNames();
    console.log(`Please specify at least one agent flag: --${agentNames.join(", --")}`);
    process.exit(1);
  }

  await deleteGuidelineFiles(flags.agents, flags.dryRun, flags.verbose);
}

async function main(): Promise<void> {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    printHelp();
    process.exit(1);
  }

  const command = args[0];
  const commandArgs = args.slice(1);

  switch (command) {
    case "list":
      await cmdList(commandArgs);
      break;
    case "sync":
      await cmdSync(commandArgs);
      break;
    case "rm":
      await cmdRm(commandArgs);
      break;
    case "help":
    case "-h":
    case "--help":
      printHelp();
      break;
    default:
      console.log(`Unknown command: ${command}`);
      process.exit(1);
  }
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
