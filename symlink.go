package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func syncSymlinks(sources []string, selectedProviders []string, cfg ProvidersConfig, specSelector func(ProviderConfig) *FileSpec, dryRun, verbose bool) (int, int, []string) {
	var created int
	var skipped int
	var operations []string

	for _, sourcePath := range sources {
		dir := filepath.Dir(sourcePath)
		filename := filepath.Base(sourcePath)

		for _, providerName := range selectedProviders {
			provider := cfg.Providers[providerName]
			spec := specSelector(provider)
			if spec == nil {
				continue
			}

			var created_, skipped_ int
			var op string

			if spec.Dir == "" {
				created_, skipped_, op = createSymlink(dir, filename, spec.File, dryRun, verbose)
			} else {
				created_, skipped_, op = createSymlinkInDir(dir, filename, spec.Dir, spec.File, dryRun, verbose)
			}

			created += created_
			skipped += skipped_
			if op != "" {
				operations = append(operations, op)
			}
		}
	}

	return created, skipped, operations
}

func createSymlink(dir, source, target string, dryRun, verbose bool) (int, int, string) {
	sourcePath := filepath.Join(dir, source)
	targetPath := filepath.Join(dir, target)

	info, err := os.Lstat(targetPath)
	if err == nil {
		if shouldSkipOrOverwrite(targetPath, source, info, sourcePath, dryRun, verbose) {
			if verbose {
				return 0, 1, fmt.Sprintf("skipped: %s (already correct)", targetPath)
			}
			return 0, 1, ""
		}

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
	subdirPath := filepath.Join(dir, subdir)
	if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
		if !dryRun {
			os.MkdirAll(subdirPath, 0755)
		}
	}

	sourcePath := filepath.Join(dir, source)
	targetPath := filepath.Join(subdirPath, target)

	symTarget, err := filepath.Rel(subdirPath, sourcePath)
	if err != nil {
		symTarget = filepath.Join("..", source)
	}

	info, err := os.Lstat(targetPath)
	if err == nil {
		if shouldSkipOrOverwrite(targetPath, symTarget, info, sourcePath, dryRun, verbose) {
			if verbose {
				return 0, 1, fmt.Sprintf("skipped: %s (already correct)", targetPath)
			}
			return 0, 1, ""
		}

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
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(targetPath)
		if err == nil && link == expectedTarget {
			return true
		}
		if !dryRun && !askForConfirmation(targetPath, expectedTarget) {
			return true
		}
		return false
	}

	sourceContent, err1 := os.ReadFile(sourcePath)
	targetContent, err2 := os.ReadFile(targetPath)

	if err1 != nil || err2 != nil {
		if !dryRun && !askForConfirmation(targetPath, "source") {
			return true
		}
		return false
	}

	if string(sourceContent) == string(targetContent) {
		return true
	}

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

// deleteManagedFiles removes files for specified agents
func deleteManagedFiles(selectedAgents []string, cfg ProvidersConfig, sourceName string, specSelector func(ProviderConfig) *FileSpec, dryRun, verbose bool) {
	var deleted int
	var notFound int
	var operations []string

	allFiles := discoverAll(cfg, sourceName, specSelector)

	for _, file := range allFiles {
		shouldDelete := false
		for _, agentName := range selectedAgents {
			if strings.EqualFold(file.Agent, agentName) {
				shouldDelete = true
				break
			}
		}

		if !shouldDelete {
			continue
		}

		if dryRun {
			deleted++
			if verbose {
				operations = append(operations, fmt.Sprintf("would delete: %s", file.Path))
			}
		} else {
			if err := os.Remove(file.Path); err != nil {
				notFound++
				if verbose {
					operations = append(operations, fmt.Sprintf("error: %s (%v)", file.Path, err))
				}
			} else {
				deleted++
				if verbose {
					operations = append(operations, fmt.Sprintf("deleted: %s", file.Path))
				}
			}
		}
	}

	formatRmSummary(deleted, notFound, verbose, operations)
}

// formatRmSummary prints a summary of deletion operations
func formatRmSummary(deleted, notFound int, verbose bool, operations []string) {
	fmt.Printf("Files deleted: %d\n", deleted)
	if notFound > 0 {
		fmt.Printf("Errors: %d\n", notFound)
	}

	if verbose && len(operations) > 0 {
		fmt.Println("Operations:")
		for _, op := range operations {
			fmt.Println(op)
		}
	}
}
