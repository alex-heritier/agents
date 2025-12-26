package main

import (
	"fmt"
	"os"
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
	default:
		unknownModule(module)
	}
}

func printModuleHelp(module string) {
	switch module {
	case "rule":
		printRuleHelp()
	default:
		unknownModule(module)
	}
}

const (
	indentFlags = "                             "
)

func unknownModule(module string) {
	fmt.Fprintf(os.Stderr, "Unknown module: %s\n", module)
	fmt.Fprintln(os.Stderr, "\nAvailable modules: rule")
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

Examples:
  agents rule list
  agents rule sync --claude --cursor
  agents rule rm --claude

For module-specific help:
  agents rule help
`

	fmt.Print(help)
}
