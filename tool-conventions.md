### Tool Conventions

Local = project or repository scope. Global = user/system scope. If a tool has no documented file-based location, it is marked as not documented.

#### Rule Locations

| Tool | Local | Global | Notes |
| --- | --- | --- | --- |
| Claude Code | `./CLAUDE.md` or `./.claude/CLAUDE.md`; `.claude/rules/*.md` | `~/.claude/CLAUDE.md`; `~/.claude/rules/*.md` | Also supports enterprise policy locations per OS. |
| Codex | `AGENTS.md` or `AGENTS.override.md` in repo directories | `$CODEX_HOME/AGENTS.md` or `AGENTS.override.md` (default `~/.codex`) | |
| Gemini CLI | `GEMINI.md` (or configured `context.fileName`) in CWD/parents/subdirs | `~/.gemini/<context-file>` (default `~/.gemini/GEMINI.md`) | Context discovery respects `.git` root and ignore rules. |
| Qwen-Code CLI | Not documented | Not documented | README/CONTRIBUTING do not specify rule/command/skill files. |
| Amp Code | `AGENTS.md` in CWD/parents/subtrees (fallback: `AGENT.md`, `CLAUDE.md`) | `~/.config/amp/AGENTS.md` and `~/.config/AGENTS.md` | Uses `$XDG_CONFIG_HOME` when set. |
| Cursor | `.cursor/rules/<rule>/RULE.md`; `AGENTS.md` in root/subdirs; legacy `.cursorrules` | User/team rules via Cursor Settings or Dashboard (not file-based) | Rules created via "New Cursor Rule" or Settings > Rules. |
| OpenCode | `AGENTS.md` in project root/subdirs | `~/.config/opencode/AGENTS.md` | Additional instruction files can be referenced in `opencode.json`. |

#### Command Locations

| Tool | Local | Global | Notes |
| --- | --- | --- | --- |
| Claude Code | `.claude/commands/*.md` | `~/.claude/commands/*.md` | |
| Codex | Not documented | Not documented | |
| Gemini CLI | `<project>/.gemini/commands/*.toml` | `~/.gemini/commands/*.toml` | |
| Qwen-Code CLI | Not documented | `~/.qwen/commands/*.toml` | Global commands defined in TOML. |
| Amp Code | `.agents/commands/*` | `~/.config/amp/commands/*` | |
| Cursor | `.cursor/commands/*` | `~/.cursor/commands/*` | Project and global command storage. |
| OpenCode | `.opencode/command/*.md` | `~/.config/opencode/command/*.md` | |

#### Skill Locations

| Tool | Local | Global | Notes |
| --- | --- | --- | --- |
| Claude Code | `.claude/skills/<name>/SKILL.md` | `~/.claude/skills/<name>/SKILL.md` | |
| Codex | `.codex/skills` in repo (CWD/parent/root) | `~/.codex/skills` (default); `$CODEX_HOME/skills` | Repo skills can be placed in CWD, parent, or repo root `.codex/skills`. |
| Gemini CLI | `.gemini/commands/*.toml` | `~/.gemini/commands/*.toml` | Implemented as commands. |
| Qwen-Code CLI | Not documented | Not documented | |
| Amp Code | `.agents/skills/` | `~/.config/amp/skills/` | Also loads `.claude/skills`. |
| Cursor | Not documented | Not documented | Uses rules/commands instead. |
| OpenCode | `.opencode/skill/<name>/SKILL.md` (also loads `.claude/skills/*/SKILL.md`) | `~/.opencode/skill/<name>/SKILL.md` | |

Sources:
- Claude Code memory and commands: https://code.claude.com/docs/en/memory , https://code.claude.com/docs/en/slash-commands , https://code.claude.com/docs/en/skills
- Codex AGENTS.md and skills: https://developers.openai.com/codex/guides/agents-md/ , https://developers.openai.com/codex/skills
- Gemini CLI context and commands: https://geminicli.com/docs/get-started/configuration , https://geminicli.com/docs/cli/custom-commands
- Qwen-Code CLI repo docs: https://github.com/QwenLM/qwen-code
- Amp Code manual: https://ampcode.com/manual
- Cursor rules: https://cursor.com/docs/context/rules
- OpenCode rules/commands/skills: https://opencode.ai/docs/rules/ , https://opencode.ai/docs/commands/ , https://opencode.ai/docs/skills/