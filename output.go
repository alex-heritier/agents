package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func formatList(files []GuidelineFile, verbose bool) {
	if len(files) == 0 {
		fmt.Println("No guideline files found.")
		return
	}

	// Print header
	fmt.Printf("%-40s %-10s %-20s %-10s\n", "Directory", "Agent", "File", "Type")
	fmt.Println(strings.Repeat("-", 80))

	cwd, _ := filepath.Abs(".")
	
	for _, f := range files {
		// Make directory relative to cwd
		relDir, err := filepath.Rel(cwd, f.Dir)
		if err != nil {
			relDir = f.Dir
		}
		if relDir == "." {
			relDir = "./"
		} else if !strings.HasPrefix(relDir, ".") {
			relDir = "./" + relDir
		}

		fileType := "file"
		if f.IsSymlink {
			fileType = "symlink"
		}

		fmt.Printf("%-40s %-10s %-20s %-10s\n", relDir, f.Agent, f.File, fileType)
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
