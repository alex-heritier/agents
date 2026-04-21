"""``agents init`` — scaffold a project ``.agents.toml`` and source AGENTS.md."""

from __future__ import annotations

from pathlib import Path
from textwrap import dedent
from typing import Annotated

import typer

from agents_cli import config as cfgmod
from agents_cli import fs, ui
from agents_cli.harnesses import ALL_HARNESSES, Harness, ResourceType, get
from agents_cli.paths import Scope, detect, render_paths
from agents_cli.templates import guideline as guideline_template


def _build_toml(source: str, targets: tuple[str, ...], method: str) -> str:
    target_list = ", ".join(f'"{t}"' for t in targets)
    return dedent(
        f"""\
        # agents-cli project configuration.
        # Run `agents sync` to propagate resources from the source
        # harness to every target harness.

        source = "{source}"
        targets = [{target_list}]
        sync_method = "{method}"  # "symlink" or "copy"

        # Optional per-resource overrides:
        # [methods]
        # skill = "symlink"
        # command = "copy"
        """
    )


def run(
    source: Annotated[
        str | None,
        typer.Option("--source", "-s", help="Primary harness you maintain (default: opencode)."),
    ] = None,
    targets: Annotated[
        str | None,
        typer.Option("--targets", "-t", help="Comma-separated target harnesses. Empty = all."),
    ] = None,
    method: Annotated[
        str | None,
        typer.Option("--method", "-m", help="Default sync method: symlink or copy."),
    ] = None,
    seed: Annotated[
        bool,
        typer.Option(
            "--seed/--no-seed", help="Create an AGENTS.md in the source harness if missing."
        ),
    ] = True,
    force: Annotated[bool, typer.Option("--force", "-f")] = False,
    dry_run: Annotated[bool, typer.Option("--dry-run", "-n")] = False,
) -> None:
    ctx = detect()
    if ctx.project_root is None:
        ui.error("No git repository detected. Run `agents init` inside a git project.")
        raise typer.Exit(2)

    chosen_source = (source or cfgmod.DEFAULT_SOURCE).lower()
    try:
        src_harness: Harness = get(chosen_source)
    except KeyError:
        ui.error(f"Unknown harness {chosen_source!r}.")
        raise typer.Exit(2) from None

    chosen_method = (method or cfgmod.DEFAULT_METHOD).lower()
    if chosen_method not in {"symlink", "copy"}:
        ui.error("--method must be 'symlink' or 'copy'.")
        raise typer.Exit(2)

    if targets is not None:
        target_ids = tuple(t.strip().lower() for t in targets.split(",") if t.strip())
    else:
        target_ids = tuple(h.id for h in ALL_HARNESSES if h.id != src_harness.id)

    for t in target_ids:
        try:
            get(t)
        except KeyError:
            ui.error(f"Unknown target harness: {t!r}")
            raise typer.Exit(2) from None

    log = fs.OperationLog()
    toml_path = cfgmod.project_config_path(ctx.project_root)
    fs.write_text(
        toml_path,
        _build_toml(src_harness.id, target_ids, chosen_method),
        dry_run=dry_run,
        force=force,
        confirm=lambda p: ui.confirm(p),
        log=log,
    )

    if seed and src_harness.guideline is not None:
        rendered = render_paths(
            src_harness, ResourceType.GUIDELINE, "", Scope.PROJECT, ctx, include_aliases=False
        )
        if rendered:
            path = Path(rendered[0].path)
            if not path.exists():
                fs.write_text(
                    path,
                    guideline_template(ctx.project_root.name),
                    dry_run=dry_run,
                    force=force,
                    confirm=lambda p: ui.confirm(p),
                    log=log,
                )

    ui.banner(
        "agents init",
        f"source={src_harness.id}  targets={','.join(target_ids) or '∅'}  method={chosen_method}",
    )
    for op in log.operations:
        if op.op is fs.Op.FAILED:
            ui.error(f"  {op.format()}")
        elif op.op is fs.Op.SKIPPED:
            ui.hint(op.format())
        elif op.op is fs.Op.CREATED:
            ui.console.print(f"  [success]+[/success] {op.format()}")
    ui.console.print()
    ui.console.print("Next steps:")
    ui.hint(f"  • edit {toml_path.name} to tweak targets / method")
    if src_harness.guideline is not None:
        ui.hint(f"  • fill out your {src_harness.guideline.project_path}")
    ui.hint("  • run `agents sync --dry-run` to preview propagation")
    ui.hint("  • run `agents sync` when you're ready")
