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
	case "list-commands", "list-cmds":
		cmdListCommands()
	case "sync":
		cmdSync()
	case "sync-commands", "sync-cmds":
		cmdSyncCommands()
	case "rm":
		cmdRm()
	case "rm-commands", "rm-cmds":
		cmdRmCommands()
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printHelp() {
	help := `Agent Guidelines Manager CLI

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
                           Flags:`
	fmt.Print(help)
	
	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		fmt.Printf("                             --%s       Create %s symlinks\n", cfg.Name, cfg.File)
	}
	
	help2 := `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  sync-commands            Find all COMMANDS directories and create command symlinks
  sync-cmds                Alias for sync-commands
                           Flags:`
	fmt.Print(help2)

	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		fmt.Printf("                             --%s       Create command symlinks for %s\n", cfg.Name, cfg.Name)
	}

	help2b := `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`
	fmt.Print(help2b)
	
	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		fmt.Printf("                             --%s       Delete %s files\n", cfg.Name, cfg.File)
	}
	
	help3 := `                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  rm-commands              Delete command files for specified agents
  rm-cmds                  Alias for rm-commands
                           Flags:`
	fmt.Print(help3)

	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		fmt.Printf("                             --%s       Delete command files for %s\n", cfg.Name, cfg.Name)
	}

	help4 := `                             --dry-run    Show what would be deleted without making changes
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
`
	fmt.Print(help4)
}

func cmdList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show detailed output")
	global := fs.Bool("g", false, "Show only user/system-wide agent guideline files")
	fs.BoolVar(global, "global", false, "Show only user/system-wide agent guideline files")

	// Dynamically register flags for each agent type
	agentFlags := make(map[string]*bool)
	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		agentFlags[agent] = fs.Bool(cfg.Name, false, "Filter by "+cfg.File+" files")
	}
	fs.Parse(os.Args[2:])

	// Determine which agents to filter by
	filterAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			filterAgents = append(filterAgents, agent)
		}
	}

	var files []GuidelineFile
	if *global {
		files = discoverGlobalOnly()
	} else {
		files = discoverAll()
	}

	// Filter by specified agents if any are provided
	if len(filterAgents) > 0 {
		filteredFiles := []GuidelineFile{}
		for _, f := range files {
			// Convert agent name to match the format (e.g., "claude" -> "CLAUDE")
			for _, agent := range filterAgents {
				if strings.ToUpper(agent) == f.Agent {
					filteredFiles = append(filteredFiles, f)
					break
				}
			}
		}
		files = filteredFiles
	}

	formatList(files, *verbose)
}

func cmdListCommands() {
	fs := flag.NewFlagSet("list-commands", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show detailed output")
	global := fs.Bool("g", false, "Show only user/system-wide command files")
	fs.BoolVar(global, "global", false, "Show only user/system-wide command files")

	// Dynamically register flags for each agent type
	agentFlags := make(map[string]*bool)
	for _, agent := range GetAgentNames() {
		agentFlags[agent] = fs.Bool(agent, false, "Filter by "+agent+" commands")
	}
	fs.Parse(os.Args[2:])

	// Determine which agents to filter by
	filterAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			filterAgents = append(filterAgents, agent)
		}
	}

	var files []CommandFile
	if *global {
		files = discoverGlobalCommands()
	} else {
		files = discoverAllCommands()
	}

	// Filter by specified agents if any are provided
	if len(filterAgents) > 0 {
		filteredFiles := []CommandFile{}
		for _, f := range files {
			for _, agent := range filterAgents {
				if strings.ToUpper(agent) == f.Provider {
					filteredFiles = append(filteredFiles, f)
					break
				}
			}
		}
		files = filteredFiles
	}

	formatCommandList(files, *verbose)
}

func cmdSync() {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be created without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	// Dynamically register flags for each agent type
	agentFlags := make(map[string]*bool)
	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		agentFlags[agent] = fs.Bool(cfg.Name, false, "Create "+cfg.File+" symlinks")
	}
	fs.Parse(os.Args[2:])

	// Check if at least one agent flag is specified
	atLeastOneAgent := false
	selectedAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			atLeastOneAgent = true
			selectedAgents = append(selectedAgents, agent)
		}
	}

	if !atLeastOneAgent {
		agentNames := GetAgentNames()
		fmt.Printf("Please specify at least one agent flag: --%s", agentNames[0])
		for _, name := range agentNames[1:] {
			fmt.Printf(", --%s", name)
		}
		fmt.Println()
		os.Exit(1)
	}

	agentsFiles := discoverAgents()
	syncSymlinks(agentsFiles, selectedAgents, *dryRun, *verbose)
}

func cmdRm() {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	// Dynamically register flags for each agent type
	agentFlags := make(map[string]*bool)
	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		agentFlags[agent] = fs.Bool(cfg.Name, false, "Delete "+cfg.File+" files")
	}
	fs.Parse(os.Args[2:])

	// Check if at least one agent flag is specified
	atLeastOneAgent := false
	selectedAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			atLeastOneAgent = true
			selectedAgents = append(selectedAgents, agent)
		}
	}

	if !atLeastOneAgent {
		agentNames := GetAgentNames()
		fmt.Printf("Please specify at least one agent flag: --%s", agentNames[0])
		for _, name := range agentNames[1:] {
			fmt.Printf(", --%s", name)
		}
		fmt.Println()
		os.Exit(1)
	}

	deleteGuidelineFiles(selectedAgents, *dryRun, *verbose)
}

func cmdSyncCommands() {
	fs := flag.NewFlagSet("sync-commands", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be created without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	// Dynamically register flags for each agent type
	agentFlags := make(map[string]*bool)
	for _, agent := range GetAgentNames() {
		agentFlags[agent] = fs.Bool(agent, false, "Create command symlinks for "+agent)
	}
	fs.Parse(os.Args[2:])

	// Check if at least one agent flag is specified
	atLeastOneAgent := false
	selectedAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			atLeastOneAgent = true
			selectedAgents = append(selectedAgents, agent)
		}
	}

	if !atLeastOneAgent {
		agentNames := GetAgentNames()
		fmt.Printf("Please specify at least one agent flag: --%s", agentNames[0])
		for _, name := range agentNames[1:] {
			fmt.Printf(", --%s", name)
		}
		fmt.Println()
		os.Exit(1)
	}

	commandDirs := discoverCommands()
	syncCommandFiles(commandDirs, selectedAgents, *dryRun, *verbose)
}

func cmdRmCommands() {
	fs := flag.NewFlagSet("rm-commands", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without making changes")
	verbose := fs.Bool("verbose", false, "Show detailed output of all operations")

	// Dynamically register flags for each agent type
	agentFlags := make(map[string]*bool)
	for _, agent := range GetAgentNames() {
		agentFlags[agent] = fs.Bool(agent, false, "Delete command files for "+agent)
	}
	fs.Parse(os.Args[2:])

	// Check if at least one agent flag is specified
	atLeastOneAgent := false
	selectedAgents := []string{}
	for agent, enabled := range agentFlags {
		if *enabled {
			atLeastOneAgent = true
			selectedAgents = append(selectedAgents, agent)
		}
	}

	if !atLeastOneAgent {
		agentNames := GetAgentNames()
		fmt.Printf("Please specify at least one agent flag: --%s", agentNames[0])
		for _, name := range agentNames[1:] {
			fmt.Printf(", --%s", name)
		}
		fmt.Println()
		os.Exit(1)
	}

	deleteCommandFiles(selectedAgents, *dryRun, *verbose)
}

type GuidelineFile struct {
	Path      string // full path
	Dir       string // directory containing the file
	Agent     string // AGENTS, CLAUDE, CURSOR
	File      string // filename
	IsSymlink bool
	Size      int64  // file size in bytes
}
