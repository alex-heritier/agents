"""``agents doctor`` — environment + config sanity check."""

from __future__ import annotations

import shutil
import sys
from pathlib import Path
from typing import Annotated

import typer

from agents_cli import __version__, ui
from agents_cli import config as cfgmod
from agents_cli.discovery import discover
from agents_cli.harnesses import ALL_HARNESSES, ResourceType
from agents_cli.paths import Scope, detect


def _summarize_source(ctx, cfg: cfgmod.AgentsConfig) -> None:
    ui.section("Configuration")
    ui.hint(f"agents-cli version: {__version__}")
    if ctx.project_root is not None:
        ui.hint(f"project root: {ctx.project_root}")
    else:
        ui.hint("project root: (none — run `agents init` inside a git repo)")
    ui.hint(f"home: {ctx.home}")
    ui.hint(f"source harness: {cfg.source}")
    ui.hint(f"targets: {', '.join(cfg.targets) if cfg.targets else '(all)'}")
    ui.hint(f"sync method: {cfg.sync_method}")
    errs = cfgmod.validate(cfg)
    if errs:
        ui.warn("Configuration issues:")
        for e in errs:
            ui.warn(f"  • {e}")


def _render_harness_grid(ctx) -> None:
    ui.section("Harness discovery")
    rows: list[tuple[str, ...]] = []
    for h in ALL_HARNESSES:
        counts = []
        for rtype in ResourceType:
            if not h.supports(rtype):
                counts.append("—")
                continue
            scope = Scope.PROJECT if ctx.in_project else Scope.GLOBAL
            present = discover(rtype, scope, ctx, harnesses=(h,))
            # Also count global scope for a fuller picture.
            global_present = discover(rtype, Scope.GLOBAL, ctx, harnesses=(h,))
            total = len(present) + (len(global_present) if scope is Scope.PROJECT else 0)
            counts.append(str(total) if total else "·")
        rows.append((h.display_name, *counts, h.notes or ""))
    ui.table(
        None,
        columns=("Harness", "Guidelines", "Skills", "Commands", "Subagents", "Notes"),
        rows=rows,
    )


def run(
    verbose: Annotated[bool, typer.Option("--verbose", "-v")] = False,
) -> None:
    ctx = detect()
    cfg = cfgmod.load(ctx.project_root)

    ui.banner("agents doctor", "diagnostic information for your setup")
    _summarize_source(ctx, cfg)
    _render_harness_grid(ctx)

    ui.section("Environment")
    ui.hint(f"python: {sys.version.split()[0]} ({sys.executable})")
    ui.hint(f"cwd: {Path.cwd()}")

    for cli_name in ("opencode", "claude", "gemini", "droid", "kilo", "codex", "cursor", "copilot"):
        path = shutil.which(cli_name)
        if path:
            ui.hint(f"  {cli_name}: {path}")
        elif verbose:
            ui.hint(f"  {cli_name}: (not on $PATH)")
