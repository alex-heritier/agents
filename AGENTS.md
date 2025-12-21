# Agent Guidelines Manager CLI - Development Guide

## Purpose
Sync AGENTS.md files (single source of truth) and COMMANDS directories across project hierarchy by creating symlinks to agent-specific guideline files and slash commands.

## Core Commands

### Guideline Files
- `list` - Discover and display all guideline files (AGENTS.md, CLAUDE.md, .cursor/rules/*, etc.) with metadata
- `sync [flags]` - Find all AGENTS.md files recursively from current directory and create symlinks
- `rm [flags]` - Delete guideline files for specified agents

### Slash Commands
- `list-commands` (alias: `list-cmds`) - Discover and display all slash command files
- `sync-commands` (alias: `sync-cmds`) - Find all COMMANDS directories and create command symlinks
- `rm-commands` (alias: `rm-cmds`) - Delete command files for specified agents

## Flags
- `--claude` - Create CLAUDE.md symlinks pointing to AGENTS.md / Create .claude/commands symlinks
- `--cursor` - Create .cursor/rules/agents.md symlinks / Create .cursor/commands symlinks
- `--copilot`, `--gemini`, `--qwen`, `--warp`, `--amp`, `--opencode` - Support for other agents
- `--dry-run` - Show what would be created without making changes
- `--verbose` - Show detailed output of all operations
- `--global` or `-g` - Show only user/system-wide files

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

## Configuration System

### Provider Configuration File
All agent configurations are defined in `providers.yaml`:
- Guideline file names and locations
- Slash command directory structures
- Global/user-level file paths

### XDG Configuration Support
Users can extend or override the default configuration by creating:
- `~/.config/agents/providers.yaml` (or `$XDG_CONFIG_HOME/agents/providers.yaml`)

This allows customization of:
- Custom agent providers
- Override default file paths
- Add new slash command conventions

## Development Guidelines
- Minimal dependencies (zero external dependencies, uses built-in YAML parser)
- Graceful degradation (CLI works even if some files are missing/malformed)
- Clear error messages (always show which file caused an issue)
- No auto-modifications without explicit `--force` flag
- Provider configurations are loaded from embedded YAML and merged with user overrides

## Quick Start (Development)
1. Modify code files:
   - `src/main.ts` - Command handling and CLI
   - `src/config.ts` - Provider configuration loading
   - `src/agents.ts` - Agent/provider management
   - `src/discovery.ts` - File discovery logic
   - `src/commands.ts` - Slash commands handling
   - `src/symlink.ts` - Symlink creation logic
   - `src/output.ts` - Output formatting
   - `providers.yaml` - Agent provider definitions
2. Run `bun run build` to compile
3. Test with `bun run dist/main.js <command>` or `bun run src/main.ts <command>` for dev
4. Add new agent types in providers.yaml (not hardcoded in TypeScript files)
5. User overrides go in ~/.config/agents/providers.yaml

## Agent Configuration Reference

For detailed information about how different AI agents use guideline files and slash commands, see [agents-conventions.md](agents-conventions.md). This comprehensive guide covers configuration options, hierarchical placement, and best practices for Claude, Cursor, Copilot, Gemini, Qwen, and other agents.
