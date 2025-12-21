# Agent Guidelines Manager CLI

A command-line tool to sync AI agent guideline files across project hierarchies. Maintain a single source of truth (AGENTS.md) and automatically create symlinks for different AI agents (Claude, Cursor, Copilot, Gemini, Qwen, and more).

## Problem

Different AI coding assistants look for guidelines in different ways:
- Some look for `CLAUDE.md` in the project root
- Cursor looks for `.cursor/rules/*.md` files
- Others have their own conventions

Managing multiple guideline files manually across nested directories is error-prone and duplicates effort.

## Solution

This tool treats `AGENTS.md` as the single source of truth and automatically creates symlinks for each agent type. One command syncs your entire project hierarchy.

## Installation

### Build from source

```bash
git clone https://github.com/alex-heritier/agents.git
cd agents
go build
```

This creates an `agents` binary in the current directory.

### Add to PATH

```bash
mv agents /usr/local/bin/
```

Then use `agents` from anywhere.

## Usage

### Guideline Files

#### List guideline files

Discover all guideline files in your project:

```bash
agents list
agents list --verbose
agents list --global  # Show only user/system-wide files
agents list --claude  # Filter by specific agent
```

#### Sync guideline files

Create symlinks for specified agents from all AGENTS.md files in your project:

```bash
# Create CLAUDE.md and .cursor/rules/agents.md symlinks
agents sync --claude --cursor

# Preview changes without applying
agents sync --claude --cursor --dry-run

# Verbose output showing each operation
agents sync --claude --cursor --verbose
```

#### Remove guideline files

Delete guideline files for specific agents:

```bash
# Delete all CLAUDE.md files
agents rm --claude

# Delete multiple agents
agents rm --cursor --gemini --qwen

# Preview deletions
agents rm --claude --dry-run --verbose
```

### Slash Commands

#### List command files

Discover all slash command files in your project:

```bash
agents list-commands
agents list-cmds  # Alias
agents list-commands --verbose
agents list-commands --global  # Show only user/system-wide commands
agents list-commands --claude  # Filter by specific agent
```

#### Sync command files

Create command symlinks for specified agents from all COMMANDS directories:

```bash
# Create symlinks in .claude/commands/ from COMMANDS/
agents sync-commands --claude

# Sync for multiple agents
agents sync-commands --claude --cursor

# Preview changes without applying
agents sync-commands --claude --dry-run --verbose
```

#### Remove command files

Delete command files for specific agents:

```bash
# Delete all Claude command files
agents rm-commands --claude

# Delete multiple agents
agents rm-commands --cursor --gemini --dry-run
```

## Supported Agents

- **claude** - Creates `CLAUDE.md`
- **cursor** - Creates `.cursor/rules/agents.md`
- **copilot** - Creates `COPILOT.md`
- **gemini** - Creates `GEMINI.md`
- **qwen** - Creates `QWEN.md`

Adding new agent types is simple: edit `agents.go` and add to the `SupportedAgents` map.

## Workflow

1. Create an `AGENTS.md` file in your project root with guidelines for your AI coding assistants
2. Create the same file in any nested directories (subdirectories with their own guidelines)
3. Run `agents sync --claude --cursor` (or any agents you use)
4. All agent-specific files are now symlinks pointing to the corresponding AGENTS.md

When you update AGENTS.md, all agent-specific files automatically reflect the changes since they're symlinks.

## Safety Features

- **Conflict detection**: If a file already exists and differs from AGENTS.md, the tool prompts you before overwriting
- **Dry-run mode**: Preview all changes with `--dry-run` before applying
- **Verbose logging**: Use `--verbose` to see detailed operations

## Examples

### Project structure before

```
myproject/
├── AGENTS.md
├── src/
│   └── AGENTS.md
└── docs/
    └── AGENTS.md
```

### Create guidelines for Claude and Cursor

```bash
agents sync --claude --cursor
```

### Project structure after

```
myproject/
├── AGENTS.md
├── CLAUDE.md -> AGENTS.md (symlink)
├── .cursor/rules/
│   └── agents.md -> ../../AGENTS.md (symlink)
├── src/
│   ├── AGENTS.md
│   ├── CLAUDE.md -> AGENTS.md (symlink)
│   └── .cursor/rules/
│       └── agents.md -> ../../../AGENTS.md (symlink)
└── docs/
    ├── AGENTS.md
    ├── CLAUDE.md -> AGENTS.md (symlink)
    └── .cursor/rules/
        └── agents.md -> ../../../AGENTS.md (symlink)
```

## Help

```bash
agents help
agents list --help
agents sync --help
agents rm --help
```

## Tech Stack

- **Language:** Go
- **Dependencies:** None (standard library only)
- **Binary size:** ~5MB single executable
- **Platforms:** macOS, Linux, Windows

## Contributing

Contributions welcome! To add a new agent type:

1. Edit `agents.go` and add to `SupportedAgents` map
2. Run `go build && go test` (if tests exist)
3. Submit a PR

## Agent Configuration Reference

For comprehensive information about how different AI agents use guideline files, see [agents-conventions.md](agents-conventions.md). This document covers:

- Configuration options for Claude, Cursor, Copilot, Gemini, Qwen, and other agents
- Hierarchical placement and precedence rules
- Auto-generation and external file referencing
- Best practices for cross-agent compatibility

## Configuration

### Provider Configuration

All agent provider configurations are defined in `providers.yaml`. This file specifies:
- Guideline file names and locations for each agent
- Slash command directory structure
- Global/user-level file paths

Users can extend or override the default configuration by creating `~/.config/agents/providers.yaml` (or `$XDG_CONFIG_HOME/agents/providers.yaml`).

### Slash Commands

The tool now supports syncing slash commands across different AI agents. Create a `COMMANDS` directory in your project with markdown files for each command, then use `agents sync-commands` to create symlinks in agent-specific command directories.

Example:
```bash
# Create a COMMANDS directory with your slash commands
mkdir COMMANDS
echo "# Review Code\nReview changes and provide feedback." > COMMANDS/review.md

# Sync to Claude's command directory
agents sync-commands --claude

# List all command files
agents list-commands
```

## TODO

- Support global/system-wide agents files (✅ DONE)
- Support custom slash commands listing and syncing (✅ DONE)

## License

MIT
