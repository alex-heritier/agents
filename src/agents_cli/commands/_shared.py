"""Reusable command-building blocks for all resource-group subcommands.

Every resource type (guideline, skill, command, subagent) shares the
same CLI verbs and flags. Rather than copy-paste them four times, we
build a ``ResourceApp`` factory here.
"""

import os
import subprocess
import sys
from pathlib import Path
from typing import Annotated

import typer

from agents_cli import config as cfgmod
from agents_cli import fs, sync, templates, ui
from agents_cli.discovery import DiscoveredResource, discover, find
from agents_cli.harnesses import ALL_HARNESSES, Harness, ResourceType, get, iter_supporting
from agents_cli.paths import Scope, ScopeContext, detect, display_path, render_paths

_METHODS = ("symlink", "copy")


def _harness_option(default_help: str) -> Annotated[str | None, typer.Option]:
    return typer.Option(
        None,
        "--harness",
        "-H",
        help=default_help,
    )


def _scope_options() -> tuple[Annotated[bool, typer.Option], Annotated[bool, typer.Option]]:
    return (
        typer.Option(False, "--project", "-p", help="Operate on project scope only."),
        typer.Option(False, "--global", "-g", help="Operate on global (~) scope only."),
    )


def pick_scope(project_only: bool, global_only: bool, ctx: ScopeContext) -> tuple[Scope, ...]:
    if project_only and global_only:
        raise typer.BadParameter("Cannot combine --project and --global.")
    if project_only:
        if not ctx.in_project:
            raise typer.BadParameter("No git project detected (no .git found above cwd).")
        return (Scope.PROJECT,)
    if global_only:
        return (Scope.GLOBAL,)
    # Default: both, biased to project when inside one.
    if ctx.in_project:
        return (Scope.PROJECT, Scope.GLOBAL)
    return (Scope.GLOBAL,)


def _parse_harness_list(raw: str | None) -> tuple[str, ...]:
    if not raw:
        return ()
    return tuple(part.strip() for part in raw.split(",") if part.strip())


def _resolve_harness_for_write(
    raw: str | None, cfg: cfgmod.AgentsConfig, rtype: ResourceType
) -> Harness:
    h = get(raw) if raw else get(cfg.source)
    if not h.supports(rtype):
        raise typer.BadParameter(
            f"{h.display_name} does not support {rtype.value}. "
            f"Pick one of: {', '.join(x.id for x in iter_supporting(rtype))}"
        )
    return h


def _row_for(res: DiscoveredResource, ctx: ScopeContext) -> tuple[str, ...]:
    scope_label = "project" if res.scope is Scope.PROJECT else "global"
    mark = ""
    if res.is_symlink:
        mark = "↪ "
    elif res.is_alias_location:
        mark = "~ "
    name = mark + res.display_name()
    size_str = _format_size(res.size)
    return (
        res.harness_id,
        scope_label,
        name,
        display_path(res.path, ctx),
        size_str,
    )


def _format_size(size: int) -> str:
    if size < 1024:
        return f"{size}B"
    if size < 1024 * 1024:
        return f"{size // 1024}K"
    return f"{size // (1024 * 1024)}M"


def _open_in_editor(path: Path) -> int:
    editor = os.environ.get("VISUAL") or os.environ.get("EDITOR") or "nano"
    try:
        return subprocess.call([editor, str(path)])
    except FileNotFoundError:
        ui.error(f"Editor '{editor}' not found. Set $EDITOR to a command on your PATH.")
        return 1


# ---------------------------------------------------------------------------
# Helpers for interactive show/edit
# ---------------------------------------------------------------------------


def _resolve_scope_for_show(ctx: ScopeContext, project_only: bool, global_only: bool) -> "Scope":
    """Scope resolution for ``show``: project when in a project, global otherwise."""
    return Scope.PROJECT if (project_only or (not global_only and ctx.in_project)) else Scope.GLOBAL


def _resolve_scope_for_edit(ctx: ScopeContext, project_only: bool, global_only: bool) -> "Scope":
    """Scope resolution for ``edit``: global when --global, else project with fallback."""
    scope = Scope.GLOBAL if global_only else Scope.PROJECT
    if scope is Scope.PROJECT and not ctx.in_project:
        scope = Scope.GLOBAL
    return scope


def _discover_candidates(
    *,
    rtype: ResourceType,
    ctx: ScopeContext,
    cfg: "cfgmod.AgentsConfig",
    scope: "Scope",
    harness_raw: str | None,
) -> list[DiscoveredResource]:
    """Discover resources filtered to the effective harness set.

    When ``harness_raw`` is ``None`` we mirror the exact-mode default: only
    the configured source harness, keeping interactive and exact modes aligned.
    """
    if harness_raw:
        harness_ids = _parse_harness_list(harness_raw)
        harness_filter: tuple[Harness, ...] | None = tuple(get(h) for h in harness_ids)
    else:
        # Same harness as exact mode would pick.
        default_h = _resolve_harness_for_write(None, cfg, rtype)
        harness_filter = (default_h,)

    results = discover(rtype, scope, ctx, harnesses=harness_filter)
    results.sort(key=lambda r: (r.scope.value, r.harness_id, r.name, str(r.path)))
    return results


def _pick_resource_interactive(
    *,
    resources: list[DiscoveredResource],
    ctx: ScopeContext,
    title: str,
) -> DiscoveredResource | None:
    """Render a numbered picker and return the chosen resource, or ``None`` on cancel.

    When there is exactly one candidate the prompt is skipped and that
    resource is returned immediately.
    """
    if len(resources) == 1:
        return resources[0]

    rows: list[tuple[str, ...]] = []
    for i, res in enumerate(resources, start=1):
        mark = ""
        if res.is_symlink:
            mark = "↪ "
        elif res.is_alias_location:
            mark = "~ "
        scope_label = "project" if res.scope is Scope.PROJECT else "global"
        rows.append(
            (
                str(i),
                res.harness_id,
                scope_label,
                mark + res.display_name(),
                display_path(res.path, ctx),
                _format_size(res.size),
            )
        )

    ui.banner(title)
    ui.table(
        None,
        columns=("#", "Harness", "Scope", "Name", "Path", "Size"),
        rows=rows,
    )

    idx = ui.prompt_choice(len(resources))
    if idx is None:
        return None
    return resources[idx]


def _show_resource(res: DiscoveredResource, ctx: ScopeContext) -> None:
    """Read ``res.entry`` and display it through ``ui.view_text``."""
    h_name = get(res.harness_id).display_name
    title = f"{h_name} · {res.rtype.value} · {res.display_name()}"
    subtitle = display_path(res.entry, ctx)
    try:
        content = res.entry.read_text()
    except OSError as e:
        ui.error(f"Read failed: {e}")
        raise typer.Exit(1) from None
    ui.view_text(title, subtitle, content)


def build_group(
    *,
    name: str,
    rtype: ResourceType,
    help_text: str,
) -> typer.Typer:
    """Build a Typer app exposing list/show/new/edit/rm/sync for a resource."""
    app = typer.Typer(
        name=name,
        help=help_text,
        no_args_is_help=True,
        rich_markup_mode="rich",
        context_settings={"help_option_names": ["-h", "--help"]},
    )

    @app.command("list", help=f"List all {rtype.value}s across harnesses.")
    def list_cmd(
        harness: Annotated[
            str | None,
            typer.Option("--harness", "-H", help="Filter by harness id (comma-separated)."),
        ] = None,
        project_only: Annotated[
            bool, typer.Option("--project", "-p", help="Project scope only.")
        ] = False,
        global_only: Annotated[
            bool, typer.Option("--global", "-g", help="Global scope only.")
        ] = False,
        verbose: Annotated[
            bool, typer.Option("--verbose", "-v", help="Show symlink targets.")
        ] = False,
    ) -> None:
        ctx = detect()
        scopes = pick_scope(project_only, global_only, ctx)
        harness_ids = _parse_harness_list(harness)
        harness_filter: tuple[Harness, ...] | None = (
            tuple(get(h) for h in harness_ids) if harness_ids else None
        )

        rows: list[tuple] = []
        for scope in scopes:
            results = discover(rtype, scope, ctx, harnesses=harness_filter)
            # Sort: scope then harness then name.
            results.sort(key=lambda r: (r.scope.value, r.harness_id, r.name))
            for r in results:
                rows.append(_row_for(r, ctx))
                if verbose and r.is_symlink and r.symlink_target is not None:
                    rows.append(("", "", f"    → {r.symlink_target}", "", ""))

        subtitle = f"scope={'/'.join(s.value for s in scopes)}"
        ui.banner(f"{rtype.value.title()}s", subtitle)
        ui.table(
            None,
            columns=("Harness", "Scope", "Name", "Path", "Size"),
            rows=rows,
        )

    @app.command("show", help=f"Show a {rtype.value}. If NAME is omitted, pick interactively.")
    def show_cmd(
        name_arg: Annotated[
            str | None,
            typer.Argument(
                metavar="NAME",
                help=f"{rtype.value} name or filename.",
                show_default=False,
            ),
        ] = None,
        harness: Annotated[
            str | None,
            typer.Option("--harness", "-H", help="Which harness's copy to show."),
        ] = None,
        project_only: Annotated[bool, typer.Option("--project", "-p")] = False,
        global_only: Annotated[bool, typer.Option("--global", "-g")] = False,
    ) -> None:
        ctx = detect()
        cfg = cfgmod.load(ctx.project_root)
        scope = _resolve_scope_for_show(ctx, project_only, global_only)

        if name_arg is not None:
            # Exact mode — unchanged behavior.
            h = _resolve_harness_for_write(harness, cfg, rtype)
            res = find(h, rtype, name_arg, scope, ctx)
            if res is None:
                ui.error(
                    f"No {rtype.value} named {name_arg!r} in {h.display_name} at {scope.value} scope."
                )
                raise typer.Exit(1)
            _show_resource(res, ctx)
        else:
            # Interactive mode.
            if not sys.stdin.isatty():
                ui.error("Interactive selection requires a TTY. Pass NAME explicitly.")
                raise typer.Exit(2)
            candidates = _discover_candidates(
                rtype=rtype, ctx=ctx, cfg=cfg, scope=scope, harness_raw=harness
            )
            if not candidates:
                ui.error(f"No {rtype.value}s found.")
                raise typer.Exit(1)
            res = _pick_resource_interactive(
                resources=candidates,
                ctx=ctx,
                title=f"Select a {rtype.value} to view",
            )
            if res is None:
                raise typer.Exit(0)
            _show_resource(res, ctx)

    @app.command("new", help=f"Create a new {rtype.value}.")
    def new_cmd(
        name_arg: Annotated[str, typer.Argument(metavar="NAME")],
        harness: Annotated[
            str | None,
            typer.Option(
                "--harness", "-H", help="Which harness to write into (default: config source)."
            ),
        ] = None,
        project_only: Annotated[bool, typer.Option("--project", "-p")] = False,
        global_only: Annotated[bool, typer.Option("--global", "-g")] = False,
        force: Annotated[
            bool, typer.Option("--force", "-f", help="Overwrite without prompting.")
        ] = False,
        edit: Annotated[
            bool, typer.Option("--edit/--no-edit", help="Open in $EDITOR after creating.")
        ] = True,
        dry_run: Annotated[
            bool, typer.Option("--dry-run", "-n", help="Show what would happen.")
        ] = False,
    ) -> None:
        ctx = detect()
        cfg = cfgmod.load(ctx.project_root)

        if project_only and global_only:
            raise typer.BadParameter("Cannot combine --project and --global.")
        scope = Scope.GLOBAL if global_only else Scope.PROJECT
        if scope is Scope.PROJECT and not ctx.in_project:
            ui.warn("No git project detected — falling back to global scope.")
            scope = Scope.GLOBAL

        h = _resolve_harness_for_write(harness, cfg, rtype)
        rendered = render_paths(h, rtype, name_arg, scope, ctx, include_aliases=False)
        if not rendered:
            ui.error(f"{h.display_name} has no {rtype.value} path at {scope.value} scope.")
            raise typer.Exit(1)
        target_path = Path(rendered[0].path)

        # rendered path already points to the entry file (e.g. SKILL.md)
        # for directory resources, so the same value is used either way.
        file_path = target_path

        content = templates.render(rtype, h, name_arg)
        log = fs.OperationLog()
        result = fs.write_text(
            file_path,
            content,
            dry_run=dry_run,
            force=force,
            confirm=lambda p: ui.confirm(p),
            log=log,
        )
        _render_log(log, dry_run=dry_run)
        if result is fs.Op.FAILED:
            raise typer.Exit(1)
        if edit and not dry_run and result is not fs.Op.SKIPPED:
            _open_in_editor(file_path)

    @app.command("edit", help=f"Edit a {rtype.value}. If NAME is omitted, pick interactively.")
    def edit_cmd(
        name_arg: Annotated[
            str | None,
            typer.Argument(
                metavar="NAME",
                help=f"{rtype.value} name or filename.",
                show_default=False,
            ),
        ] = None,
        harness: Annotated[str | None, typer.Option("--harness", "-H")] = None,
        project_only: Annotated[bool, typer.Option("--project", "-p")] = False,
        global_only: Annotated[bool, typer.Option("--global", "-g")] = False,
    ) -> None:
        ctx = detect()
        cfg = cfgmod.load(ctx.project_root)
        scope = _resolve_scope_for_edit(ctx, project_only, global_only)

        if name_arg is not None:
            # Exact mode — unchanged behavior.
            h = _resolve_harness_for_write(harness, cfg, rtype)
            res = find(h, rtype, name_arg, scope, ctx)
            if res is None:
                ui.error(
                    f"No {rtype.value} named {name_arg!r} in {h.display_name} at {scope.value} scope."
                )
                raise typer.Exit(1)
            raise typer.Exit(_open_in_editor(res.entry))
        else:
            # Interactive mode.
            if not sys.stdin.isatty():
                ui.error("Interactive selection requires a TTY. Pass NAME explicitly.")
                raise typer.Exit(2)
            candidates = _discover_candidates(
                rtype=rtype, ctx=ctx, cfg=cfg, scope=scope, harness_raw=harness
            )
            if not candidates:
                ui.error(f"No {rtype.value}s found.")
                raise typer.Exit(1)
            res = _pick_resource_interactive(
                resources=candidates,
                ctx=ctx,
                title=f"Select a {rtype.value} to edit",
            )
            if res is None:
                raise typer.Exit(0)
            raise typer.Exit(_open_in_editor(res.entry))

    @app.command("rm", help=f"Delete a {rtype.value}.")
    def rm_cmd(
        name_arg: Annotated[str, typer.Argument(metavar="NAME")],
        harness: Annotated[
            str | None,
            typer.Option(
                "--harness",
                "-H",
                help="Which harness to delete from. Omit to delete from all harnesses that have it.",
            ),
        ] = None,
        project_only: Annotated[bool, typer.Option("--project", "-p")] = False,
        global_only: Annotated[bool, typer.Option("--global", "-g")] = False,
        force: Annotated[
            bool, typer.Option("--force", "-f", help="Delete without prompting.")
        ] = False,
        dry_run: Annotated[bool, typer.Option("--dry-run", "-n", help="Preview only.")] = False,
    ) -> None:
        ctx = detect()
        scopes = pick_scope(project_only, global_only, ctx)
        if harness:
            harnesses = tuple(get(h) for h in _parse_harness_list(harness))
        else:
            harnesses = tuple(h for h in ALL_HARNESSES if h.supports(rtype))

        log = fs.OperationLog()
        for scope in scopes:
            for h in harnesses:
                res = find(h, rtype, name_arg, scope, ctx)
                if res is None:
                    continue
                fs.remove(
                    res.path,
                    dry_run=dry_run,
                    force=force,
                    confirm=lambda p: ui.confirm(p),
                    log=log,
                )
        _render_log(log, dry_run=dry_run)

    @app.command("sync", help=f"Sync {rtype.value}s from source harness to targets.")
    def sync_cmd_inner(
        source: Annotated[
            str | None,
            typer.Option("--from", help="Source harness (default: config)."),
        ] = None,
        to: Annotated[
            str | None,
            typer.Option("--to", help="Target harnesses, comma-separated (default: config)."),
        ] = None,
        method: Annotated[
            str | None,
            typer.Option("--method", help=f"sync method: {'/'.join(_METHODS)} (default: config)."),
        ] = None,
        link: Annotated[bool, typer.Option("--link", help="Force symlink method.")] = False,
        copy: Annotated[bool, typer.Option("--copy", help="Force copy method.")] = False,
        project_only: Annotated[bool, typer.Option("--project", "-p")] = False,
        global_only: Annotated[bool, typer.Option("--global", "-g")] = False,
        names: Annotated[
            str | None,
            typer.Option("--name", help="Only sync these (comma-separated) names."),
        ] = None,
        dry_run: Annotated[bool, typer.Option("--dry-run", "-n")] = False,
        force: Annotated[bool, typer.Option("--force", "-f")] = False,
    ) -> None:
        if link and copy:
            raise typer.BadParameter("Pass --link or --copy, not both.")
        if link:
            method = "symlink"
        if copy:
            method = "copy"
        if method and method not in _METHODS:
            raise typer.BadParameter(f"--method must be one of {_METHODS}")

        ctx = detect()
        cfg = cfgmod.load(ctx.project_root)
        scopes = pick_scope(project_only, global_only, ctx)
        targets = _parse_harness_list(to)
        chosen_names = _parse_harness_list(names)

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
                names=chosen_names,
                confirm=lambda p: ui.confirm(p),
            )
            log.extend(partial)

        _render_log(log, dry_run=dry_run)

    return app


def _render_log(log: fs.OperationLog, *, dry_run: bool) -> None:
    for op in log.operations:
        if op.op is fs.Op.FAILED:
            ui.error(f"  {op.format()}")
        elif op.op is fs.Op.SKIPPED:
            ui.hint(op.format())
        elif op.op is fs.Op.CREATED:
            ui.console.print(f"  [success]+[/success] {op.format()}")
        elif op.op is fs.Op.UPDATED:
            ui.console.print(f"  [accent]~[/accent] {op.format()}")
        elif op.op is fs.Op.DELETED:
            ui.console.print(f"  [warn]-[/warn] {op.format()}")
    ui.console.print()
    ui.format_summary(
        created=log.created,
        updated=log.updated,
        skipped=log.skipped,
        deleted=log.deleted,
        failed=log.failed,
        dry_run=dry_run,
    )
