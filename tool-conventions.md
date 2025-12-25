### Tool Conventions: Rule, Command, and Skill Locations

Local = project or repository scope. Global = user/system scope. If a tool has no documented file-based location, it is marked as not documented.

| Tool | Rules (local) | Rules (global) | Commands (local) | Commands (global) | Skills (local) | Skills (global) | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Claude Code | `./CLAUDE.md` or `./.claude/CLAUDE.md`; `.claude/rules/*.md` | `~/.claude/CLAUDE.md`; `~/.claude/rules/*.md` | `.claude/commands/*.md` | `~/.claude/commands/*.md` | `.claude/skills/<name>/SKILL.md` | `~/.claude/skills/<name>/SKILL.md` | Also supports enterprise policy locations per OS. |
| Codex | `AGENTS.md` or `AGENTS.override.md` in repo directories | `$CODEX_HOME/AGENTS.md` or `AGENTS.override.md` (default `~/.codex`) | Not documented | Not documented | `.codex/skills` in repo (CWD/parent/root) | `$CODEX_HOME/skills` (default `~/.codex/skills`); `/etc/codex/skills` | Repo skills can be placed in CWD, parent, or repo root `.codex/skills`. |
| Gemini CLI | `GEMINI.md` (or configured `context.fileName`) in CWD/parents/subdirs | `~/.gemini/<context-file>` (default `~/.gemini/GEMINI.md`) | `<project>/.gemini/commands/*.toml` | `~/.gemini/commands/*.toml` | Not documented | Not documented | Context discovery respects `.git` root and ignore rules. |
| Qwen-Code CLI | Not documented | Not documented | Not documented | Not documented | Not documented | Not documented | README/CONTRIBUTING do not specify rule/command/skill files. |
| Amp Code | `AGENTS.md` in CWD/parents/subtrees (fallback: `AGENT.md`, `CLAUDE.md`) | `~/.config/amp/AGENTS.md` and `~/.config/AGENTS.md` | `.agents/commands/*` | `~/.config/amp/commands/*` | Not documented | Not documented | Uses `$XDG_CONFIG_HOME` when set. |
| Cursor | `.cursor/rules/<rule>/RULE.md`; `AGENTS.md` in root/subdirs; legacy `.cursorrules` | User/team rules via Cursor Settings or Dashboard (not file-based) | Not documented | Not documented | Not documented | Not documented | Rules created via "New Cursor Rule" or Settings > Rules. |
| OpenCode | `AGENTS.md` in project root/subdirs | `~/.config/opencode/AGENTS.md` | `.opencode/command/*.md` | `~/.config/opencode/command/*.md` | `.opencode/skill/<name>/SKILL.md` (also loads `.claude/skills/*/SKILL.md`) | `~/.opencode/skill/<name>/SKILL.md` | Additional instruction files can be referenced in `opencode.json`. |

Sources:
- Claude Code memory and commands: https://code.claude.com/docs/en/memory , https://code.claude.com/docs/en/slash-commands , https://code.claude.com/docs/en/skills
- Codex AGENTS.md and skills: https://developers.openai.com/codex/guides/agents-md/ , https://developers.openai.com/codex/skills
- Gemini CLI context and commands: https://geminicli.com/docs/get-started/configuration , https://geminicli.com/docs/cli/custom-commands
- Qwen-Code CLI repo docs: https://github.com/QwenLM/qwen-code
- Amp Code manual: https://ampcode.com/manual
- Cursor rules: https://cursor.com/docs/context/rules
- OpenCode rules/commands/skills: https://opencode.ai/docs/rules/ , https://opencode.ai/docs/commands/ , https://opencode.ai/docs/skills/
