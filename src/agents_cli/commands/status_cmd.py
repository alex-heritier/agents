"""``agents status`` — one-screen overview of resources in the current scope."""

from __future__ import annotations

from typing import Annotated

import typer

from agents_cli import config as cfgmod
from agents_cli import ui
from agents_cli.discovery import discover
from agents_cli.harnesses import ALL_HARNESSES, ResourceType, get
from agents_cli.paths import Scope, ScopeContext, detect, display_path


def run(
    project_only: Annotated[bool, typer.Option("--project", "-p")] = False,
    global_only: Annotated[bool, typer.Option("--global", "-g")] = False,
) -> None:
    ctx = detect()
    cfg = cfgmod.load(ctx.project_root)

    if project_only and global_only:
        raise typer.BadParameter("Cannot combine --project and --global.")
    scopes: tuple[Scope, ...]
    if project_only:
        scopes = (Scope.PROJECT,)
    elif global_only:
        scopes = (Scope.GLOBAL,)
    else:
        scopes = (Scope.PROJECT, Scope.GLOBAL) if ctx.in_project else (Scope.GLOBAL,)

    ui.banner(
        "agents status",
        (
            f"source={cfg.source}  "
            f"targets={','.join(cfg.targets) if cfg.targets else 'all'}  "
            f"method={cfg.sync_method}"
        ),
    )

    try:
        source_harness = get(cfg.source)
    except KeyError:
        ui.error(f"Source harness {cfg.source!r} is unknown — run `agents doctor`.")
        raise typer.Exit(1) from None

    for rtype in ResourceType:
        rows: list[tuple[str, ...]] = []
        for scope in scopes:
            source_res = discover(rtype, scope, ctx, harnesses=(source_harness,))
            source_res = [r for r in source_res if not r.is_alias_location]
            if not source_res:
                continue
            for sr in source_res:
                row: list[str] = [
                    scope.value,
                    _pretty_name(sr, ctx),
                ]
                for h in ALL_HARNESSES:
                    if h.id == source_harness.id:
                        row.append(_mark(True, "source"))
                        continue
                    if not h.supports(rtype):
                        row.append("—")
                        continue
                    tgt_list = discover(rtype, scope, ctx, harnesses=(h,))
                    match = _find_match(sr, tgt_list, rtype)
                    row.append(
                        _mark(match is not None, "ok" if match and match.is_symlink else "copy")
                    )
                rows.append(tuple(row))
        columns = ["Scope", "Name", *(h.id for h in ALL_HARNESSES)]
        ui.section(f"{rtype.value.title()}s")
        ui.table(None, columns=columns, rows=rows)


def _pretty_name(res, ctx: ScopeContext) -> str:
    if res.rtype is ResourceType.GUIDELINE:
        # Distinguish guideline files by their directory (e.g. ./packages/api)
        return display_path(res.path.parent, ctx)
    return res.name


def _find_match(src_res, tgt_list, rtype: ResourceType):
    """Match a source resource against target-harness candidates.

    Guidelines: match by containing directory (names may differ: AGENTS.md
    vs CLAUDE.md vs GEMINI.md). Other types: match by name.
    """
    if rtype is ResourceType.GUIDELINE:
        src_dir = src_res.path.parent.resolve()
        for t in tgt_list:
            try:
                if t.path.parent.resolve() == src_dir:
                    return t
            except OSError:
                continue
        return None
    for t in tgt_list:
        if not t.is_alias_location and t.name == src_res.name:
            return t
    return None


def _mark(present: bool, label: str) -> str:
    if not present:
        return "·"
    if label == "source":
        return "★"
    if label == "ok":
        return "↪"
    return "●"
