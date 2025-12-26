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
