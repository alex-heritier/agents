package main

// FileSpec defines a file specification with directory, filename, and global paths
type FileSpec struct {
	File   string   `json:"file"`
	Dir    string   `json:"dir"`
	Global []string `json:"global,omitempty"`
}

type ToolConfig struct {
	Name       string    `json:"name,omitempty"`
	Guidelines *FileSpec `json:"guidelines,omitempty"`
	Commands   *FileSpec `json:"commands,omitempty"`
	Skills     *FileSpec `json:"skills,omitempty"`
}

type ToolsConfig struct {
	Standard string                `json:"standard,omitempty"`
	Tools    map[string]ToolConfig `json:"tools"`
}

// ManagedFile represents a discovered file
type ManagedFile struct {
	Path      string
	Dir       string
	Tool      string
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
