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
- Minimal dependencies (uses only Node.js built-in modules)
- Graceful degradation (CLI works even if some files are missing/malformed)
- Clear error messages (always show which file caused an issue)
- No auto-modifications without explicit `--force` flag

## Quick Start (Development)
1. Modify code files in `src/` (index.ts, discovery.ts, symlink.ts, output.ts, agents.ts)
2. Run `bun run src/index.ts <command>` to test
3. Run `bun run build` to compile standalone binary
4. Add new agent types in `src/agents.ts` under SupportedAgents map

## Project Structure
```
src/
├── index.ts      # CLI entry point and command router
├── agents.ts     # Agent configurations (claude, cursor, etc.)
├── discovery.ts  # File discovery logic
├── output.ts     # Output formatting
├── symlink.ts    # Symlink management
└── types.ts      # TypeScript type definitions
```

## Agent Configuration Reference

For detailed information about how different AI agents use guideline files, see [agents-conventions.md](agents-conventions.md). This comprehensive guide covers configuration options, hierarchical placement, and best practices for Claude, Cursor, Copilot, Gemini, Qwen, and other agents.
