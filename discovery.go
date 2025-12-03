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

		// Skip ignored directories
		if info.IsDir() && ignoreDir[info.Name()] {
			return filepath.SkipDir
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
		})

		return nil
	})

	return files
}

func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}
