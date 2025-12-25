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

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "list":
		cmdList(cmdArgs)
	case "sync":
		cmdSync(cmdArgs)
	case "rm":
		cmdRm(cmdArgs)
	case "list-commands":
		cmdListCommands(cmdArgs)
	case "sync-commands":
		cmdSyncCommands(cmdArgs)
	case "rm-commands":
		cmdRmCommands(cmdArgs)
	case "list-skills":
		cmdListSkills(cmdArgs)
	case "sync-skills":
		cmdSyncSkills(cmdArgs)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printHelp() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	help := `Agent Guidelines Manager CLI

Usage: agents <command> [flags]

Commands:
  list                     Discover and display all guideline files with metadata
                           Flags:
                             --verbose    Show detailed output
                             -g           Show only user/system-wide agent guideline files
                             --global     Show only user/system-wide agent guideline files
                             --<agent>    Filter by specific agent files (e.g., --claude, --cursor)

  sync                     Find all guideline source files and create symlinks
                           Flags:`

	fmt.Print(help)

	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }) {
		flagName := getProviderFlagName(cfg, agent)
		guidelines := cfg.Providers[agent].Guidelines
		if guidelines != nil {
			fmt.Printf("                             --%-10s Create %s symlinks\n", flagName, guidelines.File)
		}
	}

	help2 := `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`

	fmt.Print(help2)

	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }) {
		flagName := getProviderFlagName(cfg, agent)
		guidelines := cfg.Providers[agent].Guidelines
		if guidelines != nil {
			fmt.Printf("                             --%-10s Delete %s files\n", flagName, guidelines.File)
		}
	}

	help3 := `                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  list-commands             Discover and display all command files with metadata
                           Flags:
                             --verbose    Show detailed output
                             --<agent>    Filter by specific command files (e.g., --claude, --cursor)

  sync-commands             Find all command source files and create symlinks
                           Flags:`

	fmt.Print(help3)

	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }) {
		flagName := getProviderFlagName(cfg, agent)
		commands := cfg.Providers[agent].Commands
		if commands != nil {
			fmt.Printf("                             --%-10s Create %s symlinks\n", flagName, commands.File)
		}
	}

	help4 := `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm-commands               Delete command files for specified agents
                           Flags:`

	fmt.Print(help4)

	for _, agent := range getProviderNames(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }) {
		flagName := getProviderFlagName(cfg, agent)
		commands := cfg.Providers[agent].Commands
		if commands != nil {
			fmt.Printf("                             --%-10s Delete %s files\n", flagName, commands.File)
		}
	}

	help5 := `                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  list-skills              Discover and display all Claude Code skills
                           Flags:
                             --verbose    Show detailed output including metadata

  sync-skills              Sync skills from source directory to .claude/skills
                           Flags:
                             --dry-run    Show what would be synced without making changes
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
  agents list-skills
  agents list-skills --verbose
  agents sync-skills
  agents sync-skills --dry-run --verbose
`

	fmt.Print(help5)
}

func cmdList(argv []string) {
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
	ensureNoUnknownFlags("list", parsed.Unknown)

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

func cmdSync(argv []string) {
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
	ensureNoUnknownFlags("sync", parsed.Unknown)

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

func cmdRm(argv []string) {
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
	ensureNoUnknownFlags("rm", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Guidelines }, parsed.Flags)
	selectedAgents := ensureProvidersSelected(cfg, selection.Available, selection.Selected)

	deleteManagedFiles(selectedAgents, cfg, cfg.Sources.Guidelines, func(p ProviderConfig) *FileSpec { return p.Guidelines }, dryRun, verbose)
}

func cmdListCommands(argv []string) {
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
	ensureNoUnknownFlags("list-commands", parsed.Unknown)

	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }, parsed.Flags)
	filterAgents := selection.Selected

	files := discoverAll(cfg, cfg.Sources.Commands, func(p ProviderConfig) *FileSpec { return p.Commands })
	if len(filterAgents) > 0 {
		files = filterFilesByProviders(files, filterAgents)
	}

	formatList(files, verbose, "No command files found.")
}

func cmdSyncCommands(argv []string) {
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
	ensureNoUnknownFlags("sync-commands", parsed.Unknown)

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

func cmdRmCommands(argv []string) {
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
	ensureNoUnknownFlags("rm-commands", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectAgentFlags(cfg, func(p ProviderConfig) *FileSpec { return p.Commands }, parsed.Flags)
	selectedAgents := ensureProvidersSelected(cfg, selection.Available, selection.Selected)

	deleteManagedFiles(selectedAgents, cfg, cfg.Sources.Commands, func(p ProviderConfig) *FileSpec { return p.Commands }, dryRun, verbose)
}

func cmdListSkills(argv []string) {
	allowedFlags := make(map[string]bool)
	allowedFlags["--verbose"] = true

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("list-skills", parsed.Unknown)

	verbose := parsed.Flags["--verbose"]
	skills := discoverSkills()
	formatSkillsList(skills, verbose)
}

func cmdSyncSkills(argv []string) {
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
	ensureNoUnknownFlags("sync-skills", parsed.Unknown)

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
