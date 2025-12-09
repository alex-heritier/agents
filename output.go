package main

import (
	"fmt"
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

		filename := f.File
		var tokensStr string
		if f.IsSymlink {
			filename = "*" + filename
			tokensStr = "-"
		} else {
			tokens := estimateTokens(f.Size)
			tokensStr = strconv.FormatInt(tokens, 10)
		}
		fmt.Printf("%-30s %-15s %-10s\n", relDir, filename, tokensStr)
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
