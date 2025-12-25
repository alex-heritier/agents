# Agent Guidelines Manager CLI - Development Guide

## Purpose
Sync AGENTS.md files (single source of truth) across project hierarchy by creating symlinks to agent-specific guideline files.

## Core Commands
The CLI follows a module-based command structure: `agents <module> <command> [flags]`

### Modules

**rule** - Manage guideline files (AGENTS.md, CLAUDE.md, .cursor/rules/*, etc.)
- `agents rule list` - Discover and display all guideline files with metadata
- `agents rule sync [flags]` - Find all guideline source files and create symlinks
- `agents rule rm [flags]` - Delete guideline files for specified agents

**command** - Manage command files for agents
- `agents command list` - Discover and display all command files with metadata
- `agents command sync [flags]` - Find all command source files and create symlinks
- `agents command rm [flags]` - Delete command files for specified agents

**skill** - Manage Claude Code skills
- `agents skill list` - Discover and display all Claude Code skills
- `agents skill sync [flags]` - Sync skills from source directory to .claude/skills

## Flags
- `--claude` - Create CLAUDE.md symlinks pointing to AGENTS.md
- `--cursor` - Create .cursor/rules/agents.md symlinks (creates .cursor/rules/ dir if needed)
- `--dry-run` - Show what would be created without making changes
- `--verbose` - Show detailed output of all operations
- `--global`, `-g` - Show only user/system-wide agent guideline files (for `rule list`)
- `--<agent>` - Filter by specific agent files (e.g., --claude, --cursor)

## Discovery Rules
1. Search recursively from current directory downward
2. Find all AGENTS.md files in every directory
3. Skip directories in ignore list: `node_modules`, `.git`, `dist`, `build`, `.cursor`
4. For each AGENTS.md found, create specified symlinks in the same directory

## Symlink Behavior
- If target symlink exists, skip or update (ask user on conflict)
- Symlinks point to AGENTS.md in same directory (relative path)
- Create parent directories as needed (e.g., `.cursor/rules/`)
- All paths are relative to respect moving the project

## Output Format
**List command:**
- Table with columns: Directory | Agent | File | Type (file/symlink)
- Show all discovered guideline files recursively

**Sync command:**
- Show summary: X AGENTS.md files found, Y symlinks created, Z skipped
- With `--verbose`: list each operation (created/skipped and path)

## Development Guidelines
- Minimal dependencies (prefer stdlib where possible)
- Graceful degradation (CLI works even if some files are missing/malformed)
- Clear error messages (always show which file caused an issue)
- No auto-modifications without explicit `--force` flag

## Quick Start (Development)
1. Modify code files (main.go, discovery.go, symlink.go, output.go, config.go, args.go, paths.go, types.go)
2. Run `go build -o agents` to compile
3. Test with `./agents <module> <command>`
4. Add new agent types in tools.json configuration file

## Examples
```bash
# List guideline files
agents rule list
agents rule list --verbose
agents rule list --claude
agents rule list --gemini --global
agents rule list --claude --cursor --verbose

# Sync guideline files
agents rule sync --claude --cursor
agents rule sync --claude --cursor --dry-run

# Delete guideline files
agents rule rm --claude
agents rule rm --cursor --gemini --dry-run

# List command files
agents command list
agents command list --claude

# Sync command files
agents command sync --claude --cursor
agents command rm --claude

# Manage skills
agents skill list
agents skill list --verbose
agents skill sync
agents skill sync --dry-run --verbose
```

## Agent Configuration Reference

For detailed information about how different AI agents use guideline files, see [agents-conventions.md](agents-conventions.md). This comprehensive guide covers configuration options, hierarchical placement, and best practices for Claude, Cursor, Copilot, Gemini, Qwen, and other agents.
