"""Path resolution and scope detection.

A *project* is the git worktree containing the current directory. A
*global* scope is anchored at the user's home directory. We detect a
project root by walking up from the cwd looking for a ``.git`` entry
(either a directory or a file — worktrees use a file).
"""

from __future__ import annotations

import os
from dataclasses import dataclass
from enum import Enum
from pathlib import Path

from agents_cli.harnesses import Harness, Kind, ResolvedPath, ResourceSpec, ResourceType


class Scope(str, Enum):
    PROJECT = "project"
    GLOBAL = "global"


def find_project_root(start: Path | None = None) -> Path | None:
    """Return the directory containing ``.git`` by walking up from ``start``.

    Returns ``None`` if no git root is found (we then treat ``start`` as a
    plain working directory, which is fine for global-only commands).
    """
    here = (start or Path.cwd()).resolve()
    for parent in (here, *here.parents):
        if (parent / ".git").exists():
            return parent
    return None


def expand(path: str | Path) -> Path:
    """Expand ``~`` and environment variables, resolve to absolute path."""
    s = os.path.expandvars(str(path))
    return Path(s).expanduser()


@dataclass(frozen=True)
class ScopeContext:
    """Runtime context: which scope(s) are active, project root, home."""

    project_root: Path | None
    home: Path

    @property
    def in_project(self) -> bool:
        return self.project_root is not None


def detect(cwd: Path | None = None) -> ScopeContext:
    root = find_project_root(cwd)
    return ScopeContext(project_root=root, home=Path.home())


def render_paths(
    harness: Harness,
    rtype: ResourceType,
    name: str,
    scope: Scope,
    ctx: ScopeContext,
    *,
    include_aliases: bool = True,
) -> list[ResolvedPath]:
    """Produce concrete filesystem paths for one harness-resource-scope.

    Returns an empty list if the harness doesn't support the resource at
    the given scope. The primary (canonical) path comes first, then any
    aliases.
    """
    spec = harness.spec(rtype)
    if spec is None:
        return []

    out: list[ResolvedPath] = []
    if scope is Scope.PROJECT:
        if ctx.project_root is None:
            return []
        base = ctx.project_root
        primary = spec.project_path
        aliases = spec.project_aliases if include_aliases else ()
    else:
        if spec.global_path is None:
            return []
        base = None  # Global paths are already absolute.
        primary = spec.global_path
        aliases = spec.global_aliases if include_aliases else ()

    rendered = _render_one(primary, base, name)
    out.append(ResolvedPath(raw=primary, path=str(rendered), is_alias=False))
    for alias in aliases:
        rendered_alias = _render_one(alias, base, name)
        out.append(ResolvedPath(raw=alias, path=str(rendered_alias), is_alias=True))
    return out


def _render_one(template: str, base: Path | None, name: str) -> Path:
    filled = template.replace("{name}", name) if "{name}" in template else template
    if base is None:
        return expand(filled)
    return (base / filled).resolve() if Path(filled).is_absolute() else (base / filled)


def display_path(path: Path, ctx: ScopeContext) -> str:
    """Pretty path: relative to project root, or ``~/...``. Does NOT follow symlinks."""
    p = path if path.is_absolute() else path.absolute()
    # Absolve ``..`` segments without following symlinks.
    p = Path(os.path.normpath(str(p)))

    rel: str | None = None
    if ctx.project_root is not None:
        try:
            rel = str(p.relative_to(ctx.project_root))
            rel = "./" + rel if rel != "." else "./"
        except ValueError:
            rel = None
    if rel is None:
        try:
            h = str(p.relative_to(ctx.home))
            rel = "~/" + h
        except ValueError:
            rel = str(p)

    return rel


def target_is_directory(spec: ResourceSpec) -> bool:
    return spec.kind is Kind.DIRECTORY
