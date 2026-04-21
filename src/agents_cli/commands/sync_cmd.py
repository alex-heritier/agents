"""Top-level ``agents sync`` — sync every resource type in one shot."""

from __future__ import annotations

from typing import Annotated

import typer

from agents_cli import config as cfgmod
from agents_cli import fs, sync, ui
from agents_cli.commands._shared import _render_log, pick_scope
from agents_cli.harnesses import ResourceType
from agents_cli.paths import detect


def _parse_list(raw: str | None) -> tuple[str, ...]:
    if not raw:
        return ()
    return tuple(part.strip() for part in raw.split(",") if part.strip())


def run_all(
    source: Annotated[
        str | None, typer.Option("--from", help="Source harness (default: config).")
    ] = None,
    to: Annotated[
        str | None, typer.Option("--to", help="Target harnesses (comma-separated).")
    ] = None,
    method: Annotated[
        str | None, typer.Option("--method", help="symlink or copy (default: config).")
    ] = None,
    link: Annotated[bool, typer.Option("--link", help="Force symlink.")] = False,
    copy: Annotated[bool, typer.Option("--copy", help="Force copy.")] = False,
    only: Annotated[
        str | None,
        typer.Option(
            "--only",
            help=(
                "Only sync these resource types (comma-separated: "
                "guideline,skill,command,subagent)."
            ),
        ),
    ] = None,
    project_only: Annotated[bool, typer.Option("--project", "-p")] = False,
    global_only: Annotated[bool, typer.Option("--global", "-g")] = False,
    dry_run: Annotated[bool, typer.Option("--dry-run", "-n")] = False,
    force: Annotated[bool, typer.Option("--force", "-f")] = False,
) -> None:
    if link and copy:
        raise typer.BadParameter("Pass --link or --copy, not both.")
    if link:
        method = "symlink"
    if copy:
        method = "copy"
    if method and method not in {"symlink", "copy"}:
        raise typer.BadParameter("--method must be 'symlink' or 'copy'")

    ctx = detect()
    cfg = cfgmod.load(ctx.project_root)
    scopes = pick_scope(project_only, global_only, ctx)
    targets = _parse_list(to)
    only_types_raw = _parse_list(only)
    only_types: set[ResourceType]
    if only_types_raw:
        only_types = set()
        for t in only_types_raw:
            try:
                only_types.add(ResourceType(t.lower()))
            except ValueError:
                raise typer.BadParameter(f"Unknown resource type: {t!r}") from None
    else:
        only_types = set(ResourceType)

    ui.banner(
        "agents sync",
        (
            f"source={source or cfg.source}  "
            f"targets={','.join(targets) if targets else ','.join(cfg.targets) or 'all'}  "
            f"method={method or cfg.sync_method}  "
            f"scope={'/'.join(s.value for s in scopes)}"
        ),
    )

    total_log = fs.OperationLog()
    for rtype in ResourceType:
        if rtype not in only_types:
            continue
        ui.section(f"{rtype.value.title()}s")
        log = fs.OperationLog()
        for scope in scopes:
            partial = sync.run_sync(
                rtype,
                scope,
                ctx,
                cfg,
                source_override=source,
                target_overrides=targets,
                method_override=method,
                dry_run=dry_run,
                force=force,
                confirm=lambda p: ui.confirm(p),
            )
            log.extend(partial)
        _render_log(log, dry_run=dry_run)
        total_log.extend(log)

    ui.section("Summary")
    ui.format_summary(
        created=total_log.created,
        updated=total_log.updated,
        skipped=total_log.skipped,
        deleted=total_log.deleted,
        failed=total_log.failed,
        dry_run=dry_run,
    )
