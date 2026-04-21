"""Discovery: find existing resources on disk.

For each ``(harness, resource_type, scope)`` tuple we know the canonical
location template. Discovery enumerates what actually exists, across the
canonical path *and* any aliases.

Special cases:

- **Guidelines** are nested: AGENTS.md can live in subdirectories too.
  At project scope we walk the tree from the project root, skipping
  noisy build/vendor directories.
- **Skills / commands / subagents** live in a single directory per
  scope, enumerated directly.
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from agents_cli.fs import is_symlink, read_symlink
from agents_cli.harnesses import ALL_HARNESSES, Harness, Kind, ResourceType
from agents_cli.paths import Scope, ScopeContext, render_paths

# Directories we always skip when walking project trees.
IGNORE_DIRS: frozenset[str] = frozenset(
    {
        ".git",
        "node_modules",
        "dist",
        "build",
        ".next",
        ".turbo",
        ".venv",
        "venv",
        "__pycache__",
        ".pytest_cache",
        ".mypy_cache",
        ".ruff_cache",
        "target",
        "vendor",
        ".gradle",
        ".idea",
        ".vscode",
        ".cache",
    }
)


@dataclass(frozen=True)
class DiscoveredResource:
    """One concrete resource found on disk."""

    harness_id: str
    rtype: ResourceType
    scope: Scope
    name: str  # slug (skill/command/subagent) or filename for guidelines
    path: Path  # file for FILE resources, directory for DIRECTORY resources
    entry: Path  # canonical file to show/edit (SKILL.md for skills)
    is_symlink: bool
    symlink_target: Path | None
    size: int
    is_alias_location: bool

    def display_name(self) -> str:
        if self.rtype is ResourceType.GUIDELINE:
            return self.path.name
        return self.name

    def container(self) -> Path:
        """Directory that conceptually owns this resource.

        For guidelines, this is the directory the guideline file lives in
        (project root, a package subdir, or ``~/.<harness>``). For other
        resource types we return the resource's own path's parent — this is
        primarily used by the status matcher for guidelines.
        """
        if self.rtype is ResourceType.GUIDELINE:
            return self.path.parent
        return self.path.parent


def _is_real_dir(path: Path) -> bool:
    """``path.is_dir()`` that does not follow symlinks (3.10/3.11 compatible)."""
    try:
        st = path.lstat()
    except OSError:
        return False
    import stat as _stat

    return _stat.S_ISDIR(st.st_mode)


def _iter_subdirs(root: Path):
    try:
        for entry in root.iterdir():
            if _is_real_dir(entry) and entry.name not in IGNORE_DIRS:
                yield entry
    except (OSError, PermissionError):
        return


def _walk_project(root: Path):
    """Breadth-first walk, skipping IGNORE_DIRS."""
    yield root
    queue = [root]
    while queue:
        current = queue.pop(0)
        for child in _iter_subdirs(current):
            yield child
            queue.append(child)


def _stat_info(path: Path) -> tuple[bool, Path | None, int]:
    sym = is_symlink(path)
    tgt = read_symlink(path) if sym else None
    try:
        if path.is_file():
            size = path.stat().st_size
        elif path.is_dir():
            total = 0
            for f in path.rglob("*"):
                try:
                    st = f.lstat()
                except OSError:
                    continue
                import stat as _stat

                if _stat.S_ISREG(st.st_mode):
                    total += st.st_size
            size = total
        else:
            size = 0
    except OSError:
        size = 0
    return sym, tgt, size


def _discover_guideline_project(harness: Harness, ctx: ScopeContext) -> list[DiscoveredResource]:
    if ctx.project_root is None or harness.guideline is None:
        return []
    spec = harness.guideline
    candidate_names = (spec.project_path, *spec.project_aliases)

    results: list[DiscoveredResource] = []
    for directory in _walk_project(ctx.project_root):
        for rel in candidate_names:
            # Only consider names that are files in a single-dir-name form.
            p = directory / rel
            if p.is_file() or p.is_symlink():
                sym, tgt, size = _stat_info(p)
                results.append(
                    DiscoveredResource(
                        harness_id=harness.id,
                        rtype=ResourceType.GUIDELINE,
                        scope=Scope.PROJECT,
                        name=rel,
                        path=p,
                        entry=p,
                        is_symlink=sym,
                        symlink_target=tgt,
                        size=size,
                        is_alias_location=(rel != spec.project_path),
                    )
                )
    return results


def _discover_guideline_global(harness: Harness, ctx: ScopeContext) -> list[DiscoveredResource]:
    spec = harness.guideline
    if spec is None or spec.global_path is None:
        return []
    out: list[DiscoveredResource] = []
    candidates: list[tuple[str, bool]] = [(spec.global_path, False)]
    candidates.extend((alias, True) for alias in spec.global_aliases)
    for candidate, is_alias in candidates:
        p = Path(candidate).expanduser()
        if p.is_file() or p.is_symlink():
            sym, tgt, size = _stat_info(p)
            out.append(
                DiscoveredResource(
                    harness_id=harness.id,
                    rtype=ResourceType.GUIDELINE,
                    scope=Scope.GLOBAL,
                    name=p.name,
                    path=p,
                    entry=p,
                    is_symlink=sym,
                    symlink_target=tgt,
                    size=size,
                    is_alias_location=is_alias,
                )
            )
    return out


def _dir_containing_resources(
    template: str, scope_root: Path | None, name_placeholder: str
) -> Path | None:
    """Given a template like ``.opencode/skills/{name}/SKILL.md``, return the
    directory containing the named items (``.opencode/skills/``), anchored to
    ``scope_root`` (or expanded from ``~`` when ``scope_root`` is ``None``).
    """
    if "{name}" not in template:
        return None
    prefix = template.split("{name}", 1)[0].rstrip("/")
    if not prefix:
        return None
    if scope_root is not None:
        return scope_root / prefix
    return Path(prefix).expanduser()


def _discover_file_resources(
    harness: Harness,
    rtype: ResourceType,
    scope: Scope,
    ctx: ScopeContext,
) -> list[DiscoveredResource]:
    spec = harness.spec(rtype)
    if spec is None:
        return []

    if scope is Scope.PROJECT:
        if ctx.project_root is None:
            return []
        candidates = [spec.project_path, *spec.project_aliases]
        scope_root: Path | None = ctx.project_root
    else:
        if spec.global_path is None:
            return []
        candidates = [spec.global_path, *spec.global_aliases]
        scope_root = None

    results: list[DiscoveredResource] = []
    for i, template in enumerate(candidates):
        container = _dir_containing_resources(template, scope_root, "{name}")
        if container is None or not container.exists():
            continue

        if spec.kind is Kind.FILE:
            suffix = template.rsplit("{name}", 1)[-1]  # e.g. ".md"
            try:
                entries = sorted(container.rglob("*" + suffix))
            except OSError:
                continue
            for entry in entries:
                if not entry.is_file() and not entry.is_symlink():
                    continue
                rel = entry.relative_to(container)
                name = str(rel).removesuffix(suffix).replace("\\", "/")
                if not name:
                    continue
                sym, tgt, size = _stat_info(entry)
                results.append(
                    DiscoveredResource(
                        harness_id=harness.id,
                        rtype=rtype,
                        scope=scope,
                        name=name,
                        path=entry,
                        entry=entry,
                        is_symlink=sym,
                        symlink_target=tgt,
                        size=size,
                        is_alias_location=(i != 0),
                    )
                )
        else:
            # DIRECTORY: each immediate subdirectory that contains the entry file.
            try:
                entries = sorted(container.iterdir())
            except OSError:
                continue
            for entry_dir in entries:
                # Accept both real directories and symlinks that resolve to
                # directories (e.g. synced skill dirs).
                if not entry_dir.is_dir():
                    continue
                entry_file = entry_dir / (spec.entry_file or "SKILL.md")
                if not (entry_file.is_file() or entry_file.is_symlink()):
                    continue
                sym, tgt, size = _stat_info(entry_dir)
                results.append(
                    DiscoveredResource(
                        harness_id=harness.id,
                        rtype=rtype,
                        scope=scope,
                        name=entry_dir.name,
                        path=entry_dir,
                        entry=entry_file,
                        is_symlink=sym,
                        symlink_target=tgt,
                        size=size,
                        is_alias_location=(i != 0),
                    )
                )

    # De-duplicate: same (harness, type, scope, name) might match multiple
    # alias templates when they refer to the same on-disk location.
    seen: dict[tuple[str, str], DiscoveredResource] = {}
    for r in results:
        key = (r.name, str(r.path.resolve(strict=False)))
        if key not in seen or not r.is_alias_location:
            seen[key] = r
    return list(seen.values())


def discover(
    rtype: ResourceType,
    scope: Scope,
    ctx: ScopeContext,
    *,
    harnesses: tuple[Harness, ...] | None = None,
) -> list[DiscoveredResource]:
    hlist = harnesses or ALL_HARNESSES
    found: list[DiscoveredResource] = []
    for h in hlist:
        if not h.supports(rtype):
            continue
        if rtype is ResourceType.GUIDELINE:
            if scope is Scope.PROJECT:
                found.extend(_discover_guideline_project(h, ctx))
            else:
                found.extend(_discover_guideline_global(h, ctx))
        else:
            found.extend(_discover_file_resources(h, rtype, scope, ctx))
    return found


def find(
    harness: Harness,
    rtype: ResourceType,
    name: str,
    scope: Scope,
    ctx: ScopeContext,
) -> DiscoveredResource | None:
    """Locate a specific resource by name (e.g. skill 'git-release')."""
    for candidate in render_paths(harness, rtype, name, scope, ctx):
        p = Path(candidate.path)
        spec = harness.spec(rtype)
        if spec is None:
            return None
        if spec.kind is Kind.DIRECTORY:
            entry_file = p.parent / (spec.entry_file or "SKILL.md")
            # p is the entry file path, so parent is the skill dir.
            skill_dir = p.parent
            if skill_dir.is_dir() and entry_file.is_file():
                sym, tgt, size = _stat_info(skill_dir)
                return DiscoveredResource(
                    harness_id=harness.id,
                    rtype=rtype,
                    scope=scope,
                    name=name,
                    path=skill_dir,
                    entry=entry_file,
                    is_symlink=sym,
                    symlink_target=tgt,
                    size=size,
                    is_alias_location=candidate.is_alias,
                )
        else:
            if p.is_file() or p.is_symlink():
                sym, tgt, size = _stat_info(p)
                return DiscoveredResource(
                    harness_id=harness.id,
                    rtype=rtype,
                    scope=scope,
                    name=name,
                    path=p,
                    entry=p,
                    is_symlink=sym,
                    symlink_target=tgt,
                    size=size,
                    is_alias_location=candidate.is_alias,
                )
    return None
