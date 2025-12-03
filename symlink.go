package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func syncSymlinks(agents []string, selectedAgents []string, dryRun, verbose bool) {
	var created int
	var skipped int
	var operations []string

	for _, agentPath := range agents {
		dir := filepath.Dir(agentPath)
		filename := filepath.Base(agentPath)

		for _, agentName := range selectedAgents {
			cfg := SupportedAgents[agentName]

			var created_, skipped_ int
			var op string

			if cfg.Dir == "" {
				// Create symlink in same directory
				created_, skipped_, op = createSymlink(dir, filename, cfg.File, dryRun, verbose)
			} else {
				// Create symlink in subdirectory
				created_, skipped_, op = createSymlinkInDir(dir, filename, cfg.Dir, cfg.File, dryRun, verbose)
			}

			created += created_
			skipped += skipped_
			if op != "" {
				operations = append(operations, op)
			}
		}
	}

	formatSyncSummary(len(agents), created, skipped, verbose, operations)
}

func createSymlink(dir, source, target string, dryRun, verbose bool) (int, int, string) {
	sourcePath := filepath.Join(dir, source)
	targetPath := filepath.Join(dir, target)

	// Check if target already exists
	info, err := os.Lstat(targetPath)
	if err == nil {
		// File exists, check if it's the right symlink or if we should overwrite
		if shouldSkipOrOverwrite(targetPath, source, info, sourcePath, dryRun, verbose) {
			if verbose {
				return 0, 1, fmt.Sprintf("skipped: %s (already correct)", targetPath)
			}
			return 0, 1, ""
		}
		
		// User wants to overwrite
		if !dryRun {
			os.Remove(targetPath)
		}
	}

	if dryRun {
		if verbose {
			return 1, 0, fmt.Sprintf("would create: %s -> %s", targetPath, source)
		}
		return 1, 0, ""
	}

	// Create relative symlink
	if err := os.Symlink(source, targetPath); err != nil {
		if verbose {
			return 0, 1, fmt.Sprintf("error: %s (%v)", targetPath, err)
		}
		return 0, 1, ""
	}

	if verbose {
		return 1, 0, fmt.Sprintf("created: %s -> %s", targetPath, source)
	}
	return 1, 0, ""
}

func createSymlinkInDir(dir, source, subdir, target string, dryRun, verbose bool) (int, int, string) {
	// Create subdirectory if needed
	subdirPath := filepath.Join(dir, subdir)
	if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
		if !dryRun {
			os.MkdirAll(subdirPath, 0755)
		}
	}

	sourcePath := filepath.Join(dir, source)
	targetPath := filepath.Join(subdirPath, target)
	symTarget := filepath.Join("..", "..", source)

	// Check if target already exists
	info, err := os.Lstat(targetPath)
	if err == nil {
		// File exists, check if it's the right symlink or if we should overwrite
		if shouldSkipOrOverwrite(targetPath, symTarget, info, sourcePath, dryRun, verbose) {
			if verbose {
				return 0, 1, fmt.Sprintf("skipped: %s (already correct)", targetPath)
			}
			return 0, 1, ""
		}
		
		// User wants to overwrite
		if !dryRun {
			os.Remove(targetPath)
		}
	}

	if dryRun {
		if verbose {
			return 1, 0, fmt.Sprintf("would create: %s -> %s", targetPath, symTarget)
		}
		return 1, 0, ""
	}

	// Create relative symlink back to AGENTS.md
	// From .cursor/rules/ we need to go back 2 levels to reach AGENTS.md
	if err := os.Symlink(symTarget, targetPath); err != nil {
		if verbose {
			return 0, 1, fmt.Sprintf("error: %s (%v)", targetPath, err)
		}
		return 0, 1, ""
	}

	if verbose {
		return 1, 0, fmt.Sprintf("created: %s -> %s", targetPath, symTarget)
	}
	return 1, 0, ""
}

// shouldSkipOrOverwrite checks if an existing file matches the expected state
// Returns true if we should skip (file is already correct), false if we should overwrite
func shouldSkipOrOverwrite(targetPath, expectedTarget string, info os.FileInfo, sourcePath string, dryRun, verbose bool) bool {
	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		// It's a symlink, check if it points to the right place
		link, err := os.Readlink(targetPath)
		if err == nil && link == expectedTarget {
			return true // Symlink is correct
		}
		// Symlink points to wrong place, ask to overwrite
		if !dryRun && !askForConfirmation(targetPath, expectedTarget) {
			return true // User said no, skip it
		}
		return false // Overwrite
	}

	// It's a regular file, compare content with source
	sourceContent, err1 := os.ReadFile(sourcePath)
	targetContent, err2 := os.ReadFile(targetPath)
	
	if err1 != nil || err2 != nil {
		// Can't read files, ask user
		if !dryRun && !askForConfirmation(targetPath, "source") {
			return true
		}
		return false
	}

	if string(sourceContent) == string(targetContent) {
		return true // Content matches, skip
	}

	// Content differs, ask user
	if !dryRun && !askForConfirmation(targetPath, "different version") {
		return true
	}
	return false
}

// askForConfirmation prompts the user for y/n confirmation
func askForConfirmation(targetPath, reason string) bool {
	fmt.Printf("\nFile already exists: %s (%s)\n", targetPath, reason)
	fmt.Print("Overwrite? (y/n): ")
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
