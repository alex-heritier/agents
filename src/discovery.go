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

		agent := ""
		if matchesPattern(filename, sourceName) {
			standardTool := getStandardTool(cfg)
			if standardTool != nil {
				standardSpec := specSelector(*standardTool)
				if standardSpec != nil && standardSpec.Dir != "" && !strings.Contains(path, standardSpec.Dir) {
					return "continue"
				}
			}
			agent = strings.ToUpper(strings.TrimSuffix(filename, ".md"))
		} else {
			for agentName, tool := range cfg.Tools {
				spec := specSelector(tool)
				if spec == nil {
					continue
				}
				if matchesPattern(filename, spec.File) {
					if spec.Dir == "" {
						agent = strings.ToUpper(agentName)
						break
					}
					if strings.Contains(path, spec.Dir) {
						agent = strings.ToUpper(agentName)
						break
					}
				}
			}
			if agent == "" {
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
			Tool:      agent,
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
		agent := inferToolFromFilename(cfg, filename)
		if agent == "" {
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
			Tool:      agent,
			File:      filename,
			IsSymlink: isSymlink,
			Size:      info.Size(),
		})
	}

	return files
}

func inferToolFromFilename(cfg *ToolsConfig, filename string) string {
	standardTool := getStandardTool(cfg)
	if standardTool != nil && standardTool.Guidelines != nil && filename == standardTool.Guidelines.File {
		return strings.ToUpper(strings.TrimSuffix(filename, ".md"))
	}

	for agentName, tool := range cfg.Tools {
		if tool.Guidelines != nil && filename == tool.Guidelines.File {
			return strings.ToUpper(agentName)
		}
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
		if spec == nil || spec.Dir == "" {
			continue
		}

		parts := strings.Split(spec.Dir, "/")
		if len(parts) > 0 {
			allowed[parts[0]] = true
		}
	}

	return allowed
}

func globalGuidelinePaths(cfg *ToolsConfig) []string {
	var paths []string

	for _, tool := range cfg.Tools {
		if tool.Guidelines != nil {
			for _, path := range tool.Guidelines.Global {
				paths = append(paths, expandHomePath(path))
			}
		}
	}

	return paths
}
