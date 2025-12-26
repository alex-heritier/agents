package main

// FileSpec defines a file specification with directory, filename, and global paths
type FileSpec struct {
	Pattern string   `json:"pattern"`
	Global  []string `json:"global,omitempty"`
}

type ToolConfig struct {
	Name    string   `json:"name,omitempty"`
	Pattern string   `json:"pattern"`
	Global  []string `json:"global,omitempty"`
}

// ToSpec returns a FileSpec from the flattened ToolConfig
func (t ToolConfig) ToSpec() *FileSpec {
	return &FileSpec{
		Pattern: t.Pattern,
		Global:  t.Global,
	}
}

type ToolsConfig struct {
	Standard string                `json:"standard,omitempty"`
	Tools    map[string]ToolConfig `json:"tools"`
}

// ManagedFile represents a discovered file
type ManagedFile struct {
	Path      string
	Dir       string
	Tools     []string
	File      string
	IsSymlink bool
	Size      int64
}
