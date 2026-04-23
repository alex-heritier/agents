# agents-cli

> **One set of AI coding guidelines, skills, commands, and subagents — synced
> to every harness you use.**

`agents-cli` is a tiny, focused CLI that lets you maintain **one** source of
truth for your AI coding setup (in, say, `opencode`) and automatically
propagate it — via symlinks or copies — to **every other harness** you
occasionally reach for: Claude Code, Gemini CLI, Droid (Factory), Kilo Code,
Codex, Cursor, Amp Code, Qwen Code, GitHub Copilot.

It works at **two scopes**:

- **Project** (repo root, detected from `.git`)
- **Global** (your `$HOME` — `~/.config/opencode`, `~/.claude`, `~/.gemini`, …)

It covers **four resource types**:

| Type         | What it is                                           |
|--------------|------------------------------------------------------|
| `guideline`  | `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` / `RULE.md`  |
| `skill`      | `<name>/SKILL.md` folders with a `name` + `description` |
| `command`    | Custom slash commands (`/commit`, `/review`)         |
| `subagent`   | Specialized agents / droids (`reviewer`, `debugger`) |

---

## Install

Requires **Python ≥ 3.10** and [`uv`](https://docs.astral.sh/uv/).

```bash
git clone https://github.com/alex-heritier/agents-cli
cd agents-cli
uv sync
```

Then either run it via uv, or install it as a shell command:

```bash
uv run agents --help            # via uv
uv tool install --editable .    # → `agents` on your $PATH
```

---

## Quickstart

```bash
cd my-project/

agents init                     # write .agents.toml + seed AGENTS.md
agents status                   # see what each harness has
agents sync --dry-run           # preview propagation
agents sync                     # do it
```

That's it. You now have one `AGENTS.md`, linked as `CLAUDE.md`, `GEMINI.md`,
and so on. Edit `AGENTS.md` and every harness sees the change instantly.

---

## Commands

### Top-level

| Command               | What it does                                         |
|-----------------------|------------------------------------------------------|
| `agents init`         | Create `.agents.toml` and (optionally) seed `AGENTS.md` |
| `agents doctor`       | Diagnose your setup, list discovered resources       |
| `agents status`       | Compact grid: what's synced where                    |
| `agents sync`         | Propagate **everything** from source → targets       |

### Per resource

Every resource type has the same verbs — `list`, `show`, `new`, `edit`, `rm`,
`sync`:

```bash
agents guideline list                   # all guideline files in the project
agents guideline sync --to claude       # sync just guidelines to claude

agents skill list                       # all skills across harnesses
agents skill new my-skill               # scaffold ~/.opencode/skills/my-skill/
agents skill show my-skill              # exact lookup
agents skill show                       # interactive picker (omit NAME)
agents skill edit my-skill              # opens $EDITOR on exact match
agents skill edit                       # interactive picker, then opens $EDITOR
agents skill sync --to claude,gemini
agents skill rm my-skill

agents rule show                        # browse all guideline files interactively
agents rule edit                        # pick a guideline and open it in $EDITOR

agents command new commit
agents subagent new reviewer
```

### Useful flags

| Flag                   | Meaning                                               |
|------------------------|-------------------------------------------------------|
| `--project` / `-p`     | Operate on project scope only                         |
| `--global` / `-g`      | Operate on global scope only                          |
| `--from <harness>`     | Override source harness                               |
| `--to <h1,h2>`         | Override target harnesses                             |
| `--link` / `--copy`    | Force symlink or copy (overrides config)              |
| `--only skill,command` | Restrict `agents sync` to specific resource types     |
| `--dry-run` / `-n`     | Preview — write nothing                               |
| `--force` / `-f`       | Don't prompt before overwriting                       |

---

## Configuration

`agents-cli` reads `.agents.toml` at the project root (and
`~/.config/agents-cli/config.toml` globally). Any field can be overridden
by the per-command CLI flags.

```toml
# .agents.toml
source = "opencode"                       # your primary harness
targets = ["claude", "gemini", "droid"]   # where to sync to
sync_method = "symlink"                   # "symlink" or "copy"

# Optional per-resource overrides:
[methods]
skill    = "symlink"
command  = "copy"
subagent = "copy"
```

---

## Supported harnesses

Every row is a real, verified path convention as of April 2026.

| Harness         | id          | Guideline               | Skills                               | Commands                              | Subagents                              |
|-----------------|-------------|-------------------------|--------------------------------------|---------------------------------------|----------------------------------------|
| OpenCode        | `opencode`  | `AGENTS.md`             | `.opencode/skills/<n>/SKILL.md`      | `.opencode/commands/<n>.md`           | `.opencode/agents/<n>.md`              |
| Claude Code     | `claude`    | `CLAUDE.md`             | `.claude/skills/<n>/SKILL.md`        | `.claude/commands/<n>.md`             | `.claude/agents/<n>.md`                |
| Gemini CLI      | `gemini`    | `GEMINI.md`             | `.gemini/skills/<n>/SKILL.md`        | `.gemini/commands/<n>.toml` (TOML!)   | `.gemini/agents/<n>.md`                |
| Droid (Factory) | `droid`     | `AGENTS.md`             | `.factory/skills/<n>/SKILL.md`       | `.factory/commands/<n>.md`            | `.factory/droids/<n>.md`               |
| Kilo Code       | `kilo`      | `AGENTS.md` / `CLAUDE.md` | `.kilo/skills/<n>/SKILL.md`        | `.kilo/commands/<n>.md`               | `.kilo/agents/<n>.md`                  |
| Codex           | `codex`     | `AGENTS.md`             | `.codex/skills/<n>/SKILL.md`         | —                                     | —                                      |
| Cursor          | `cursor`    | `AGENTS.md`             | —                                    | `.cursor/commands/<n>.md`             | —                                      |
| Amp Code        | `amp`       | `AGENTS.md`             | `.agents/skills/<n>/SKILL.md`        | `.agents/commands/<n>.md`             | —                                      |
| Qwen Code       | `qwen`      | `QWEN.md`               | —                                    | `.qwen/commands/<n>.toml`             | —                                      |
| GitHub Copilot  | `copilot`   | `.github/copilot-instructions.md` | —                          | —                                     | —                                      |

`agents-cli` also knows each harness's **alias paths** (e.g. OpenCode reading
`CLAUDE.md`, Codex reading `AGENTS.md` from multiple locations, skills honoring
the `.agents/skills/` cross-tool standard, etc.), so discovery finds files no
matter which convention they were created under.

---

## How sync works

1. **Detect source**: Whichever harness owns your files (`opencode` by default).
2. **Enumerate** every resource that exists under the source.
3. **For each target harness**, compute the destination path from that target's
   conventions (not the source's).
4. **Symlink or copy** — your choice. Symlinks use the shortest relative path
   possible, so your repo stays portable.
5. **Skip intelligently**: same path, identical content, format mismatch
   (e.g. markdown source → TOML target for Gemini commands), or missing source.

Every operation is logged; `--dry-run` shows exactly what would happen.

---

## Safety

- `--dry-run` is always available.
- Interactive confirmation before overwriting anything that isn't already a
  symlink to the correct source. Use `--force` to skip.
- Symlinks pointing at the correct target are detected and skipped silently.

---

## Development

```bash
uv sync                     # install project + dev deps
uv run agents --help        # run from checkout
uv run pytest               # tests
uv run ruff check .         # lint
uv run ruff format .        # format
uv run mypy src             # types
```

See `AGENTS.md` for contributor-level guidelines.

---

## License

MIT
