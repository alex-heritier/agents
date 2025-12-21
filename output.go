package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"strconv"
)

func estimateTokens(size int64) int64 {
	return size / 4
}

func formatList(files []GuidelineFile, verbose bool) {
	if len(files) == 0 {
		fmt.Println("No guideline files found.")
		return
	}

	// Print header
	fmt.Printf("%-30s %-15s %-10s\n", "Directory", "File", "Tokens")
	fmt.Println(strings.Repeat("-", 55))

	cwd, _ := filepath.Abs(".")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "" // If we can't get home directory, disable ~-relative display
	}
	
	// Determine if we should show paths relative to home directory (~)
	// We do this when all files are in standard global agent locations
	showRelativeToHome := false
	if len(files) > 0 && homeDir != "" {
		// Standard global agent directories
		standardGlobalPatterns := []string{
			filepath.Join(homeDir, ".claude"),
			filepath.Join(homeDir, ".codex"),
			filepath.Join(homeDir, ".gemini"),
			filepath.Join(homeDir, ".config"),
		}
		
		allFilesAreGlobal := true
		for _, f := range files {
			isGlobal := false
			for _, pattern := range standardGlobalPatterns {
				if strings.HasPrefix(f.Dir, pattern) {
					isGlobal = true
					break
				}
			}
			if !isGlobal {
				allFilesAreGlobal = false
				break
			}
		}
		
		if allFilesAreGlobal {
			showRelativeToHome = true
		}
	}
	
	for _, f := range files {
		var displayDir string
		
		if showRelativeToHome && homeDir != "" && strings.HasPrefix(f.Dir, homeDir) {
			// Make directory relative to home directory and prefix with ~/
			relDir, err := filepath.Rel(homeDir, f.Dir)
			if err != nil || relDir == "." {
				displayDir = "~/"
			} else {
				displayDir = "~/" + relDir
			}
		} else {
			// Make directory relative to cwd
			relDir, err := filepath.Rel(cwd, f.Dir)
			if err != nil {
				displayDir = f.Dir
			} else if relDir == "." {
				displayDir = "./"
			} else if !strings.HasPrefix(relDir, ".") {
				displayDir = "./" + relDir
			} else {
				displayDir = relDir + "/"
			}
		}

		filename := f.File
		var tokensStr string
		if f.IsSymlink {
			filename = "*" + filename
			tokensStr = "-"
		} else {
			tokens := estimateTokens(f.Size)
			tokensStr = strconv.FormatInt(tokens, 10)
		}
		fmt.Printf("%-30s %-15s %-10s\n", displayDir, filename, tokensStr)
	}

	if verbose {
		fmt.Printf("\nTotal: %d files found\n", len(files))
	}
}

func formatSyncSummary(found, created, skipped int, verbose bool, operations []string) {
	fmt.Printf("AGENTS.md files found: %d\n", found)
	fmt.Printf("Symlinks created: %d\n", created)
	fmt.Printf("Symlinks skipped: %d\n", skipped)

	if verbose && len(operations) > 0 {
		fmt.Println("\nOperations:")
		for _, op := range operations {
			fmt.Println(op)
		}
	}
}

// formatCommandList formats and displays a list of command files
func formatCommandList(files []CommandFile, verbose bool) {
	if len(files) == 0 {
		fmt.Println("No command files found.")
		return
	}

	// Print header
	fmt.Printf("%-30s %-20s %-15s %-10s\n", "Directory", "Command", "Provider", "Tokens")
	fmt.Println(strings.Repeat("-", 75))

	cwd, _ := filepath.Abs(".")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	// Determine if we should show paths relative to home directory
	showRelativeToHome := false
	if len(files) > 0 && homeDir != "" {
		standardGlobalPatterns := []string{
			filepath.Join(homeDir, ".claude"),
			filepath.Join(homeDir, ".cursor"),
			filepath.Join(homeDir, ".config"),
		}

		allFilesAreGlobal := true
		for _, f := range files {
			isGlobal := false
			for _, pattern := range standardGlobalPatterns {
				if strings.HasPrefix(f.Dir, pattern) {
					isGlobal = true
					break
				}
			}
			if !isGlobal {
				allFilesAreGlobal = false
				break
			}
		}

		if allFilesAreGlobal {
			showRelativeToHome = true
		}
	}

	for _, f := range files {
		var displayDir string

		if showRelativeToHome && homeDir != "" && strings.HasPrefix(f.Dir, homeDir) {
			relDir, err := filepath.Rel(homeDir, f.Dir)
			if err != nil || relDir == "." {
				displayDir = "~/"
			} else {
				displayDir = "~/" + relDir
			}
		} else {
			relDir, err := filepath.Rel(cwd, f.Dir)
			if err != nil {
				displayDir = f.Dir
			} else if relDir == "." {
				displayDir = "./"
			} else if !strings.HasPrefix(relDir, ".") {
				displayDir = "./" + relDir
			} else {
				displayDir = relDir + "/"
			}
		}

		cmdName := f.Name
		var tokensStr string
		if f.IsSymlink {
			cmdName = "*" + cmdName
			tokensStr = "-"
		} else {
			tokens := estimateTokens(f.Size)
			tokensStr = strconv.FormatInt(tokens, 10)
		}
		fmt.Printf("%-30s %-20s %-15s %-10s\n", displayDir, cmdName, f.Provider, tokensStr)
	}

	if verbose {
		fmt.Printf("\nTotal: %d command files found\n", len(files))
	}
}
