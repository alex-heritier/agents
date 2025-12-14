package main

import (
	"os"
	"path/filepath"
	"strings"
)

var ignoreDir = map[string]bool{
	"node_modules": true,
	".git":         true,
	"dist":         true,
	"build":        true,
	".cursor":      true,
}

// discoverAgents finds all AGENTS.md files recursively from current directory
func discoverAgents() []string {
	var agents []string
	cwd, _ := os.Getwd()
	
	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip ignored directories
		if info.IsDir() && ignoreDir[info.Name()] {
			return filepath.SkipDir
		}

		// Match AGENTS.md files
		if !info.IsDir() && info.Name() == "AGENTS.md" {
			agents = append(agents, path)
		}

		return nil
	})

	return agents
}

// discoverAll finds all guideline files (AGENTS.md, CLAUDE.md, .cursor/rules/*)
func discoverAll() []GuidelineFile {
	var files []GuidelineFile
	cwd, _ := os.Getwd()

	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip ignored directories (but not .cursor since we need to discover cursor guidelines)
		if info.IsDir() {
			if ignoreDir[info.Name()] && info.Name() != ".cursor" {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			return nil
		}

		dir := filepath.Dir(path)
		filename := info.Name()

		// Determine agent type
		var agent string
		if filename == "AGENTS.md" {
			agent = "AGENTS"
		} else {
			// Check if this matches any agent configuration
			for agentName, cfg := range SupportedAgents {
				if filename == cfg.File {
					// Check if it's in the right directory (if specified)
					if cfg.Dir == "" {
						agent = strings.ToUpper(agentName)
						break
					} else if strings.Contains(path, cfg.Dir) {
						agent = strings.ToUpper(agentName)
						break
					}
				}
			}
			if agent == "" {
				return nil
			}
		}

		// Check if symlink
		isSymlink := isSymlink(path)

		files = append(files, GuidelineFile{
			Path:      path,
			Dir:       dir,
			Agent:     agent,
			File:      filename,
			IsSymlink: isSymlink,
			Size:      info.Size(),
		})

		return nil
	})

	return files
}

// globalGuidelinePaths returns the standard locations for global agent guideline files
func globalGuidelinePaths() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{} // Return empty if we can't get home directory
	}
	
	return []string{
		filepath.Join(homeDir, ".claude", "CLAUDE.md"),
		filepath.Join(homeDir, ".codex", "AGENTS.md"),
		filepath.Join(homeDir, ".gemini", "GEMINI.md"),
		filepath.Join(homeDir, ".config", "opencode", "AGENTS.md"),
		filepath.Join(homeDir, ".config", "amp", "AGENTS.md"),
		filepath.Join(homeDir, ".config", "AGENTS.md"),
		filepath.Join(homeDir, "AGENTS.md"),
	}
}

// discoverGlobalOnly finds only user/system-wide agent guideline files
func discoverGlobalOnly() []GuidelineFile {
	var files []GuidelineFile
	
	globalLocations := globalGuidelinePaths()
	if len(globalLocations) == 0 {
		return files // If we can't get home directory, return empty
	}
	
	for _, location := range globalLocations {
		if fileExists(location) {
			info, err := os.Stat(location)
			if err != nil {
				continue
			}
			
			filename := filepath.Base(location)
			dir := filepath.Dir(location)
			
			// Determine agent type
			agent := inferAgentFromFilename(filename)
			if agent == "" {
				continue
			}
			
			// Check if symlink
			isSymlink := isSymlink(location)
			
			files = append(files, GuidelineFile{
				Path:      location,
				Dir:       dir,
				Agent:     agent,
				File:      filename,
				IsSymlink: isSymlink,
				Size:      info.Size(),
			})
		}
	}
	
	return files
}

// inferAgentFromFilename determines the agent type from a filename
func inferAgentFromFilename(filename string) string {
	if filename == "AGENTS.md" {
		return "AGENTS"
	}
	
	// Check if this matches any agent configuration
	for agentName, cfg := range SupportedAgents {
		if filename == cfg.File {
			return strings.ToUpper(agentName)
		}
	}
	
	return ""
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}
