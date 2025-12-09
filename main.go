package main

import (
	"flag"
	"fmt"
	"os"
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

  sync                     Find all AGENTS.md files and create symlinks
                           Flags:`
	fmt.Print(help)
	
	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		fmt.Printf("                             --%s       Create %s symlinks\n", cfg.Name, cfg.File)
	}
	
	help2 := `                             --dry-run    Show what would be created without making changes
                             --verbose    Show detailed output of all operations

  rm                       Delete guideline files for specified agents
                           Flags:`
	fmt.Print(help2)
	
	for _, agent := range GetAgentNames() {
		cfg := SupportedAgents[agent]
		fmt.Printf("                             --%s       Delete %s files\n", cfg.Name, cfg.File)
	}
	
	help3 := `                             --dry-run    Show what would be deleted without making changes
                             --verbose    Show detailed output of all operations

  help                     Show this help message

Examples:
  agents list
  agents list --verbose
  agents sync --claude --cursor
  agents sync --claude --cursor --dry-run
  agents rm --claude
  agents rm --cursor --gemini --dry-run
`
	fmt.Print(help3)
}

func cmdList() {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	verbose := fs.Bool("verbose", false, "Show detailed output")
	fs.Parse(os.Args[2:])

	files := discoverAll()
	formatList(files, *verbose)
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

type GuidelineFile struct {
	Path      string // full path
	Dir       string // directory containing the file
	Agent     string // AGENTS, CLAUDE, CURSOR
	File      string // filename
	IsSymlink bool
	Size      int64  // file size in bytes
}
