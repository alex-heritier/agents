package main

import (
	"fmt"
	"os"
	"strings"
)

// formatList formats and displays a list of managed files
func formatList(files []ManagedFile, verbose bool, emptyMessage string) {
	if len(files) == 0 {
		fmt.Println(emptyMessage)
		return
	}

	fmt.Printf("%-30s %-15s %-10s\n", "Directory", "File", "Tokens")
	fmt.Println(strings.Repeat("-", 55))

	cwd, _ := os.Getwd()
	homeDir := getHomeDir()

	standardGlobalPatterns := []string{
		pathJoin(homeDir, ".claude"),
		pathJoin(homeDir, ".codex"),
		pathJoin(homeDir, ".gemini"),
		pathJoin(homeDir, ".config"),
	}

	allFilesAreGlobal := true
	for _, file := range files {
		isGlobal := false
		for _, pattern := range standardGlobalPatterns {
			if strings.HasPrefix(file.Dir, pattern) {
				isGlobal = true
				break
			}
		}
		if !isGlobal {
			allFilesAreGlobal = false
			break
		}
	}

	for _, file := range files {
		displayDir := formatRelativeDir(file.Dir, cwd, homeDir, allFilesAreGlobal)
		filename := file.File
		if file.IsSymlink {
			filename = "*" + filename
		}

		tokensStr := "-"
		if !file.IsSymlink {
			tokensStr = fmt.Sprintf("%d", estimateTokens(file.Size))
		}

		fmt.Printf("%-30s %-15s %-10s\n", displayDir, filename, tokensStr)
	}

	if verbose {
		fmt.Printf("\nTotal: %d files found\n", len(files))
	}
}

// formatSyncSummary formats and displays a sync summary
func formatSyncSummary(sourceName string, found, created, skipped int, verbose bool, operations []string) {
	fmt.Printf("%s files found: %d\n", sourceName, found)
	fmt.Printf("Symlinks created: %d\n", created)
	fmt.Printf("Symlinks skipped: %d\n", skipped)

	if verbose && len(operations) > 0 {
		fmt.Println("\nOperations:")
		for _, op := range operations {
			fmt.Println(op)
		}
	}
}

// estimateTokens estimates the number of tokens in a file
func estimateTokens(size int64) int64 {
	return size / 4
}
