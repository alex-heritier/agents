"""Sync engine: propagate resources from a source harness to targets.

Sync is *always* based on an existing on-disk source. We never invent
content. For guidelines, sync is per-directory — if the source has a
nested AGENTS.md in ``packages/foo/``, the same nested path is mirrored
for every target harness (and only targets that use a different filename
actually receive a link/copy).

Default behaviour for each resource type:

========  ===========  ==============================================
Type      Default      Reason
========  ===========  ==============================================
guideline symlink      Many files, content identical — symlinks keep
                        them in sync automatically.
skill     symlink      Directory — symlinking the whole tree is cheap.
command   symlink      Format (.md) is identical across harnesses that
                        use markdown. Gemini/Qwen (TOML) is skipped
                        with a warning since the format doesn't match.
subagent  symlink      Same as command.
========  ===========  ==============================================

Users can force ``--copy`` to materialize independent copies (useful
when one target needs divergent content).
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from pathlib import Path

from agents_cli import fs
from agents_cli.config import AgentsConfig
from agents_cli.discovery import DiscoveredResource, discover
from agents_cli.harnesses import ALL_HARNESSES, Harness, Kind, ResourceType, get
from agents_cli.paths import Scope, ScopeContext, render_paths


class Method(str, Enum):
    SYMLINK = "symlink"
    COPY = "copy"


@dataclass
class SyncPlan:
    source_harness: Harness
    target_harnesses: tuple[Harness, ...]
    rtype: ResourceType
    scope: Scope
    method: Method
    dry_run: bool
    force: bool

    def to_log(self) -> fs.OperationLog:
        return fs.OperationLog()


def _method_for(rtype: ResourceType, cfg: AgentsConfig, override: str | None) -> Method:
    if override is not None:
        return Method(override)
    return Method(cfg.method_for(rtype))


def resolve_source(cfg: AgentsConfig, override: str | None = None) -> Harness:
    return get(override or cfg.source)


def resolve_targets(
    cfg: AgentsConfig,
    overrides: tuple[str, ...] | None = None,
    *,
    exclude_source: Harness | None = None,
) -> tuple[Harness, ...]:
    """Resolve the list of target harnesses.

    If ``overrides`` is empty and config targets is empty, default to
    "all harnesses that support the current source's resource set".
    """
    if overrides:
        ids = overrides
    elif cfg.targets:
        ids = cfg.targets
    else:
        ids = tuple(h.id for h in ALL_HARNESSES)

    harnesses: list[Harness] = []
    for hid in ids:
        h = get(hid)
        if exclude_source is not None and h.id == exclude_source.id:
            continue
        harnesses.append(h)
    return tuple(harnesses)


def _target_path_for(
    source: DiscoveredResource,
    target: Harness,
    ctx: ScopeContext,
) -> Path | None:
    """Translate a source resource to its target-harness path."""
    if not target.supports(source.rtype):
        return None

    # Guidelines are location-sensitive (nested). We anchor on the
    # directory of the source file and use the *same* relative offset
    # under the project root for the target.
    if source.rtype is ResourceType.GUIDELINE:
        spec = target.spec(source.rtype)
        assert spec is not None
        if source.scope is Scope.PROJECT:
            if ctx.project_root is None:
                return None
            # Source sits in <project>/<sub>/<filename>
            try:
                rel_dir = source.path.parent.relative_to(ctx.project_root)
            except ValueError:
                return None
            return ctx.project_root / rel_dir / spec.project_path
        if spec.global_path is None:
            return None
        return Path(spec.global_path).expanduser()

    # Everything else uses the standard rendered path.
    rendered = render_paths(
        target, source.rtype, source.name, source.scope, ctx, include_aliases=False
    )
    if not rendered:
        return None
    tgt_spec = target.spec(source.rtype)
    path = Path(rendered[0].path)
    # For DIRECTORY resources (skills), the rendered path points at
    # ``<name>/SKILL.md`` but we want to link/copy the directory itself.
    if tgt_spec is not None and tgt_spec.kind is Kind.DIRECTORY:
        path = path.parent
    return path


def _format_mismatch(source: DiscoveredResource, target: Harness) -> str | None:
    """Return a reason string if the target uses a different file format."""
    src_spec = get(source.harness_id).spec(source.rtype)
    tgt_spec = target.spec(source.rtype)
    if src_spec is None or tgt_spec is None:
        return None
    if src_spec.format is not tgt_spec.format and tgt_spec.kind is Kind.FILE:
        return (
            f"{target.id} expects {tgt_spec.format.value} but source is "
            f"{src_spec.format.value} — skipping (convert manually)"
        )
    return None


def _sync_one(
    source_res: DiscoveredResource,
    target: Harness,
    method: Method,
    ctx: ScopeContext,
    *,
    dry_run: bool,
    force: bool,
    confirm: fs.ConfirmFn,
    log: fs.OperationLog,
) -> None:
    mismatch = _format_mismatch(source_res, target)
    if mismatch is not None:
        log.record(
            fs.Operation(
                fs.Op.SKIPPED,
                method.value,
                Path("<skipped>"),
                source_res.path,
                reason=mismatch,
            )
        )
        return

    target_path = _target_path_for(source_res, target, ctx)
    if target_path is None:
        return  # Target doesn't support this resource/scope combo.

    # Don't sync a path to itself.
    try:
        if source_res.path.resolve(strict=False) == target_path.resolve(strict=False):
            log.record(
                fs.Operation(
                    fs.Op.SKIPPED,
                    method.value,
                    target_path,
                    source_res.path,
                    reason="same path",
                )
            )
            return
    except OSError:
        pass

    spec = target.spec(source_res.rtype)
    assert spec is not None

    # For DIRECTORY resources (skills), the source is the directory itself.
    src = source_res.path

    if method is Method.SYMLINK:
        fs.symlink(src, target_path, dry_run=dry_run, force=force, confirm=confirm, log=log)
    else:
        fs.copy(src, target_path, dry_run=dry_run, force=force, confirm=confirm, log=log)


def run_sync(
    rtype: ResourceType,
    scope: Scope,
    ctx: ScopeContext,
    cfg: AgentsConfig,
    *,
    source_override: str | None = None,
    target_overrides: tuple[str, ...] = (),
    method_override: str | None = None,
    dry_run: bool = False,
    force: bool = False,
    names: tuple[str, ...] = (),
    confirm: fs.ConfirmFn = fs.default_confirm,
) -> fs.OperationLog:
    """Sync all resources of ``rtype`` from the source harness to targets."""
    source = resolve_source(cfg, source_override)
    if not source.supports(rtype):
        log = fs.OperationLog()
        log.record(
            fs.Operation(
                fs.Op.SKIPPED,
                "sync",
                Path("<none>"),
                reason=f"{source.display_name} does not support {rtype.value}",
            )
        )
        return log

    targets = resolve_targets(cfg, target_overrides, exclude_source=source)
    targets = tuple(t for t in targets if t.supports(rtype))

    method = _method_for(rtype, cfg, method_override)
    log = fs.OperationLog()

    source_resources = discover(rtype, scope, ctx, harnesses=(source,))

    # Filter out resources that live in alias locations for the source
    # (we only sync "real" sources). Also filter by names if specified.
    selected: list[DiscoveredResource] = []
    for r in source_resources:
        if r.is_alias_location:
            continue
        if names and r.name not in names and r.path.name not in names:
            continue
        selected.append(r)

    if not selected:
        log.record(
            fs.Operation(
                fs.Op.SKIPPED,
                "sync",
                Path("<none>"),
                reason=f"no {rtype.value} resources found in {source.display_name}",
            )
        )
        return log

    for sr in selected:
        for t in targets:
            _sync_one(
                sr,
                t,
                method,
                ctx,
                dry_run=dry_run,
                force=force,
                confirm=confirm,
                log=log,
            )

    return log
