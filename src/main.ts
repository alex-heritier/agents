#!/usr/bin/env bun

import { parseArgs } from 'util';
import { getAgentNames, getSupportedAgents } from './agents.js';
import { discoverAgents, discoverAll, discoverGlobalOnly } from './discovery.js';
import { discoverCommands, discoverAllCommands, discoverGlobalCommands, syncCommandFiles, deleteCommandFiles } from './commands.js';
import { syncSymlinks, deleteGuidelineFiles } from './symlink.js';
import { formatList, formatCommandList } from './output.js';

async function main() {
  const args = process.argv.slice(2);

  if (args.length === 0) {
    printHelp();
    process.exit(1);
  }

  const command = args[0];

  switch (command) {
    case 'list':
      await cmdList(args.slice(1));
      break;
    case 'list-commands':
    case 'list-cmds':
      await cmdListCommands(args.slice(1));
      break;
    case 'sync':
      await cmdSync(args.slice(1));
      break;
    case 'sync-commands':
    case 'sync-cmds':
      await cmdSyncCommands(args.slice(1));
      break;
    case 'rm':
      await cmdRm(args.slice(1));
      break;
    case 'rm-commands':
    case 'rm-cmds':
      await cmdRmCommands(args.slice(1));
      break;
    case 'help':
    case '-h':
    case '--help':
      printHelp();
      break;
    default:
      console.log(`Unknown command: ${command}`);
      process.exit(1);
  }
}

async function printHelp() {
  const agentNames = await getAgentNames();
  const agents = await getSupportedAgents();

  console.log(`Agent Guidelines Manager CLI

Usage: agents <command> [flags]

Commands:
  list                     Discover and display all guideline files with metadata
                           Flags:
                             --verbose    Show detailed output
                             -g           Show only user/system-wide agent guideline files
                             --global     Show only user/system-wide agent guideline files
                             --<agent>    Filter by specific agent files (e.g., --claude, --cursor)

  list-commands            Discover and display all slash command files
  list-cmds                Alias for list-commands
                           Flags:
                             --verbose    Show detailed output
                             -g           Show only user/system-wide command files
                             --global     Show only user/system-wide command files
                             --<agent>    Filter by specific agent (e.g., --claude, --cursor)

  sync                     Find all AGENTS.md files and create symlinks
                           Flags:`);

  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Create ${cfg.file} symlinks`);
  }

  console.log(`                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  sync-commands            Find all COMMANDS directories and create command symlinks
  sync-cmds                Alias for sync-commands
                           Flags:`);

  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Create command symlinks for ${cfg.name}`);
  }

  console.log(`                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`);

  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Delete ${cfg.file} files`);
  }

  console.log(`                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  rm-commands              Delete command files for specified agents
  rm-cmds                  Alias for rm-commands
                           Flags:`);

  for (const agent of agentNames) {
    const cfg = agents[agent];
    console.log(`                             --${cfg.name}       Delete command files for ${cfg.name}`);
  }

  console.log(`                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  help                     Show this help message

Examples:
  # Guideline files
  agents list
  agents list --verbose
  agents list --claude
  agents list --gemini --global
  agents sync --claude --cursor
  agents sync --claude --cursor --dry-run
  agents rm --claude
  agents rm --cursor --gemini --dry-run

  # Slash commands
  agents list-commands
  agents list-commands --verbose
  agents list-commands --claude --global
  agents sync-commands --claude --cursor
  agents sync-commands --claude --dry-run
  agents rm-commands --claude
`);
}

async function cmdList(args: string[]) {
  const flags = parseFlags(args, ['verbose', 'g', 'global']);
  const verbose = flags.verbose;
  const global = flags.g || flags.global;

  // Get agent names and determine which to filter by
  const agentNames = await getAgentNames();
  const filterAgents: string[] = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      filterAgents.push(agent);
    }
  }

  let files = global ? await discoverGlobalOnly() : await discoverAll();

  // Filter by specified agents if any are provided
  if (filterAgents.length > 0) {
    files = files.filter((f) => filterAgents.some((agent) => agent.toUpperCase() === f.agent));
  }

  formatList(files, verbose);
}

async function cmdListCommands(args: string[]) {
  const flags = parseFlags(args, ['verbose', 'g', 'global']);
  const verbose = flags.verbose;
  const global = flags.g || flags.global;

  const agentNames = await getAgentNames();
  const filterAgents: string[] = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      filterAgents.push(agent);
    }
  }

  let files = global ? await discoverGlobalCommands() : await discoverAllCommands();

  if (filterAgents.length > 0) {
    files = files.filter((f) => filterAgents.some((agent) => agent.toUpperCase() === f.provider));
  }

  formatCommandList(files, verbose);
}

async function cmdSync(args: string[]) {
  const flags = parseFlags(args, ['dry-run', 'verbose']);
  const dryRun = flags['dry-run'];
  const verbose = flags.verbose;

  const agentNames = await getAgentNames();
  const selectedAgents: string[] = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }

  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(', --')}`);
    process.exit(1);
  }

  const agentsFiles = await discoverAgents();
  await syncSymlinks(agentsFiles, selectedAgents, dryRun, verbose);
}

async function cmdSyncCommands(args: string[]) {
  const flags = parseFlags(args, ['dry-run', 'verbose']);
  const dryRun = flags['dry-run'];
  const verbose = flags.verbose;

  const agentNames = await getAgentNames();
  const selectedAgents: string[] = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }

  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(', --')}`);
    process.exit(1);
  }

  const commandDirs = await discoverCommands();
  await syncCommandFiles(commandDirs, selectedAgents, dryRun, verbose);
}

async function cmdRm(args: string[]) {
  const flags = parseFlags(args, ['dry-run', 'verbose']);
  const dryRun = flags['dry-run'];
  const verbose = flags.verbose;

  const agentNames = await getAgentNames();
  const selectedAgents: string[] = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }

  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(', --')}`);
    process.exit(1);
  }

  await deleteGuidelineFiles(selectedAgents, dryRun, verbose);
}

async function cmdRmCommands(args: string[]) {
  const flags = parseFlags(args, ['dry-run', 'verbose']);
  const dryRun = flags['dry-run'];
  const verbose = flags.verbose;

  const agentNames = await getAgentNames();
  const selectedAgents: string[] = [];
  for (const agent of agentNames) {
    if (flags[agent]) {
      selectedAgents.push(agent);
    }
  }

  if (selectedAgents.length === 0) {
    console.log(`Please specify at least one agent flag: --${agentNames.join(', --')}`);
    process.exit(1);
  }

  await deleteCommandFiles(selectedAgents, dryRun, verbose);
}

/**
 * Simple flag parser
 */
function parseFlags(args: string[], knownFlags: string[]): Record<string, boolean> {
  const flags: Record<string, boolean> = {};

  for (const arg of args) {
    if (arg.startsWith('--')) {
      const flag = arg.slice(2);
      flags[flag] = true;
    } else if (arg.startsWith('-')) {
      const flag = arg.slice(1);
      flags[flag] = true;
    }
  }

  return flags;
}

// Run the main function
main().catch((err) => {
  console.error('Error:', err);
  process.exit(1);
});
