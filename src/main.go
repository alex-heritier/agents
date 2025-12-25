package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		os.Exit(1)
	}

	module := args[0]

	if module == "help" || module == "-h" || module == "--help" {
		printHelp()
		os.Exit(0)
	}

	if len(args) < 2 {
		printModuleHelp(module)
		os.Exit(0)
	}

	command := args[1]
	cmdArgs := args[2:]

	switch module {
	case "rule":
		switch command {
		case "list":
			cmdRuleList(cmdArgs)
		case "sync":
			cmdRuleSync(cmdArgs)
		case "rm":
			cmdRuleRm(cmdArgs)
		case "help", "-h", "--help":
			printRuleHelp()
		default:
			fmt.Fprintf(os.Stderr, "Unknown command for module 'rule': %s\n", command)
			os.Exit(1)
		}
	case "command":
		switch command {
		case "list":
			cmdCommandList(cmdArgs)
		case "sync":
			cmdCommandSync(cmdArgs)
		case "rm":
			cmdCommandRm(cmdArgs)
		case "help", "-h", "--help":
			printCommandHelp()
		default:
			fmt.Fprintf(os.Stderr, "Unknown command for module 'command': %s\n", command)
			os.Exit(1)
		}
	case "skill":
		switch command {
		case "list":
			cmdSkillList(cmdArgs)
		case "sync":
			cmdSkillSync(cmdArgs)
		case "help", "-h", "--help":
			printSkillHelp()
		default:
			fmt.Fprintf(os.Stderr, "Unknown command for module 'skill': %s\n", command)
			os.Exit(1)
		}
	default:
		unknownModule(module)
	}
}

func printModuleHelp(module string) {
	switch module {
	case "rule":
		printRuleHelp()
	case "command":
		printCommandHelp()
	case "skill":
		printSkillHelp()
	default:
		unknownModule(module)
	}
}

const (
	indentFlags = "                             "
)

func unknownModule(module string) {
	fmt.Fprintf(os.Stderr, "Unknown module: %s\n", module)
	fmt.Fprintln(os.Stderr, "\nAvailable modules: rule, command, skill")
	fmt.Fprintln(os.Stderr, "\nUsage: agents <module> <command> [flags]")
	os.Exit(1)
}

func printFlag(flag, description string) {
	fmt.Printf("%s%-10s %s\n", indentFlags, flag, description)
}

func printHelp() {
	help := `Agent Guidelines Manager CLI

Usage: agents <module> <command> [flags]

Modules:
  rule                     Manage guideline files (AGENTS.md, CLAUDE.md, etc.)
                           Commands: list, sync, rm

  command                  Manage command files for agents
                           Commands: list, sync, rm

  skill                    Manage Claude Code skills
                           Commands: list, sync

Examples:
  agents rule list
  agents rule sync --claude --cursor
  agents rule rm --claude
  agents command list
  agents command sync --claude
  agents skill list
  agents skill sync --dry-run

For module-specific help:
  agents rule help
  agents command help
  agents skill help
`

	fmt.Print(help)
}

func printRuleHelp() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Agent Guidelines Manager CLI - Rule Module")
	fmt.Println()
	fmt.Println("Usage: agents rule <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                     Discover and display all guideline files with metadata")
	fmt.Printf("%sFlags:\n", indentFlags)
	printFlag("--verbose", "Show detailed output")
	printFlag("-g", "Show only user/system-wide agent guideline files")
	printFlag("--global", "Show only user/system-wide agent guideline files")
	printFlag("--<agent>", "Filter by specific agent files (e.g., --claude, --cursor)")

	fmt.Println()
	fmt.Println("  sync                     Find all guideline source files and create symlinks")
	fmt.Printf("%sFlags:\n", indentFlags)
	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }) {
		flagName := getProviderFlagName(cfg, agent)
		guidelines := cfg.Providers[agent].Guidelines
		if guidelines != nil {
			printFlag("--"+flagName, fmt.Sprintf("Create %s symlinks", guidelines.File))
		}
	}
	printFlag("--dry-run", "Show what would be created without making changes")
	printFlag("--verbose", "Show detailed output of all operations")

	fmt.Println()
	fmt.Println("  rm                       Delete guideline files for specified agents")
	fmt.Printf("%sFlags:\n", indentFlags)
	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }) {
		flagName := getProviderFlagName(cfg, agent)
		guidelines := cfg.Providers[agent].Guidelines
		if guidelines != nil {
			printFlag("--"+flagName, fmt.Sprintf("Delete %s files", guidelines.File))
		}
	}
	printFlag("--dry-run", "Show what would be deleted without making changes")
	printFlag("--verbose", "Show detailed output of all operations")

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  agents rule list")
	fmt.Println("  agents rule list --verbose")
	fmt.Println("  agents rule list --claude")
	fmt.Println("  agents rule list --gemini --global")
	fmt.Println("  agents rule list --claude --cursor --verbose")
	fmt.Println("  agents rule sync --claude --cursor")
	fmt.Println("  agents rule sync --claude --cursor --dry-run")
	fmt.Println("  agents rule rm --claude")
	fmt.Println("  agents rule rm --cursor --gemini --dry-run")
}

func printCommandHelp() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Agent Guidelines Manager CLI - Command Module")
	fmt.Println()
	fmt.Println("Usage: agents command <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                     Discover and display all command files with metadata")
	fmt.Printf("%sFlags:\n", indentFlags)
	printFlag("--verbose", "Show detailed output")
	printFlag("--<agent>", "Filter by specific command files (e.g., --claude, --cursor)")

	fmt.Println()
	fmt.Println("  sync                     Find all command source files and create symlinks")
	fmt.Printf("%sFlags:\n", indentFlags)
	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }) {
		flagName := getProviderFlagName(cfg, agent)
		commands := cfg.Providers[agent].Commands
		if commands != nil {
			printFlag("--"+flagName, fmt.Sprintf("Create %s symlinks", commands.File))
		}
	}
	printFlag("--dry-run", "Show what would be created without making changes")
	printFlag("--verbose", "Show detailed output of all operations")

	fmt.Println()
	fmt.Println("  rm                       Delete command files for specified agents")
	fmt.Printf("%sFlags:\n", indentFlags)
	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }) {
		flagName := getProviderFlagName(cfg, agent)
		commands := cfg.Providers[agent].Commands
		if commands != nil {
			printFlag("--"+flagName, fmt.Sprintf("Delete %s files", commands.File))
		}
	}
	printFlag("--dry-run", "Show what would be deleted without making changes")
	printFlag("--verbose", "Show detailed output of all operations")

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  agents command list")
	fmt.Println("  agents command list --claude")
	fmt.Println("  agents command sync --claude --cursor")
	fmt.Println("  agents command rm --claude")
}

func printSkillHelp() {
	fmt.Println("Agent Guidelines Manager CLI - Skill Module")
	fmt.Println()
	fmt.Println("Usage: agents skill <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list                     Discover and display all Claude Code skills")
	fmt.Printf("%sFlags:\n", indentFlags)
	printFlag("--verbose", "Show detailed output including metadata")

	fmt.Println()
	fmt.Println("  sync                     Sync skills from source directory to .claude/skills")
	fmt.Printf("%sFlags:\n", indentFlags)
	printFlag("--dry-run", "Show what would be synced without making changes")
	printFlag("--verbose", "Show detailed output of all operations")

	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  agents skill list")
	fmt.Println("  agents skill list --verbose")
	fmt.Println("  agents skill sync")
	fmt.Println("  agents skill sync --dry-run --verbose")
}

func cmdRuleList(argv []string) {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--verbose"] = true
	allowedFlags["--global"] = true
	allowedFlags["-g"] = true

	for _, name := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }) {
		allowedFlags["--"+getProviderFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("rule list", parsed.Unknown)

	verbose := parsed.Flags["--verbose"]
	global := parsed.Flags["-g"] || parsed.Flags["--global"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }, parsed.Flags)
	filterAgents := selection.Selected

	var files []ManagedFile
	if global {
		files = discoverGlobalOnly(cfg)
	} else {
		files = discoverAll(cfg, cfg.Sources.Guidelines, func(p ProviderConfig) *FileSpec { return p.Guidelines })
	}

	if len(filterAgents) > 0 {
		files = filterFilesByProviders(files, filterAgents)
	}

	formatList(files, verbose, "No guideline files found.")
}

func cmdRuleSync(argv []string) {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--dry-run"] = true
	allowedFlags["--verbose"] = true

	for _, name := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }) {
		allowedFlags["--"+getProviderFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("rule sync", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }, parsed.Flags)
	selectedAgents := ensureProvidersSelected(cfg, selection.Available, selection.Selected)

	sourceFiles := discoverSources(cfg.Sources.Guidelines)
	result := syncSymlinks(
		sourceFiles,
		selectedAgents,
		cfg,
		func(p ProviderConfig) *FileSpec { return p.Guidelines },
		dryRun,
		verbose,
	)

	formatSyncSummary(cfg.Sources.Guidelines, len(sourceFiles), result.Created, result.Skipped, verbose, result.Operations)
}

func cmdRuleRm(argv []string) {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--dry-run"] = true
	allowedFlags["--verbose"] = true

	for _, name := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }) {
		allowedFlags["--"+getProviderFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("rule rm", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }, parsed.Flags)
	selectedAgents := ensureProvidersSelected(cfg, selection.Available, selection.Selected)

	deleteManagedFiles(selectedAgents, cfg, cfg.Sources.Guidelines, func(p ProviderConfig) *FileSpec { return p.Guidelines }, dryRun, verbose)
}

func cmdCommandList(argv []string) {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--verbose"] = true

	for _, name := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }) {
		allowedFlags["--"+getProviderFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("command list", parsed.Unknown)

	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }, parsed.Flags)
	filterAgents := selection.Selected

	files := discoverAll(cfg, cfg.Sources.Commands, func(p ProviderConfig) *FileSpec { return p.Commands })
	if len(filterAgents) > 0 {
		files = filterFilesByProviders(files, filterAgents)
	}

	formatList(files, verbose, "No command files found.")
}

func cmdCommandSync(argv []string) {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--dry-run"] = true
	allowedFlags["--verbose"] = true

	for _, name := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }) {
		allowedFlags["--"+getProviderFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("command sync", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }, parsed.Flags)
	selectedAgents := ensureProvidersSelected(cfg, selection.Available, selection.Selected)

	sourceFiles := discoverSources(cfg.Sources.Commands)
	result := syncSymlinks(
		sourceFiles,
		selectedAgents,
		cfg,
		func(p ProviderConfig) *FileSpec { return p.Commands },
		dryRun,
		verbose,
	)

	formatSyncSummary(cfg.Sources.Commands, len(sourceFiles), result.Created, result.Skipped, verbose, result.Operations)
}

func cmdCommandRm(argv []string) {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--dry-run"] = true
	allowedFlags["--verbose"] = true

	for _, name := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }) {
		allowedFlags["--"+getProviderFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("command rm", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }, parsed.Flags)
	selectedAgents := ensureProvidersSelected(cfg, selection.Available, selection.Selected)

	deleteManagedFiles(selectedAgents, cfg, cfg.Sources.Commands, func(p ProviderConfig) *FileSpec { return p.Commands }, dryRun, verbose)
}

func cmdSkillList(argv []string) {
	allowedFlags := make(map[string]bool)
	allowedFlags["--verbose"] = true

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("skill list", parsed.Unknown)

	verbose := parsed.Flags["--verbose"]
	skills := discoverSkills()
	formatSkillsList(skills, verbose)
}

func cmdSkillSync(argv []string) {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--dry-run"] = true
	allowedFlags["--verbose"] = true

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("skill sync", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	sourceSkillDirs := discoverSourceSkills(cfg.Sources.Skills)
	cwd, _ := os.Getwd()
	targetDir := pathJoin(cwd, ".claude", "skills")

	result := syncSkills(sourceSkillDirs, targetDir, dryRun, verbose)

	formatSkillsSyncSummary(len(sourceSkillDirs), result.Created, result.Skipped, verbose, result.Operations)
}

// Helper functions

func getProviderNames(cfg *ProvidersConfig, specSelector func(ProviderConfig) *FileSpec) []string {
	names := []string{}
	for name, provider := range cfg.Providers {
		if specSelector(provider) != nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func getProviderFlagName(cfg *ProvidersConfig, providerName string) string {
	provider, ok := cfg.Providers[providerName]
	if !ok || provider.Name == "" {
		return providerName
	}
	return provider.Name
}

type AgentSelection struct {
	Available []string
	Selected  []string
}

func collectAgentFlags(cfg *ProvidersConfig, specSelector func(ProviderConfig) *FileSpec, flags map[string]bool) AgentSelection {
	available := getProviderNames(cfg, specSelector)
	selected := []string{}

	for _, providerName := range available {
		flagName := "--" + getProviderFlagName(cfg, providerName)
		if flags[flagName] {
			selected = append(selected, providerName)
		}
	}

	return AgentSelection{
		Available: available,
		Selected:  selected,
	}
}

func ensureProvidersSelected(cfg *ProvidersConfig, available, selected []string) []string {
	if len(selected) > 0 {
		return selected
	}

	if len(available) == 0 {
		fmt.Fprintln(os.Stderr, "No agents configured.")
		os.Exit(1)
	}

	first := getProviderFlagName(cfg, available[0])
	rest := []string{}
	for i := 1; i < len(available); i++ {
		rest = append(rest, "--"+getProviderFlagName(cfg, available[i]))
	}

	msg := fmt.Sprintf("Please specify at least one agent flag: --%s", first)
	if len(rest) > 0 {
		msg += ", " + strings.Join(rest, ", ")
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
	return nil
}

func filterFilesByProviders(files []ManagedFile, providers []string) []ManagedFile {
	lower := make([]string, len(providers))
	for i, p := range providers {
		lower[i] = strings.ToLower(p)
	}

	filtered := []ManagedFile{}
	for _, file := range files {
		fileLower := strings.ToLower(file.Agent)
		for _, provider := range lower {
			if fileLower == provider {
				filtered = append(filtered, file)
				break
			}
		}
	}
	return filtered
}

func ensureNoUnknownFlags(commandName string, unknownFlags []string) {
	if len(unknownFlags) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "Unknown flags for %s: %s\n", commandName, strings.Join(unknownFlags, ", "))
	os.Exit(1)
}
