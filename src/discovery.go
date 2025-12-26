package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ignoreDir = map[string]bool{
	"node_modules": true,
	".git":         true,
	"dist":         true,
	"build":        true,
	".cursor":      true,
}

// discoverSources finds all source files with the given name or pattern
func discoverSources(sourceName string) []string {
	cwd, err := os.Getwd()
	if err != nil {
		return []string{}
	}

	sources := []string{}
	walk(cwd, func(path string, info os.FileInfo) string {
		if info.IsDir() && ignoreDir[info.Name()] {
			return "skip"
		}
		if !info.IsDir() && matchesPattern(info.Name(), sourceName) {
			sources = append(sources, path)
		}
		return "continue"
	})

	return sources
}

// matchesPattern checks if filename matches the given pattern (supports wildcards)
func matchesPattern(filename, pattern string) bool {
	if filename == pattern {
		return true
	}

	if strings.Contains(pattern, "*") {
		matched, _ := filepath.Match(pattern, filename)
		return matched
	}

	return false
}

// matchToolPattern checks if a file matches a tool's pattern
// If pattern contains separators, it matches against the relative path.
// Otherwise, it matches against the filename (effectively ** / pattern).
func matchToolPattern(pattern, filename, relPath string) bool {
	// Normalize pattern to use OS separators
	pattern = filepath.FromSlash(pattern)

	// If pattern contains separators, match against relative path
	if strings.Contains(pattern, string(os.PathSeparator)) {
		matched, _ := filepath.Match(pattern, relPath)
		return matched
	}

	// Otherwise match against filename
	matched, _ := filepath.Match(pattern, filename)
	return matched
}

func discoverAll(cfg *ToolsConfig, sourceName string, specSelector func(ToolConfig) *FileSpec) []ManagedFile {
	cwd, err := os.Getwd()
	if err != nil {
		return []ManagedFile{}
	}

	allowedDirs := allowedToolDirs(cfg, specSelector)
	files := []ManagedFile{}

	walk(cwd, func(path string, info os.FileInfo) string {
		if info.IsDir() {
			if ignoreDir[info.Name()] && !allowedDirs[info.Name()] {
				return "skip"
			}
			return "continue"
		}

		dir := pathDirname(path)
		filename := info.Name()

		// Calculate relative path for pattern matching
		relPath, err := filepath.Rel(cwd, path)
		if err != nil {
			return "continue" // Should not happen for paths walked from cwd
		}

		var matchedTools []string
		if matchesPattern(filename, sourceName) {
			standardTool := getStandardTool(cfg)
			if standardTool != nil {
				standardSpec := specSelector(*standardTool)
				// Verify it matches the standard tool pattern
				if standardSpec != nil && standardSpec.Pattern != "" {
					if !matchToolPattern(standardSpec.Pattern, filename, relPath) {
						return "continue"
					}
				}
			}
			matchedTools = append(matchedTools, strings.ToUpper(strings.TrimSuffix(filename, ".md")))
		} else {
			// Sort keys for deterministic iteration order, though with []string it matters less
			// but good for consistency
			var toolNames []string
			for name := range cfg.Tools {
				toolNames = append(toolNames, name)
			}
			sort.Strings(toolNames)

			for _, agentName := range toolNames {
				tool := cfg.Tools[agentName]
				spec := specSelector(tool)
				if spec == nil {
					continue
				}

				if matchToolPattern(spec.Pattern, filename, relPath) {
					matchedTools = append(matchedTools, strings.ToUpper(agentName))
				}
			}
			if len(matchedTools) == 0 {
				return "continue"
			}
		}

		// Get lstat to check for symlink
		linfo, err := os.Lstat(path)
		isSymlink := false
		if err == nil {
			isSymlink = linfo.Mode()&os.ModeSymlink != 0
		}

		files = append(files, ManagedFile{
			Path:      path,
			Dir:       dir,
			Tools:     matchedTools,
			File:      filename,
			IsSymlink: isSymlink,
			Size:      info.Size(),
		})

		return "continue"
	})

	return files
}

func discoverGlobalOnly(cfg *ToolsConfig) []ManagedFile {
	files := []ManagedFile{}

	for _, location := range globalGuidelinePaths(cfg) {
		if !fileExists(location) {
			continue
		}

		info, err := os.Stat(location)
		if err != nil {
			continue
		}

		filename := pathBasename(location)
		dir := pathDirname(location)
		tools := inferToolsFromFilename(cfg, filename)
		if len(tools) == 0 {
			continue
		}

		linfo, err := os.Lstat(location)
		isSymlink := false
		if err == nil {
			isSymlink = linfo.Mode()&os.ModeSymlink != 0
		}

		files = append(files, ManagedFile{
			Path:      location,
			Dir:       dir,
			Tools:     tools,
			File:      filename,
			IsSymlink: isSymlink,
			Size:      info.Size(),
		})
	}

	return files
}

func inferToolsFromFilename(cfg *ToolsConfig, filename string) []string {
	var matchedTools []string

	standardTool := getStandardTool(cfg)
	if standardTool != nil {
		// Use base of pattern for comparison
		base := filepath.Base(filepath.FromSlash(standardTool.Pattern))
		matched, _ := filepath.Match(base, filename)
		if matched {
			matchedTools = append(matchedTools, strings.ToUpper(strings.TrimSuffix(filename, ".md")))
		}
	}

	var toolNames []string
	for name := range cfg.Tools {
		toolNames = append(toolNames, name)
	}
	sort.Strings(toolNames)

	for _, agentName := range toolNames {
		tool := cfg.Tools[agentName]
		base := filepath.Base(filepath.FromSlash(tool.Pattern))
		matched, _ := filepath.Match(base, filename)
		if matched {
			matchedTools = append(matchedTools, strings.ToUpper(agentName))
		}
	}

	return matchedTools
}

// Keep the old signature for backward compatibility if needed, but updated logic
func inferToolFromFilename(cfg *ToolsConfig, filename string) string {
	tools := inferToolsFromFilename(cfg, filename)
	if len(tools) > 0 {
		return tools[0]
	}
	return ""
}

func getStandardTool(cfg *ToolsConfig) *ToolConfig {
	if cfg.Standard == "" {
		return nil
	}
	tool, exists := cfg.Tools[cfg.Standard]
	if !exists {
		return nil
	}
	return &tool
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// isSymlink checks if a path is a symlink
func isSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// walk recursively walks a directory tree
func walk(root string, visitor func(string, os.FileInfo) string) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := pathJoin(root, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		decision := visitor(fullPath, info)
		if decision == "skip" {
			continue
		}

		if entry.IsDir() {
			walk(fullPath, visitor)
		}
	}
}

func allowedToolDirs(cfg *ToolsConfig, specSelector func(ToolConfig) *FileSpec) map[string]bool {
	allowed := make(map[string]bool)

	for _, tool := range cfg.Tools {
		spec := specSelector(tool)
		if spec == nil || spec.Pattern == "" {
			continue
		}

		// Normalize pattern
		pattern := filepath.FromSlash(spec.Pattern)

		// If pattern contains separators, check if it starts with an allowed dir
		if strings.Contains(pattern, string(os.PathSeparator)) {
			parts := strings.Split(pattern, string(os.PathSeparator))
			if len(parts) > 1 && parts[0] != "" && !strings.Contains(parts[0], "*") {
				allowed[parts[0]] = true
			}
		}
	}

	return allowed
}

func globalGuidelinePaths(cfg *ToolsConfig) []string {
	var paths []string

	for _, tool := range cfg.Tools {
		for _, path := range tool.Global {
			paths = append(paths, expandHomePath(path))
		}
	}

	return paths
}
