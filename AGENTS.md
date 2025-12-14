# Agent Guidelines Manager CLI - Development Guide

## Purpose
Sync AGENTS.md files (single source of truth) across project hierarchy by creating symlinks to agent-specific guideline files.

## Core Commands
- `list` - Discover and display all guideline files (AGENTS.md, CLAUDE.md, .cursor/rules/*, etc.) with metadata
- `sync [flags]` - Find all AGENTS.md files recursively from current directory and create symlinks

## Flags
- `--claude` - Create CLAUDE.md symlinks pointing to AGENTS.md
- `--cursor` - Create .cursor/rules/agents.md symlinks (creates .cursor/rules/ dir if needed)
- `--dry-run` - Show what would be created without making changes
- `--verbose` - Show detailed output of all operations

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
1. Modify code files (main.go, discovery.go, symlink.go, output.go, agents.go)
2. Run `go build` to compile
3. Test with `./agents <command>`
4. Add new agent types in agents.go under SupportedAgents map

## Agent Configuration Reference

For detailed information about how different AI agents use guideline files, see [agents-conventions.md](agents-conventions.md). This comprehensive guide covers configuration options, hierarchical placement, and best practices for Claude, Cursor, Copilot, Gemini, Qwen, and other agents.
