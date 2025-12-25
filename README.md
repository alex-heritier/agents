# Agent Guidelines Manager CLI

A command-line tool to sync AI agent guideline files across project hierarchies. Maintain a single source of truth (AGENTS.md) and automatically create symlinks for different AI agents (Claude, Cursor, Copilot, Gemini, Qwen, and more).

## Problem

Different AI coding assistants look for guidelines in different ways:
- Some look for `CLAUDE.md` in the project root
- Cursor looks for `.cursor/rules/*.md` files
- Others have their own conventions

Managing multiple guideline files manually across nested directories is error-prone and duplicates effort.

## Solution

This tool treats `AGENTS.md` as the single source of truth and automatically creates symlinks for each agent type. It also manages custom/slash command files using `COMMANDS.md` as a source file. One command syncs your entire project hierarchy.

## Installation

### Build from source

```bash
git clone https://github.com/alex-heritier/agents.git
cd agents
go build -o agents
```

This creates an `agents` binary.

### Add to PATH

```bash
# Option 1: Install to /usr/local/bin
sudo cp agents /usr/local/bin/agents

# Option 2: Create a symlink
sudo ln -s "$PWD/agents" /usr/local/bin/agents
```

Then use `agents` from anywhere.

## Usage

### List guideline files

Discover all guideline files in your project:

```bash
agents list
agents list --verbose
```

### List command files

Discover all command files in your project:

```bash
agents list-commands
agents list-commands --verbose
```

### List skills

Discover all Claude Code skills (both global and project-level):

```bash
agents list-skills
agents list-skills --verbose
```

### Sync guideline files

Create symlinks for specified agents from all AGENTS.md files in your project:

```bash
# Create CLAUDE.md and .cursor/rules/agents.md symlinks
agents sync --claude --cursor

# Preview changes without applying
agents sync --claude --cursor --dry-run

# Verbose output showing each operation
agents sync --claude --cursor --verbose
```

### Sync command files

Create symlinks for specified agents from all COMMANDS.md files in your project:

```bash
# Create .claude/commands/commands.md and .cursor/commands/commands.md symlinks
agents sync-commands --claude --cursor

# Preview changes without applying
agents sync-commands --claude --cursor --dry-run

# Verbose output showing each operation
agents sync-commands --claude --cursor --verbose
```

### Sync skills

Sync skills from source `skills/` directories to `.claude/skills/`:

```bash
# Sync all skills from skills/ to .claude/skills/
agents sync-skills

# Preview changes without applying
agents sync-skills --dry-run

# Verbose output showing each operation
agents sync-skills --verbose
```

### Remove guideline files

Delete guideline files for specific agents:

```bash
# Delete all CLAUDE.md files
agents rm --claude

# Delete multiple agents
agents rm --cursor --gemini --qwen

# Preview deletions
agents rm --claude --dry-run --verbose
```

### Remove command files

Delete command files for specific agents:

```bash
# Delete all Claude command files
agents rm-commands --claude

# Delete multiple agents
agents rm-commands --claude --cursor

# Preview deletions
agents rm-commands --claude --dry-run --verbose
```

## Supported Agents

- **claude** - Creates `CLAUDE.md`
- **cursor** - Creates `.cursor/rules/agents.md`
- **copilot** - Creates `COPILOT.md`
- **gemini** - Creates `GEMINI.md`
- **qwen** - Creates `QWEN.md`

Adding new agent types is simple: edit `providers.json` and add to the `providers` map.

## Workflow

### Guidelines and Commands

1. Create an `AGENTS.md` file in your project root with guidelines for your AI coding assistants
2. Create the same file in any nested directories (subdirectories with their own guidelines)
3. Run `agents sync --claude --cursor` (or any agents you use)
4. All agent-specific files are now symlinks pointing to the corresponding AGENTS.md

To sync custom/slash commands, add `COMMANDS.md` files and run `agents sync-commands`.

When you update AGENTS.md, all agent-specific files automatically reflect the changes since they're symlinks.

### Skills

1. Create a `skills/` directory in your project root
2. Add skill subdirectories with `SKILL.md` files (see [Claude Code Skills](https://code.claude.com/docs/en/skills) for format)
3. Run `agents sync-skills` to copy skills to `.claude/skills/`
4. Use `agents list-skills` to view all available skills (global and project-level)

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

## Skill File Format

Skills are reusable modules that extend Claude's capabilities. Each skill requires:

- A directory named after the skill (e.g., `skills/my-skill/`)
- A `SKILL.md` file with YAML frontmatter:

```yaml
---
name: my-skill-name
description: Clear description of what this skill does and when to use it
license: MIT
allowed-tools: Read, Grep, Glob
---

# Skill Instructions

[Detailed instructions for Claude on how to use this skill]

## Examples
- Example usage 1
- Example usage 2
```

**Required fields:**
- `name` - Unique identifier (lowercase, hyphenated)
- `description` - Maximum 1024 characters; describes purpose and use cases

**Optional fields:**
- `license` - License information
- `allowed-tools` - Comma-separated list of tools Claude can use

For more information, see [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills).

## Help

```bash
agents help
agents list --help
agents sync --help
agents rm --help
agents list-skills --help
agents sync-skills --help
```

## Tech Stack

- **Language:** Go
- **Dependencies:** None (standard library only)
- **Platforms:** macOS, Linux, Windows

## Contributing

Contributions welcome! To add a new agent type:

1. Edit `providers.json` and add to the `providers` map
2. Run `go build` (and tests if added)
3. Submit a PR

## Agent Configuration Reference

For comprehensive information about how different AI agents use guideline files, see [agents-conventions.md](agents-conventions.md). This document covers:

- Configuration options for Claude, Cursor, Copilot, Gemini, Qwen, and other agents
- Hierarchical placement and precedence rules
- Auto-generation and external file referencing
- Best practices for cross-agent compatibility

## Provider Configuration

Provider definitions live in `providers.json` (file names, directories, and source file names). You can extend or override these definitions by creating an optional file at:

```
$XDG_CONFIG_HOME/agents/providers.json
```

If `XDG_CONFIG_HOME` is not set, the tool falls back to `~/.config/agents/providers.json`.

## TODO

- Support global/system-wide commands files

## License

MIT
