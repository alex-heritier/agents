#!/usr/bin/env bun
import { getProviderConfig } from "./config";
import { discoverAll, discoverGlobalOnly, discoverSources } from "./discovery";
import { formatList, formatSyncSummary } from "./output";
import { deleteManagedFiles, syncSymlinks } from "./symlink";
import type { FileSpec, ProviderConfig, ProvidersConfig } from "./types";

const [command, ...args] = process.argv.slice(2);

if (!command) {
  printHelp();
  process.exit(1);
}

switch (command) {
  case "list":
    cmdList(args);
    break;
  case "sync":
    cmdSync(args);
    break;
  case "rm":
    cmdRm(args);
    break;
  case "list-commands":
    cmdListCommands(args);
    break;
  case "sync-commands":
    cmdSyncCommands(args);
    break;
  case "rm-commands":
    cmdRmCommands(args);
    break;
  case "help":
  case "-h":
  case "--help":
    printHelp();
    break;
  default:
    console.error(`Unknown command: ${command}`);
    process.exit(1);
}

function printHelp() {
  const cfg = getProviderConfig();
  const help = `Agent Guidelines Manager CLI

Usage: agents <command> [flags]

Commands:
  list                     Discover and display all guideline files with metadata
                           Flags:
                             --verbose    Show detailed output
                             -g           Show only user/system-wide agent guideline files
                             --global     Show only user/system-wide agent guideline files
                             --<agent>    Filter by specific agent files (e.g., --claude, --cursor)

  sync                     Find all guideline source files and create symlinks
                           Flags:`;
  process.stdout.write(help);

  for (const agent of getProviderNames(cfg, (provider) => provider.guidelines)) {
    const flagName = getProviderFlagName(cfg, agent);
    const guidelines = cfg.providers[agent].guidelines;
    if (guidelines) {
      process.stdout.write(`                             --${flagName}       Create ${guidelines.file} symlinks\n`);
    }
  }

  const help2 = `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`;
  process.stdout.write(help2);

  for (const agent of getProviderNames(cfg, (provider) => provider.guidelines)) {
    const flagName = getProviderFlagName(cfg, agent);
    const guidelines = cfg.providers[agent].guidelines;
    if (guidelines) {
      process.stdout.write(`                             --${flagName}       Delete ${guidelines.file} files\n`);
    }
  }

  const help3 = `                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  list-commands             Discover and display all command files with metadata
                           Flags:
                             --verbose    Show detailed output
                             --<agent>    Filter by specific command files (e.g., --claude, --cursor)

  sync-commands             Find all command source files and create symlinks
                           Flags:`;
  process.stdout.write(help3);

  for (const agent of getProviderNames(cfg, (provider) => provider.commands)) {
    const flagName = getProviderFlagName(cfg, agent);
    const commands = cfg.providers[agent].commands;
    if (commands) {
      process.stdout.write(`                             --${flagName}       Create ${commands.file} symlinks\n`);
    }
  }

  const help4 = `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm-commands               Delete command files for specified agents
                           Flags:`;
  process.stdout.write(help4);

  for (const agent of getProviderNames(cfg, (provider) => provider.commands)) {
    const flagName = getProviderFlagName(cfg, agent);
    const commands = cfg.providers[agent].commands;
    if (commands) {
      process.stdout.write(`                             --${flagName}       Delete ${commands.file} files\n`);
    }
  }

  const help5 = `                             --dry-run    Show what would be deleted without making changes
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
  agents rm --cursor --gemini --dry-run
  agents list-commands
  agents list-commands --claude
  agents sync-commands --claude --cursor
  agents rm-commands --claude
`;
  process.stdout.write(help5);
}

function cmdList(argv: string[]) {
  const cfg = getProviderConfig();
  const { flags } = parseArgs(argv);
  const verbose = flags.has("--verbose");
  const global = flags.has("-g") || flags.has("--global");

  const agentFlags = collectAgentFlags(cfg, (provider) => provider.guidelines, flags);
  const filterAgents = agentFlags.selected;

  const files = global
    ? discoverGlobalOnly(cfg)
    : discoverAll(cfg, cfg.sources.guidelines, (provider) => provider.guidelines);

  const filtered = filterAgents.length > 0 ? filterFilesByProviders(files, filterAgents) : files;
  formatList(filtered, verbose, "No guideline files found.");
}

function cmdSync(argv: string[]) {
  const cfg = getProviderConfig();
  const { flags } = parseArgs(argv);
  const dryRun = flags.has("--dry-run");
  const verbose = flags.has("--verbose");

  const selection = collectAgentFlags(cfg, (provider) => provider.guidelines, flags);
  const selectedAgents = ensureProvidersSelected(cfg, selection.available, selection.selected);

  const sourceFiles = discoverSources(cfg.sources.guidelines);
  const { created, skipped, operations } = syncSymlinks(
    sourceFiles,
    selectedAgents,
    cfg,
    (provider) => provider.guidelines,
    dryRun,
    verbose,
  );

  formatSyncSummary(cfg.sources.guidelines, sourceFiles.length, created, skipped, verbose, operations);
}

function cmdRm(argv: string[]) {
  const cfg = getProviderConfig();
  const { flags } = parseArgs(argv);
  const dryRun = flags.has("--dry-run");
  const verbose = flags.has("--verbose");

  const selection = collectAgentFlags(cfg, (provider) => provider.guidelines, flags);
  const selectedAgents = ensureProvidersSelected(cfg, selection.available, selection.selected);

  deleteManagedFiles(selectedAgents, cfg, cfg.sources.guidelines, (provider) => provider.guidelines, dryRun, verbose);
}

function cmdListCommands(argv: string[]) {
  const cfg = getProviderConfig();
  const { flags } = parseArgs(argv);
  const verbose = flags.has("--verbose");

  const selection = collectAgentFlags(cfg, (provider) => provider.commands, flags);
  const filterAgents = selection.selected;

  const files = discoverAll(cfg, cfg.sources.commands, (provider) => provider.commands);
  const filtered = filterAgents.length > 0 ? filterFilesByProviders(files, filterAgents) : files;
  formatList(filtered, verbose, "No command files found.");
}

function cmdSyncCommands(argv: string[]) {
  const cfg = getProviderConfig();
  const { flags } = parseArgs(argv);
  const dryRun = flags.has("--dry-run");
  const verbose = flags.has("--verbose");

  const selection = collectAgentFlags(cfg, (provider) => provider.commands, flags);
  const selectedAgents = ensureProvidersSelected(cfg, selection.available, selection.selected);

  const sourceFiles = discoverSources(cfg.sources.commands);
  const { created, skipped, operations } = syncSymlinks(
    sourceFiles,
    selectedAgents,
    cfg,
    (provider) => provider.commands,
    dryRun,
    verbose,
  );

  formatSyncSummary(cfg.sources.commands, sourceFiles.length, created, skipped, verbose, operations);
}

function cmdRmCommands(argv: string[]) {
  const cfg = getProviderConfig();
  const { flags } = parseArgs(argv);
  const dryRun = flags.has("--dry-run");
  const verbose = flags.has("--verbose");

  const selection = collectAgentFlags(cfg, (provider) => provider.commands, flags);
  const selectedAgents = ensureProvidersSelected(cfg, selection.available, selection.selected);

  deleteManagedFiles(selectedAgents, cfg, cfg.sources.commands, (provider) => provider.commands, dryRun, verbose);
}

function parseArgs(argv: string[]) {
  const flags = new Set<string>();
  for (const arg of argv) {
    if (arg.startsWith("--") || arg.startsWith("-")) {
      flags.add(arg);
    }
  }
  return { flags };
}

function getProviderNames(cfg: ProvidersConfig, specSelector: (provider: ProviderConfig) => FileSpec | undefined): string[] {
  return Object.entries(cfg.providers)
    .filter(([, provider]) => Boolean(specSelector(provider)))
    .map(([name]) => name)
    .sort();
}

function getProviderFlagName(cfg: ProvidersConfig, providerName: string): string {
  return cfg.providers[providerName]?.name || providerName;
}

function collectAgentFlags(
  cfg: ProvidersConfig,
  specSelector: (provider: ProviderConfig) => FileSpec | undefined,
  flags: Set<string>,
) {
  const available = getProviderNames(cfg, specSelector);
  const selected: string[] = [];

  for (const providerName of available) {
    const flagName = `--${getProviderFlagName(cfg, providerName)}`;
    if (flags.has(flagName)) {
      selected.push(providerName);
    }
  }

  return { available, selected };
}

function ensureProvidersSelected(cfg: ProvidersConfig, available: string[], selected: string[]) {
  if (selected.length > 0) {
    return selected;
  }

  if (available.length === 0) {
    console.error("No agents configured.");
    process.exit(1);
  }

  const first = getProviderFlagName(cfg, available[0]);
  const rest = available.slice(1).map((name) => `--${getProviderFlagName(cfg, name)}`);
  console.error(`Please specify at least one agent flag: --${first}${rest.length ? ", " + rest.join(", ") : ""}`);
  process.exit(1);
}

function filterFilesByProviders(files: ReturnType<typeof discoverAll>, providers: string[]) {
  const lower = providers.map((provider) => provider.toLowerCase());
  return files.filter((file) => lower.includes(file.agent.toLowerCase()));
}
