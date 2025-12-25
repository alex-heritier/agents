package main

import (
	"os"
	"path/filepath"
	"strings"
)

// pathBasename returns the last element of path
func pathBasename(path string) string {
	return filepath.Base(path)
}

// pathDirname returns all but the last element of path
func pathDirname(path string) string {
	return filepath.Dir(path)
}

// pathJoin joins path elements
func pathJoin(parts ...string) string {
	return filepath.Join(parts...)
}

// pathRelative returns the relative path from 'from' to 'to'
func pathRelative(from, to string) string {
	rel, err := filepath.Rel(from, to)
	if err != nil {
		return to
	}
	return rel
}

// getHomeDir returns the user's home directory
func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}

// expandHomePath expands ~ in paths to the home directory
func expandHomePath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home := getHomeDir()
	if home == "" {
		return path
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return pathJoin(home, path[2:])
	}
	return path
}

// formatRelativeDir formats a directory path relative to cwd or home
func formatRelativeDir(targetDir, cwd, homeDir string, showRelativeToHome bool) string {
	if showRelativeToHome && homeDir != "" && strings.HasPrefix(targetDir, homeDir) {
		relDir := pathRelative(homeDir, targetDir)
		if relDir == "" || relDir == "." {
			return "~/"
		}
		return "~/" + relDir
	}
	relDir := pathRelative(cwd, targetDir)
	if relDir == "" || relDir == "." {
		return "./"
	}
	if !strings.HasPrefix(relDir, ".") {
		return "./" + relDir
	}
	return relDir + "/"
}
