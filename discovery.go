package main

import (
	"os"
	"regexp"
	"strings"
)

var ignoreDir = map[string]bool{
	"node_modules": true,
	".git":         true,
	"dist":         true,
	"build":        true,
	".cursor":      true,
}

// discoverSources finds all source files with the given name
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
		if !info.IsDir() && info.Name() == sourceName {
			sources = append(sources, path)
		}
		return "continue"
	})

	return sources
}

// discoverAll discovers all managed files
func discoverAll(cfg *ProvidersConfig, sourceName string, specSelector func(ProviderConfig) *FileSpec) []ManagedFile {
	cwd, err := os.Getwd()
	if err != nil {
		return []ManagedFile{}
	}

	allowedDirs := allowedProviderDirs(cfg, specSelector)
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
		if filename == sourceName {
			agent = strings.ToUpper(strings.TrimSuffix(sourceName, ".md"))
		} else {
			for agentName, provider := range cfg.Providers {
				spec := specSelector(provider)
				if spec == nil {
					continue
				}
				if filename == spec.File {
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
			Agent:     agent,
			File:      filename,
			IsSymlink: isSymlink,
			Size:      info.Size(),
		})

		return "continue"
	})

	return files
}

// discoverGlobalOnly discovers global guideline files
func discoverGlobalOnly(cfg *ProvidersConfig) []ManagedFile {
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
		agent := inferProviderFromFilename(cfg, filename)
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
			Agent:     agent,
			File:      filename,
			IsSymlink: isSymlink,
			Size:      info.Size(),
		})
	}

	return files
}

// inferProviderFromFilename infers the provider name from a filename
func inferProviderFromFilename(cfg *ProvidersConfig, filename string) string {
	if filename == cfg.Sources.Guidelines {
		return strings.ToUpper(strings.TrimSuffix(filename, ".md"))
	}

	for agentName, provider := range cfg.Providers {
		if provider.Guidelines != nil && filename == provider.Guidelines.File {
			return strings.ToUpper(agentName)
		}
	}

	return ""
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

// allowedProviderDirs returns a set of allowed provider directories
func allowedProviderDirs(cfg *ProvidersConfig, specSelector func(ProviderConfig) *FileSpec) map[string]bool {
	allowed := make(map[string]bool)

	for _, provider := range cfg.Providers {
		spec := specSelector(provider)
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

// globalGuidelinePaths returns expanded global guideline paths
func globalGuidelinePaths(cfg *ProvidersConfig) []string {
	paths := make([]string, len(cfg.GlobalGuidelines))
	for i, path := range cfg.GlobalGuidelines {
		paths[i] = expandHomePath(path)
	}
	return paths
}

// discoverSkills discovers all skills in global and project directories
func discoverSkills() []SkillFile {
	skills := []SkillFile{}

	// Global skills
	globalSkillsDir := pathJoin(getHomeDir(), ".claude", "skills")
	if dirExists(globalSkillsDir) {
		globalSkills := discoverSkillsInDir(globalSkillsDir, "global")
		skills = append(skills, globalSkills...)
	}

	// Project skills
	cwd, err := os.Getwd()
	if err == nil {
		projectSkillsDir := pathJoin(cwd, ".claude", "skills")
		if dirExists(projectSkillsDir) {
			projectSkills := discoverSkillsInDir(projectSkillsDir, "project")
			skills = append(skills, projectSkills...)
		}
	}

	return skills
}

// discoverSourceSkills discovers source skill directories
func discoverSourceSkills(sourceDirName string) []string {
	cwd, err := os.Getwd()
	if err != nil {
		return []string{}
	}

	skillDirs := []string{}

	walk(cwd, func(path string, info os.FileInfo) string {
		if info.IsDir() && ignoreDir[info.Name()] {
			return "skip"
		}

		// Skip .claude directory
		if info.IsDir() && info.Name() == ".claude" {
			return "skip"
		}

		if info.IsDir() && info.Name() == sourceDirName {
			// Found a source skills directory
			entries, err := os.ReadDir(path)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						skillFile := pathJoin(path, entry.Name(), "SKILL.md")
						if fileExists(skillFile) {
							skillDirs = append(skillDirs, pathJoin(path, entry.Name()))
						}
					}
				}
			}
			return "skip"
		}

		return "continue"
	})

	return skillDirs
}

// discoverSkillsInDir discovers skills in a specific directory
func discoverSkillsInDir(dir, location string) []SkillFile {
	skills := []SkillFile{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return skills
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillPath := pathJoin(dir, skillName)
		skillFile := pathJoin(skillPath, "SKILL.md")

		if !fileExists(skillFile) {
			continue
		}

		metadata, parseErr := parseSkillMetadata(skillFile)

		skill := SkillFile{
			Path:      skillFile,
			Dir:       skillPath,
			SkillName: skillName,
			Location:  location,
		}

		if parseErr != nil {
			skill.Error = parseErr.Error()
		} else {
			skill.Metadata = metadata
		}

		skills = append(skills, skill)
	}

	return skills
}

// parseSkillMetadata parses YAML frontmatter from a skill file
func parseSkillMetadata(skillFile string) (*SkillMetadata, error) {
	content, err := os.ReadFile(skillFile)
	if err != nil {
		return nil, err
	}

	// Match frontmatter
	frontmatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---`)
	matches := frontmatterRegex.FindSubmatch(content)
	if matches == nil {
		return nil, &SkillError{"No frontmatter found"}
	}

	frontmatter := string(matches[1])
	metadata := &SkillMetadata{}

	// Parse simple YAML key-value pairs
	lines := strings.Split(frontmatter, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			metadata.Name = value
		case "description":
			metadata.Description = value
		case "license":
			metadata.License = value
		case "allowed-tools":
			metadata.AllowedTools = value
		}
	}

	if metadata.Name == "" || metadata.Description == "" {
		return nil, &SkillError{"Missing required fields (name, description)"}
	}

	return metadata, nil
}

// dirExists checks if a directory exists
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// SkillError is a custom error type for skill parsing
type SkillError struct {
	Message string
}

func (e *SkillError) Error() string {
	return e.Message
}
