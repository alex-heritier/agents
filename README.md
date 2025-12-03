# Agent Guidelines Manager CLI

## Purpose
CLI tool to discover, validate, merge, and manage AI agent configuration files (AGENTS.md, CLAUDE.md, .cursor/rules/*, etc.) across project hierarchies.

## Tech Stack
- **Language:** Go
- **Rationale:** Single binary, minimal runtime dependencies, fast execution, native file I/O and CLI support

## Why This Tool?
Managing multiple guideline files across projects and subdirectories can become complex. This tool centralizes discovery, validation, and merging to ensure consistency and catch conflicts early.
