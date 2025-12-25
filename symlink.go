package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SyncResult holds the result of a sync operation
type SyncResult struct {
	Created    int
	Skipped    int
	Operations []string
}

// syncSymlinks creates symlinks for the specified providers
func syncSymlinks(
	sources []string,
	selectedProviders []string,
	cfg *ProvidersConfig,
	specSelector func(ProviderConfig) *FileSpec,
	dryRun bool,
	verbose bool,
) SyncResult {
	result := SyncResult{
		Created:    0,
		Skipped:    0,
		Operations: []string{},
	}

	for _, sourcePath := range sources {
		dir := pathDirname(sourcePath)
		filename := pathBasename(sourcePath)

		for _, providerName := range selectedProviders {
			provider, ok := cfg.Providers[providerName]
			if !ok {
				continue
			}

			spec := specSelector(provider)
			if spec == nil {
				continue
			}

			var opResult SyncResult
			if spec.Dir == "" {
				opResult = createSymlink(dir, filename, spec.File, dryRun, verbose)
			} else {
				opResult = createSymlinkInDir(dir, filename, spec.Dir, spec.File, dryRun, verbose)
			}

			result.Created += opResult.Created
			result.Skipped += opResult.Skipped
			result.Operations = append(result.Operations, opResult.Operations...)
		}
	}

	return result
}

// deleteManagedFiles deletes managed files for selected providers
func deleteManagedFiles(
	selectedProviders []string,
	cfg *ProvidersConfig,
	sourceName string,
	specSelector func(ProviderConfig) *FileSpec,
	dryRun bool,
	verbose bool,
) {
	deleted := 0
	notFound := 0
	operations := []string{}

	allFiles := discoverAll(cfg, sourceName, specSelector)

	for _, file := range allFiles {
		shouldDelete := false
		for _, provider := range selectedProviders {
			if strings.EqualFold(file.Agent, provider) {
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
			continue
		}

		err := os.Remove(file.Path)
		if err != nil {
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

	formatRmSummary(deleted, notFound, verbose, operations)
}

// createSymlink creates a symlink in the same directory
func createSymlink(dir, source, target string, dryRun, verbose bool) SyncResult {
	result := SyncResult{}
	sourcePath := pathJoin(dir, source)
	targetPath := pathJoin(dir, target)

	if exists(targetPath) {
		if shouldSkipOrOverwrite(targetPath, source, sourcePath, dryRun) {
			result.Skipped = 1
			if verbose {
				result.Operations = append(result.Operations, fmt.Sprintf("skipped: %s (already correct)", targetPath))
			}
			return result
		}
		if !dryRun {
			os.Remove(targetPath)
		}
	}

	if dryRun {
		result.Created = 1
		if verbose {
			result.Operations = append(result.Operations, fmt.Sprintf("would create: %s -> %s", targetPath, source))
		}
		return result
	}

	err := os.Symlink(source, targetPath)
	if err != nil {
		result.Skipped = 1
		if verbose {
			result.Operations = append(result.Operations, fmt.Sprintf("error: %s (%v)", targetPath, err))
		}
	} else {
		result.Created = 1
		if verbose {
			result.Operations = append(result.Operations, fmt.Sprintf("created: %s -> %s", targetPath, source))
		}
	}

	return result
}

// createSymlinkInDir creates a symlink in a subdirectory
func createSymlinkInDir(dir, source, subdir, target string, dryRun, verbose bool) SyncResult {
	result := SyncResult{}
	subdirPath := pathJoin(dir, subdir)

	if !exists(subdirPath) && !dryRun {
		os.MkdirAll(subdirPath, 0755)
	}

	sourcePath := pathJoin(dir, source)
	targetPath := pathJoin(subdirPath, target)
	symTarget := pathRelative(subdirPath, sourcePath)

	if exists(targetPath) {
		if shouldSkipOrOverwrite(targetPath, symTarget, sourcePath, dryRun) {
			result.Skipped = 1
			if verbose {
				result.Operations = append(result.Operations, fmt.Sprintf("skipped: %s (already correct)", targetPath))
			}
			return result
		}
		if !dryRun {
			os.Remove(targetPath)
		}
	}

	if dryRun {
		result.Created = 1
		if verbose {
			result.Operations = append(result.Operations, fmt.Sprintf("would create: %s -> %s", targetPath, symTarget))
		}
		return result
	}

	err := os.Symlink(symTarget, targetPath)
	if err != nil {
		result.Skipped = 1
		if verbose {
			result.Operations = append(result.Operations, fmt.Sprintf("error: %s (%v)", targetPath, err))
		}
	} else {
		result.Created = 1
		if verbose {
			result.Operations = append(result.Operations, fmt.Sprintf("created: %s -> %s", targetPath, symTarget))
		}
	}

	return result
}

// shouldSkipOrOverwrite checks if a symlink should be skipped or overwritten
func shouldSkipOrOverwrite(targetPath, expectedTarget, sourcePath string, dryRun bool) bool {
	if !exists(targetPath) {
		return false
	}

	// Check if it's a symlink
	if isSymlink(targetPath) {
		link, err := os.Readlink(targetPath)
		if err == nil && link == expectedTarget {
			return true
		}

		if !dryRun && !askForConfirmation(targetPath, expectedTarget) {
			return true
		}
		return false
	}

	// Check if content matches
	sourceContent := safeReadFile(sourcePath)
	targetContent := safeReadFile(targetPath)

	if sourceContent == nil || targetContent == nil {
		if !dryRun && !askForConfirmation(targetPath, "source") {
			return true
		}
		return false
	}

	if bytes.Equal(sourceContent, targetContent) {
		return true
	}

	if !dryRun && !askForConfirmation(targetPath, "different version") {
		return true
	}
	return false
}

// askForConfirmation prompts the user for confirmation
func askForConfirmation(targetPath, reason string) bool {
	fmt.Printf("\nFile already exists: %s (%s)\nOverwrite? (y/n): ", targetPath, reason)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	normalized := strings.ToLower(strings.TrimSpace(response))
	return normalized == "y" || normalized == "yes"
}

// safeReadFile safely reads a file
func safeReadFile(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return content
}

// exists checks if a path exists
func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// formatRmSummary formats the removal summary
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

// syncSkills copies skill directories to the target location
func syncSkills(sourceSkillDirs []string, targetBaseDir string, dryRun, verbose bool) SyncResult {
	result := SyncResult{
		Created:    0,
		Skipped:    0,
		Operations: []string{},
	}

	// Ensure target directory exists
	if !exists(targetBaseDir) && !dryRun {
		os.MkdirAll(targetBaseDir, 0755)
	}

	for _, sourceDir := range sourceSkillDirs {
		skillName := pathBasename(sourceDir)
		targetDir := pathJoin(targetBaseDir, skillName)

		if exists(targetDir) {
			result.Skipped++
			if verbose {
				result.Operations = append(result.Operations, fmt.Sprintf("skipped: %s (already exists)", targetDir))
			}
			continue
		}

		if dryRun {
			result.Created++
			if verbose {
				result.Operations = append(result.Operations, fmt.Sprintf("would sync: %s -> %s", sourceDir, targetDir))
			}
			continue
		}

		err := copyDir(sourceDir, targetDir)
		if err != nil {
			result.Skipped++
			if verbose {
				result.Operations = append(result.Operations, fmt.Sprintf("error: %s (%v)", sourceDir, err))
			}
		} else {
			result.Created++
			if verbose {
				result.Operations = append(result.Operations, fmt.Sprintf("synced: %s -> %s", sourceDir, targetDir))
			}
		}
	}

	return result
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}
