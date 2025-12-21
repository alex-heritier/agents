package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CommandFile represents a slash command file
type CommandFile struct {
	Path      string // full path
	Dir       string // directory containing the file
	Provider  string // CLAUDE, CURSOR, etc.
	Name      string // command name (filename without extension)
	File      string // filename with extension
	IsSymlink bool
	Size      int64 // file size in bytes
}

// discoverCommands finds all COMMANDS directories recursively
func discoverCommands() []string {
	var commandDirs []string
	cwd, _ := os.Getwd()

	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip ignored directories
		if info.IsDir() && ignoreDir[info.Name()] {
			return filepath.SkipDir
		}

		// Match COMMANDS directories
		if info.IsDir() && info.Name() == "COMMANDS" {
			commandDirs = append(commandDirs, path)
		}

		return nil
	})

	return commandDirs
}

// discoverAllCommands finds all slash command files for all providers
func discoverAllCommands() []CommandFile {
	var files []CommandFile
	cwd, _ := os.Getwd()

	config, err := LoadConfig()
	if err != nil {
		return files
	}

	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip ignored directories (but allow provider-specific dirs)
		if info.IsDir() {
			if ignoreDir[info.Name()] {
				// Don't skip .cursor and similar provider directories
				isProviderDir := false
				for _, provider := range config.Providers {
					if strings.Contains(path, provider.Commands.Dir) {
						isProviderDir = true
						break
					}
				}
				if !isProviderDir {
					return filepath.SkipDir
				}
			}
		}

		if info.IsDir() {
			return nil
		}

		dir := filepath.Dir(path)
		filename := info.Name()

		// Check if this file matches any provider's command configuration
		for _, provider := range config.Providers {
			// Check if file is in the provider's commands directory
			if !strings.Contains(path, provider.Commands.Dir) {
				continue
			}

			// Check if file has the right extension
			if provider.Commands.Extension != "" && !strings.HasSuffix(filename, provider.Commands.Extension) {
				continue
			}

			// Extract command name (filename without extension)
			cmdName := filename
			if provider.Commands.Extension != "" {
				cmdName = strings.TrimSuffix(filename, provider.Commands.Extension)
			}

			// Check if symlink
			isSymlink := isSymlink(path)

			files = append(files, CommandFile{
				Path:      path,
				Dir:       dir,
				Provider:  strings.ToUpper(provider.Name),
				Name:      cmdName,
				File:      filename,
				IsSymlink: isSymlink,
				Size:      info.Size(),
			})
			break
		}

		return nil
	})

	return files
}

// discoverGlobalCommands finds only user/system-wide slash commands
func discoverGlobalCommands() []CommandFile {
	var files []CommandFile

	config, err := LoadConfig()
	if err != nil {
		return files
	}

	for _, provider := range config.Providers {
		expandedProvider := ExpandProviderPaths(provider)
		for _, globalPath := range expandedProvider.Commands.GlobalPaths {
			if !dirExists(globalPath) {
				continue
			}

			// Walk the global commands directory
			filepath.Walk(globalPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}

				filename := info.Name()

				// Check if file has the right extension
				if provider.Commands.Extension != "" && !strings.HasSuffix(filename, provider.Commands.Extension) {
					return nil
				}

				// Extract command name
				cmdName := filename
				if provider.Commands.Extension != "" {
					cmdName = strings.TrimSuffix(filename, provider.Commands.Extension)
				}

				isSymlink := isSymlink(path)

				files = append(files, CommandFile{
					Path:      path,
					Dir:       filepath.Dir(path),
					Provider:  strings.ToUpper(provider.Name),
					Name:      cmdName,
					File:      filename,
					IsSymlink: isSymlink,
					Size:      info.Size(),
				})

				return nil
			})
		}
	}

	return files
}

// syncCommandFiles creates command files for specified providers from COMMANDS directory
func syncCommandFiles(commandDirs []string, selectedProviders []string, dryRun, verbose bool) {
	var created int
	var skipped int
	var operations []string

	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	for _, commandDir := range commandDirs {
		parentDir := filepath.Dir(commandDir)

		// Read all files in COMMANDS directory
		entries, err := os.ReadDir(commandDir)
		if err != nil {
			if verbose {
				operations = append(operations, fmt.Sprintf("error reading %s: %v", commandDir, err))
			}
			continue
		}

		for _, providerName := range selectedProviders {
			provider, ok := config.Providers[providerName]
			if !ok {
				continue
			}

			// Create provider's commands directory
			targetDir := filepath.Join(parentDir, provider.Commands.Dir)
			if _, err := os.Stat(targetDir); os.IsNotExist(err) {
				if !dryRun {
					if err := os.MkdirAll(targetDir, 0755); err != nil {
						if verbose {
							operations = append(operations, fmt.Sprintf("error creating %s: %v", targetDir, err))
						}
						continue
					}
				}
			}

			// Sync each command file
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				sourceFile := entry.Name()

				// Skip if file doesn't have markdown extension (or matches provider extension)
				if !strings.HasSuffix(sourceFile, ".md") {
					continue
				}

				sourcePath := filepath.Join(commandDir, sourceFile)

				// Determine target filename
				targetFile := sourceFile
				if provider.Commands.Extension != "" && provider.Commands.Extension != filepath.Ext(sourceFile) {
					// Change extension if needed
					baseName := strings.TrimSuffix(sourceFile, filepath.Ext(sourceFile))
					targetFile = baseName + provider.Commands.Extension
				}

				targetPath := filepath.Join(targetDir, targetFile)

				// Calculate relative path from target to source
				relPath, err := filepath.Rel(targetDir, sourcePath)
				if err != nil {
					relPath = sourcePath
				}

				// Create symlink
				c, s, op := createCommandSymlink(targetPath, relPath, dryRun, verbose)
				created += c
				skipped += s
				if op != "" {
					operations = append(operations, op)
				}
			}
		}
	}

	formatCommandsSyncSummary(len(commandDirs), created, skipped, verbose, operations)
}

// createCommandSymlink creates a symlink for a command file
func createCommandSymlink(targetPath, sourcePath string, dryRun, verbose bool) (int, int, string) {
	// Check if target already exists
	info, err := os.Lstat(targetPath)
	if err == nil {
		// File exists, check if it's the right symlink
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(targetPath)
			if err == nil && link == sourcePath {
				// Symlink is already correct
				if verbose {
					return 0, 1, fmt.Sprintf("skipped: %s (already correct)", targetPath)
				}
				return 0, 1, ""
			}
		}

		// File exists but is not the right symlink
		if !dryRun {
			os.Remove(targetPath)
		}
	}

	if dryRun {
		if verbose {
			return 1, 0, fmt.Sprintf("would create: %s -> %s", targetPath, sourcePath)
		}
		return 1, 0, ""
	}

	// Create symlink
	if err := os.Symlink(sourcePath, targetPath); err != nil {
		if verbose {
			return 0, 1, fmt.Sprintf("error: %s (%v)", targetPath, err)
		}
		return 0, 1, ""
	}

	if verbose {
		return 1, 0, fmt.Sprintf("created: %s -> %s", targetPath, sourcePath)
	}
	return 1, 0, ""
}

// deleteCommandFiles removes command files for specified providers
func deleteCommandFiles(selectedProviders []string, dryRun, verbose bool) {
	var deleted int
	var notFound int
	var operations []string

	allFiles := discoverAllCommands()

	for _, file := range allFiles {
		// Check if this file matches any of the selected providers
		shouldDelete := false
		for _, providerName := range selectedProviders {
			if strings.EqualFold(file.Provider, providerName) {
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

// copyCommandFile copies a file from source to target
func copyCommandFile(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	return err
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// formatCommandsSyncSummary prints a summary of command sync operations
func formatCommandsSyncSummary(found, created, skipped int, verbose bool, operations []string) {
	fmt.Printf("COMMANDS directories found: %d\n", found)
	fmt.Printf("Command symlinks created: %d\n", created)
	fmt.Printf("Command symlinks skipped: %d\n", skipped)

	if verbose && len(operations) > 0 {
		fmt.Println("\nOperations:")
		for _, op := range operations {
			fmt.Println(op)
		}
	}
}
