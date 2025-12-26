package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

func cmdRuleList(argv []string) {
	cfg, err := getToolConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--verbose"] = true
	allowedFlags["--global"] = true
	allowedFlags["-g"] = true

	for _, name := range getToolNames(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }) {
		allowedFlags["--"+getToolFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("rule list", parsed.Unknown)

	verbose := parsed.Flags["--verbose"]
	global := parsed.Flags["-g"] || parsed.Flags["--global"]

	selection := collectToolFlags(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }, parsed.Flags)
	filterAgents := selection.Selected

	var files []ManagedFile
	if global {
		files = discoverGlobalOnly(cfg)
	} else {
		files = discoverAll(cfg, getStandardGuidelineFile(cfg), func(p ToolConfig) *FileSpec { return p.Guidelines })
	}

	if len(filterAgents) > 0 {
		files = filterFilesByTools(files, filterAgents)
	}

	formatList(files, verbose, "No guideline files found.")
}

func cmdRuleSync(argv []string) {
	cfg, err := getToolConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--dry-run"] = true
	allowedFlags["--verbose"] = true

	for _, name := range getToolNames(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }) {
		allowedFlags["--"+getToolFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("rule sync", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectToolFlags(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }, parsed.Flags)
	selectedAgents := ensureToolsSelected(cfg, selection.Available, selection.Selected)

	sourceFiles := discoverSources(getStandardGuidelineFile(cfg))
	result := syncSymlinks(
		sourceFiles,
		selectedAgents,
		cfg,
		func(p ToolConfig) *FileSpec { return p.Guidelines },
		dryRun,
		verbose,
	)

	formatSyncSummary(getStandardGuidelineFile(cfg), len(sourceFiles), result.Created, result.Skipped, verbose, result.Operations)
}

func cmdRuleRm(argv []string) {
	cfg, err := getToolConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	allowedFlags := make(map[string]bool)
	allowedFlags["--dry-run"] = true
	allowedFlags["--verbose"] = true

	for _, name := range getToolNames(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }) {
		allowedFlags["--"+getToolFlagName(cfg, name)] = true
	}

	parsed := parseArgs(argv, allowedFlags)
	if parsed.Help {
		printHelp()
		os.Exit(0)
	}
	ensureNoUnknownFlags("rule rm", parsed.Unknown)

	dryRun := parsed.Flags["--dry-run"]
	verbose := parsed.Flags["--verbose"]

	selection := collectToolFlags(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }, parsed.Flags)
	selectedAgents := ensureToolsSelected(cfg, selection.Available, selection.Selected)

	deleteManagedFiles(selectedAgents, cfg, getStandardGuidelineFile(cfg), func(p ToolConfig) *FileSpec { return p.Guidelines }, dryRun, verbose)
}

func printRuleHelp() {
	cfg, err := getToolConfig()
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
	for _, agent := range getToolNames(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }) {
		flagName := getToolFlagName(cfg, agent)
		guidelines := cfg.Tools[agent].Guidelines
		if guidelines != nil {
			printFlag("--"+flagName, fmt.Sprintf("Create %s symlinks", guidelines.File))
		}
	}
	printFlag("--dry-run", "Show what would be created without making changes")
	printFlag("--verbose", "Show detailed output of all operations")

	fmt.Println()
	fmt.Println("  rm                       Delete guideline files for specified agents")
	fmt.Printf("%sFlags:\n", indentFlags)
	for _, agent := range getToolNames(cfg, func(p ToolConfig) *FileSpec { return p.Guidelines }) {
		flagName := getToolFlagName(cfg, agent)
		guidelines := cfg.Tools[agent].Guidelines
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

func getToolNames(cfg *ToolsConfig, specSelector func(ToolConfig) *FileSpec) []string {
	names := []string{}
	for name, tool := range cfg.Tools {
		if specSelector(tool) != nil {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func getToolFlagName(cfg *ToolsConfig, toolName string) string {
	tool, ok := cfg.Tools[toolName]
	if !ok || tool.Name == "" {
		return toolName
	}
	return tool.Name
}

type ToolSelection struct {
	Available []string
	Selected  []string
}

func collectToolFlags(cfg *ToolsConfig, specSelector func(ToolConfig) *FileSpec, flags map[string]bool) ToolSelection {
	available := getToolNames(cfg, specSelector)
	selected := []string{}

	for _, toolName := range available {
		flagName := "--" + getToolFlagName(cfg, toolName)
		if flags[flagName] {
			selected = append(selected, toolName)
		}
	}

	return ToolSelection{
		Available: available,
		Selected:  selected,
	}
}

func ensureToolsSelected(cfg *ToolsConfig, available, selected []string) []string {
	if len(selected) > 0 {
		return selected
	}

	if len(available) == 0 {
		fmt.Fprintln(os.Stderr, "No agents configured.")
		os.Exit(1)
	}

	first := getToolFlagName(cfg, available[0])
	rest := []string{}
	for i := 1; i < len(available); i++ {
		rest = append(rest, "--"+getToolFlagName(cfg, available[i]))
	}

	msg := fmt.Sprintf("Please specify at least one agent flag: --%s", first)
	if len(rest) > 0 {
		msg += ", " + strings.Join(rest, ", ")
	}
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
	return nil
}

func filterFilesByTools(files []ManagedFile, tools []string) []ManagedFile {
	lower := make([]string, len(tools))
	for i, p := range tools {
		lower[i] = strings.ToLower(p)
	}

	filtered := []ManagedFile{}
	for _, file := range files {
		fileLower := strings.ToLower(file.Tool)
		for _, tool := range lower {
			if fileLower == tool {
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

func getStandardGuidelineFile(cfg *ToolsConfig) string {
	standardTool := getStandardTool(cfg)
	if standardTool != nil && standardTool.Guidelines != nil {
		return standardTool.Guidelines.File
	}
	return "AGENTS.md"
}
