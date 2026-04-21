# agents-cli — contributor guide

> Source of truth for **how this project is built**. (User-facing docs live in
> `README.md`.)

## What this is

A Python 3.10+ CLI, built on [Typer][typer] + [Rich][rich], packaged with
[`uv`][uv]. Its only job: maintain one set of AI-agent configuration files
(guidelines, skills, slash commands, subagents) and sync them — via symlink
or copy — to every harness the user might touch (OpenCode, Claude Code,
Gemini CLI, Droid, Kilo, Codex, Cursor, Amp, Qwen, Copilot).

## Repository layout

```
src/agents_cli/
  app.py                # Typer app wiring + top-level commands
  harnesses.py          # Knowledge base: paths for every harness
  paths.py              # Scope detection (project/global), display_path
  fs.py                 # Safe filesystem ops (symlink/copy/rm/write) + OperationLog
  discovery.py          # "What resources exist on disk?"
  sync.py               # Propagation engine
  templates.py          # Default content for `new` commands
  ui.py                 # Rich-based output helpers
  config.py             # .agents.toml loader + merger
  commands/
    _shared.py          # Builds the list/show/new/edit/rm/sync verbs per type
    guideline_group.py
    skill_group.py
    command_group.py
    subagent_group.py
    init_cmd.py
    doctor_cmd.py
    status_cmd.py
    sync_cmd.py
```

The one file you'll edit most often for new harnesses: `harnesses.py`.

## Dev workflow

```bash
uv sync                       # install project + dev deps
uv run agents --help          # run the CLI from source
uv run agents doctor          # sanity-check against current repo
uv run pytest                 # tests
uv run ruff check .           # lint
uv run ruff format .          # format
uv run mypy src               # types
make check                    # everything above in one go
```

We target **Python 3.10+**; don't use 3.11/3.12-only APIs (e.g.
`Path.is_dir(follow_symlinks=…)`).

## Design principles

1. **Be conservative with the filesystem.** Every write goes through
   `fs.py`, which supports `dry_run`, `force`, interactive confirmation,
   identical-content skipping, and full operation logging.
2. **Aliases are first-class.** Every harness has primary paths and
   alias paths. Discovery must find both. Writes always use the primary.
3. **Symlinks are relative.** Portable repos > absolute paths.
4. **Harness knowledge = data, not code.** `harnesses.py` is declarative;
   adding a harness should not require new logic anywhere else.
5. **Delightful CLI.** Rich banners, tables, sections, clear status marks.
   Use `ui.py`, not bare prints.
6. **Guidelines are nested.** Project walk handles subdir `AGENTS.md`.
   All other resource types are flat per scope.

## Adding a new harness

1. Open `src/agents_cli/harnesses.py`.
2. Define a `Harness(id=…, guideline=…, skill=…, command=…, subagent=…)`.
3. Add it to `ALL_HARNESSES`.
4. Run `agents doctor` — it should appear in the discovery grid.
5. Add a row to the README table.
6. If it has unusual rules (format mismatch, special aliases), verify
   `sync.py`'s `_format_mismatch` does the right thing.

## Adding a new resource type

Rare. If you do:

1. Add a value to `ResourceType` in `harnesses.py`.
2. Extend each `Harness` to specify (or explicitly omit) its `ResourceSpec`.
3. Teach `discovery._discover_file_resources` / `sync.run_sync` about it.
4. Add a command group under `commands/`, mirroring `skill_group.py`.
5. Register it in `app.py`.

## Testing

- Manual smoke: `uv run agents doctor`, then `agents sync --dry-run` in a
  scratch repo (see the README Quickstart).
- Unit tests live in `tests/`. Tests should be filesystem-hermetic: use
  `tmp_path` and never touch `$HOME`.

## Style

- `ruff` enforces import order, modern Python idioms, pytest style.
- Prefer `dataclass(frozen=True)` for records.
- Keep modules small and purpose-focused.
- Public functions get docstrings; private helpers don't need them unless
  non-obvious.
- No redundant `# this does X` comments. Comments explain *why*.

## Commit / PR hygiene

- One logical change per commit; keep them atomic.
- Don't commit binary artifacts, virtualenvs, or `.DS_Store`.
- `uv.lock` is checked in; update it with `uv lock` whenever deps change.

## License

MIT — same as the project.

[typer]: https://typer.tiangolo.com/
[rich]: https://rich.readthedocs.io/
[uv]: https://docs.astral.sh/uv/
