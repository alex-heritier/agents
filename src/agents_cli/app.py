"""Top-level Typer application.

The CLI is organized around one resource type per subcommand group:

    agents guideline  ...    # AGENTS.md, CLAUDE.md, GEMINI.md, ...
    agents skill      ...    # SKILL.md folders
    agents command    ...    # slash commands
    agents subagent   ...    # specialized agents

Each group exposes the same set of verbs: ``list``, ``show``, ``new``,
``edit``, ``rm``, ``sync``. There is also a top-level ``sync`` that
does everything in one shot, plus ``init`` / ``doctor`` / ``status``
for onboarding.
"""

from __future__ import annotations

from typing import Annotated

import typer

from agents_cli import __version__, ui
from agents_cli.commands import (
    command_group,
    doctor_cmd,
    guideline_group,
    init_cmd,
    skill_group,
    status_cmd,
    subagent_group,
    sync_cmd,
)

app = typer.Typer(
    name="agents",
    help=(
        "Maintain one source of truth for AI coding [primary]guidelines[/primary], "
        "[primary]skills[/primary], [primary]commands[/primary], and [primary]subagents[/primary] "
        "— then sync them to every harness (opencode, Claude Code, Gemini CLI, droid, kilo, Codex, Cursor, Amp, …)."
    ),
    no_args_is_help=True,
    rich_markup_mode="rich",
    context_settings={"help_option_names": ["-h", "--help"]},
    add_completion=False,
)


def _version_callback(value: bool) -> None:
    if value:
        ui.console.print(f"agents-cli [primary]{__version__}[/primary]")
        raise typer.Exit(0)


@app.callback()
def _root(
    version: Annotated[
        bool,
        typer.Option(
            "--version",
            "-V",
            callback=_version_callback,
            is_eager=True,
            help="Show version and exit.",
        ),
    ] = False,
) -> None:
    """Global options."""


# Register resource-specific subcommand groups.
app.add_typer(
    guideline_group.app, name="guideline", help="Manage agent guideline files (AGENTS.md et al.)"
)
app.add_typer(guideline_group.app, name="rule", help="Alias for [primary]guideline[/primary].")
app.add_typer(
    skill_group.app, name="skill", help="Manage Anthropic-style skills (SKILL.md folders)."
)
app.add_typer(command_group.app, name="command", help="Manage custom slash commands.")
app.add_typer(command_group.app, name="cmd", help="Alias for [primary]command[/primary].")
app.add_typer(subagent_group.app, name="subagent", help="Manage specialized subagents.")
app.add_typer(subagent_group.app, name="agent", help="Alias for [primary]subagent[/primary].")

# Top-level convenience commands.
app.command("init", help="Create a project agents.toml in the current git root.")(init_cmd.run)
app.command("doctor", help="Show detected harnesses, scope, and config — a health check.")(
    doctor_cmd.run
)
app.command("status", help="One-screen overview of what's present and in sync.")(status_cmd.run)
app.command("sync", help="Sync every resource type from source harness to all targets.")(
    sync_cmd.run_all
)


def main() -> None:  # entry point for ``agents``
    app()


if __name__ == "__main__":
    main()
