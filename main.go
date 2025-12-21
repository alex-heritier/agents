package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "list":
		cmdList()
	case "sync":
		cmdSync()
	case "rm":
		cmdRm()
	case "list-commands":
		cmdListCommands()
	case "sync-commands":
		cmdSyncCommands()
	case "rm-commands":
		cmdRmCommands()
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printHelp() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
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

	for _, agent := range getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	}) {
		flagName := getProviderFlagName(cfg, agent)
		guidelines := cfg.Providers[agent].Guidelines
		fmt.Printf("                             --%s       Create %s symlinks\n", flagName, guidelines.File)
	}

	help2 := `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`
	fmt.Print(help2)

	for _, agent := range getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	}) {
		flagName := getProviderFlagName(cfg, agent)
		guidelines := cfg.Providers[agent].Guidelines
		fmt.Printf("                             --%s       Delete %s files\n", flagName, guidelines.File)
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

	for _, agent := range getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	}) {
		flagName := getProviderFlagName(cfg, agent)
		commands := cfg.Providers[agent].Commands
		fmt.Printf("                             --%s       Create %s symlinks\n", flagName, commands.File)
	}

	help4 := `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm-commands               Delete command files for specified agents
                           Flags:`
	fmt.Print(help4)

	for _, agent := range getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	}) {
		flagName := getProviderFlagName(cfg, agent)
		commands := cfg.Providers[agent].Commands
		fmt.Printf("                             --%s       Delete %s files\n", flagName, commands.File)
	}

	help5 := `                             --dry-run    Show what would be deleted without making changes
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
`
	fmt.Print(help5)
}

func cmdList() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("list", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show detailed output")
	global := fs.Bool("g", false, "Show only user/system-wide agent guideline files")
	fs.BoolVar(global, "global", false, "Show only user/system-wide agent guideline files")

	agentFlags := registerProviderFlags(fs, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	}, "Filter by ")
	fs.Parse(os.Args[2:])

	filterAgents := collectSelectedProviders(agentFlags)

	var files []ManagedFile
	if *global {
		files = discoverGlobalOnly(cfg)
	} else {
		files = discoverAll(cfg, cfg.Sources.Guidelines, func(provider ProviderConfig) *FileSpec {
			return provider.Guidelines
		})
	}

	if len(filterAgents) > 0 {
		files = filterFilesByProviders(cfg, files, filterAgents)
	}

	formatList(files, *verbose, "No guideline files found.")
}

func cmdSync() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be created without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	agentFlags := registerProviderFlags(fs, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	}, "Create ")
	fs.Parse(os.Args[2:])

	availableProviders := getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	})
	selectedAgents, ok := ensureProvidersSelected(cfg, availableProviders, agentFlags)
	if !ok {
		os.Exit(1)
	}

	sourceFiles := discoverSources(cfg.Sources.Guidelines)
	created, skipped, operations := syncSymlinks(sourceFiles, selectedAgents, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	}, *dryRun, *verbose)

	formatSyncSummary(cfg.Sources.Guidelines, len(sourceFiles), created, skipped, *verbose, operations)
}

func cmdRm() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	agentFlags := registerProviderFlags(fs, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	}, "Delete ")
	fs.Parse(os.Args[2:])

	availableProviders := getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	})
	selectedAgents, ok := ensureProvidersSelected(cfg, availableProviders, agentFlags)
	if !ok {
		os.Exit(1)
	}

	deleteManagedFiles(selectedAgents, cfg, cfg.Sources.Guidelines, func(provider ProviderConfig) *FileSpec {
		return provider.Guidelines
	}, *dryRun, *verbose)
}

func cmdListCommands() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("list-commands", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show detailed output")

	agentFlags := registerProviderFlags(fs, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	}, "Filter by ")
	fs.Parse(os.Args[2:])

	filterAgents := collectSelectedProviders(agentFlags)

	files := discoverAll(cfg, cfg.Sources.Commands, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	})

	if len(filterAgents) > 0 {
		files = filterFilesByProviders(cfg, files, filterAgents)
	}

	formatList(files, *verbose, "No command files found.")
}

func cmdSyncCommands() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("sync-commands", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be created without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	agentFlags := registerProviderFlags(fs, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	}, "Create ")
	fs.Parse(os.Args[2:])

	availableProviders := getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	})
	selectedAgents, ok := ensureProvidersSelected(cfg, availableProviders, agentFlags)
	if !ok {
		os.Exit(1)
	}

	sourceFiles := discoverSources(cfg.Sources.Commands)
	created, skipped, operations := syncSymlinks(sourceFiles, selectedAgents, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	}, *dryRun, *verbose)

	formatSyncSummary(cfg.Sources.Commands, len(sourceFiles), created, skipped, *verbose, operations)
}

func cmdRmCommands() {
	cfg, err := getProviderConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("rm-commands", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	agentFlags := registerProviderFlags(fs, cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	}, "Delete ")
	fs.Parse(os.Args[2:])

	availableProviders := getProviderNames(cfg, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	})
	selectedAgents, ok := ensureProvidersSelected(cfg, availableProviders, agentFlags)
	if !ok {
		os.Exit(1)
	}

	deleteManagedFiles(selectedAgents, cfg, cfg.Sources.Commands, func(provider ProviderConfig) *FileSpec {
		return provider.Commands
	}, *dryRun, *verbose)
}

func registerProviderFlags(fs *flag.FlagSet, cfg ProvidersConfig, specSelector func(ProviderConfig) *FileSpec, descriptionPrefix string) map[string]*bool {
	agentFlags := make(map[string]*bool)
	for _, agent := range getProviderNames(cfg, specSelector) {
		flagName := getProviderFlagName(cfg, agent)
		provider := cfg.Providers[agent]
		file := ""
		if specSelector(provider) != nil {
			file = specSelector(provider).File
		}
		agentFlags[agent] = fs.Bool(flagName, false, descriptionPrefix+file+" files")
	}
	return agentFlags
}

func ensureProvidersSelected(cfg ProvidersConfig, availableProviders []string, agentFlags map[string]*bool) ([]string, bool) {
	selectedAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			selectedAgents = append(selectedAgents, agent)
		}
	}

	if len(selectedAgents) == 0 {
		if len(availableProviders) == 0 {
			fmt.Println("No agents configured.")
			return nil, false
		}
		fmt.Printf("Please specify at least one agent flag: --%s", getProviderFlagName(cfg, availableProviders[0]))
		for _, name := range availableProviders[1:] {
			fmt.Printf(", --%s", getProviderFlagName(cfg, name))
		}
		fmt.Println()
		return nil, false
	}

	return selectedAgents, true
}

func collectSelectedProviders(agentFlags map[string]*bool) []string {
	filterAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			filterAgents = append(filterAgents, agent)
		}
	}
	return filterAgents
}

func filterFilesByProviders(cfg ProvidersConfig, files []ManagedFile, providers []string) []ManagedFile {
	filteredFiles := []ManagedFile{}
	for _, f := range files {
		for _, provider := range providers {
			if strings.EqualFold(f.Agent, provider) {
				filteredFiles = append(filteredFiles, f)
				break
			}
		}
	}
	return filteredFiles
}

type ManagedFile struct {
	Path      string // full path
	Dir       string // directory containing the file
	Agent     string // AGENTS, CLAUDE, CURSOR
	File      string // filename
	IsSymlink bool
	Size      int64 // file size in bytes
}
