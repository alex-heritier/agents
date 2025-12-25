package main

// FileSpec defines a file specification with directory and filename
type FileSpec struct {
	File string `json:"file"`
	Dir  string `json:"dir"`
}

// ProviderConfig defines configuration for a specific provider
type ProviderConfig struct {
	Name       string    `json:"name,omitempty"`
	Guidelines *FileSpec `json:"guidelines,omitempty"`
	Commands   *FileSpec `json:"commands,omitempty"`
	Skills     *FileSpec `json:"skills,omitempty"`
}

// SourcesConfig defines source file names
type SourcesConfig struct {
	Guidelines string `json:"guidelines"`
	Commands   string `json:"commands"`
	Skills     string `json:"skills"`
}

// ProvidersConfig is the root configuration structure
type ProvidersConfig struct {
	Sources          SourcesConfig             `json:"sources"`
	GlobalGuidelines []string                  `json:"globalGuidelines"`
	Providers        map[string]ProviderConfig `json:"providers"`
}

// ManagedFile represents a discovered file
type ManagedFile struct {
	Path      string
	Dir       string
	Agent     string
	File      string
	IsSymlink bool
	Size      int64
}

// SkillMetadata contains metadata from a skill's frontmatter
type SkillMetadata struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	License      string                 `json:"license,omitempty"`
	AllowedTools string                 `json:"allowedTools,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SkillFile represents a discovered skill
type SkillFile struct {
	Path      string
	Dir       string
	SkillName string
	Location  string // "global" or "project"
	Metadata  *SkillMetadata
	Error     string
}
