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

// discoverSources finds all source files recursively from current directory
func discoverSources(sourceName string) []string {
	var sources []string
	cwd, _ := os.Getwd()

	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && ignoreDir[info.Name()] {
			return filepath.SkipDir
		}

		if !info.IsDir() && info.Name() == sourceName {
			sources = append(sources, path)
		}

		return nil
	})

	return sources
}

// discoverAll finds all managed files (source plus provider-specific files)
func discoverAll(cfg ProvidersConfig, sourceName string, specSelector func(ProviderConfig) *FileSpec) []ManagedFile {
	var files []ManagedFile
	cwd, _ := os.Getwd()
	allowedDirs := allowedProviderDirs(cfg, specSelector)

	filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if ignoreDir[info.Name()] && !allowedDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		dir := filepath.Dir(path)
		filename := info.Name()

		var agent string
		if filename == sourceName {
			agent = strings.TrimSuffix(strings.ToUpper(sourceName), ".MD")
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
				return nil
			}
		}

		isSymlink := isSymlink(path)

		files = append(files, ManagedFile{
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

func allowedProviderDirs(cfg ProvidersConfig, specSelector func(ProviderConfig) *FileSpec) map[string]bool {
	allowed := map[string]bool{}
	for _, provider := range cfg.Providers {
		spec := specSelector(provider)
		if spec == nil || spec.Dir == "" {
			continue
		}

		parts := strings.Split(spec.Dir, string(filepath.Separator))
		if len(parts) > 0 {
			allowed[parts[0]] = true
		}
	}
	return allowed
}

// globalGuidelinePaths returns the standard locations for global agent guideline files
func globalGuidelinePaths(cfg ProvidersConfig) []string {
	var paths []string
	for _, raw := range cfg.GlobalGuidelines {
		paths = append(paths, expandHomePath(raw))
	}
	return paths
}

// discoverGlobalOnly finds only user/system-wide agent guideline files
func discoverGlobalOnly(cfg ProvidersConfig) []ManagedFile {
	var files []ManagedFile

	globalLocations := globalGuidelinePaths(cfg)
	if len(globalLocations) == 0 {
		return files
	}

	for _, location := range globalLocations {
		if fileExists(location) {
			info, err := os.Stat(location)
			if err != nil {
				continue
			}

			filename := filepath.Base(location)
			dir := filepath.Dir(location)

			agent := inferProviderFromFilename(cfg, filename)
			if agent == "" {
				continue
			}

			isSymlink := isSymlink(location)

			files = append(files, ManagedFile{
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

// inferProviderFromFilename determines the provider type from a filename
func inferProviderFromFilename(cfg ProvidersConfig, filename string) string {
	if filename == cfg.Sources.Guidelines {
		return strings.TrimSuffix(strings.ToUpper(filename), ".MD")
	}

	for agentName, provider := range cfg.Providers {
		spec := provider.Guidelines
		if spec == nil {
			continue
		}
		if filename == spec.File {
			return strings.ToUpper(agentName)
		}
	}

	return ""
}

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

func expandHomePath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	return filepath.Join(home, strings.TrimPrefix(path, "~/"))
}
