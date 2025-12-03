package main

import (
	"fmt"
	"os"
	"path/filepath"
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
	targetPath := filepath.Join(dir, target)

	// Check if target already exists
	if _, err := os.Lstat(targetPath); err == nil {
		if verbose {
			return 0, 1, fmt.Sprintf("skipped: %s (already exists)", targetPath)
		}
		return 0, 1, ""
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

	targetPath := filepath.Join(subdirPath, target)

	// Check if target already exists
	if _, err := os.Lstat(targetPath); err == nil {
		if verbose {
			return 0, 1, fmt.Sprintf("skipped: %s (already exists)", targetPath)
		}
		return 0, 1, ""
	}

	if dryRun {
		if verbose {
			return 1, 0, fmt.Sprintf("would create: %s -> %s", targetPath, filepath.Join("..", source))
		}
		return 1, 0, ""
	}

	// Create relative symlink back to AGENTS.md
	// From .cursor/rules/ we need to go back 2 levels to reach AGENTS.md
	symTarget := filepath.Join("..", "..", source)
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
