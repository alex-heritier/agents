"""Filesystem operations with safety, dry-run, and verbose logging.

Every mutation goes through this module so we can centralize safety
checks (no accidental overwrite, no symlink loops) and produce a clean
``OperationLog`` for summaries.
"""

from __future__ import annotations

import filecmp
import os
import shutil
from collections.abc import Callable
from dataclasses import dataclass, field
from enum import Enum
from pathlib import Path


class Op(str, Enum):
    CREATED = "created"
    UPDATED = "updated"
    SKIPPED = "skipped"
    DELETED = "deleted"
    FAILED = "failed"


@dataclass
class Operation:
    op: Op
    kind: str  # "symlink" | "copy" | "delete" | "mkdir"
    target: Path
    source: Path | None = None
    reason: str = ""

    def format(self) -> str:
        # Collapse redundant verb/kind pairs: e.g. "deleted delete" → "deleted",
        # "skipped write" → "skipped write" stays, but "created write" → "wrote".
        verb_map = {
            (Op.DELETED, "delete"): "deleted",
            (Op.CREATED, "write"): "wrote",
            (Op.UPDATED, "write"): "updated",
            (Op.CREATED, "mkdir"): "mkdir",
        }
        action = verb_map.get((self.op, self.kind), f"{self.op.value} {self.kind}")
        if self.source is not None:
            base = f"{action}: {self.target} ← {self.source}"
        else:
            base = f"{action}: {self.target}"
        if self.reason:
            return f"{base}  ({self.reason})"
        return base


@dataclass
class OperationLog:
    operations: list[Operation] = field(default_factory=list)

    def record(self, op: Operation) -> None:
        self.operations.append(op)

    def extend(self, other: OperationLog) -> None:
        self.operations.extend(other.operations)

    def count(self, op: Op) -> int:
        return sum(1 for o in self.operations if o.op is op)

    @property
    def created(self) -> int:
        return self.count(Op.CREATED)

    @property
    def updated(self) -> int:
        return self.count(Op.UPDATED)

    @property
    def skipped(self) -> int:
        return self.count(Op.SKIPPED)

    @property
    def deleted(self) -> int:
        return self.count(Op.DELETED)

    @property
    def failed(self) -> int:
        return self.count(Op.FAILED)


ConfirmFn = Callable[[str], bool]


def default_confirm(_: str) -> bool:
    return False


def is_symlink(p: Path) -> bool:
    try:
        return p.is_symlink()
    except OSError:
        return False


def lexists(p: Path) -> bool:
    """True if the path exists as a file/dir/symlink (including broken symlinks)."""
    return p.exists() or p.is_symlink()


def read_symlink(p: Path) -> Path | None:
    try:
        return Path(os.readlink(p))
    except OSError:
        return None


def trees_equal(src: Path, dst: Path) -> bool:
    """Compare two files or two directories by contents (shallow cmp for files)."""
    if src.is_dir() and dst.is_dir():
        cmp = filecmp.dircmp(src, dst)
        return (
            not cmp.left_only
            and not cmp.right_only
            and not cmp.diff_files
            and all(trees_equal(src / sub, dst / sub) for sub in cmp.common_dirs)
        )
    if src.is_file() and dst.is_file():
        return filecmp.cmp(src, dst, shallow=False)
    return False


def _relative_symlink_target(link_location: Path, source: Path) -> Path:
    """Compute a stable relative path from the *directory containing* the link
    to the source. Falls back to absolute when the two live on different roots.
    """
    try:
        return Path(os.path.relpath(source, start=link_location.parent))
    except ValueError:
        return source


def ensure_parent(p: Path, *, dry_run: bool = False, log: OperationLog | None = None) -> None:
    if p.parent.exists():
        return
    if dry_run:
        if log is not None:
            log.record(Operation(Op.CREATED, "mkdir", p.parent, reason="dry-run"))
        return
    p.parent.mkdir(parents=True, exist_ok=True)
    if log is not None:
        log.record(Operation(Op.CREATED, "mkdir", p.parent))


def symlink(
    source: Path,
    target: Path,
    *,
    dry_run: bool = False,
    force: bool = False,
    confirm: ConfirmFn = default_confirm,
    log: OperationLog | None = None,
) -> Op:
    """Create or update a symlink from ``target`` -> ``source`` (relative).

    - If ``target`` exists and is already a correct symlink, SKIPPED.
    - If ``target`` exists and is a matching file/dir, SKIPPED (no-op).
    - If ``target`` exists and is different, ask for confirmation (or use
      ``force``) before overwriting.
    """
    source = Path(source)
    target = Path(target)
    log = log if log is not None else OperationLog()
    link_value = _relative_symlink_target(target, source)

    if not source.exists() and not source.is_symlink():
        log.record(
            Operation(Op.FAILED, "symlink", target, source, reason=f"source missing: {source}")
        )
        return Op.FAILED

    if lexists(target):
        if is_symlink(target):
            existing = read_symlink(target)
            if existing is not None and Path(str(existing)) == link_value:
                log.record(
                    Operation(Op.SKIPPED, "symlink", target, source, reason="already correct")
                )
                return Op.SKIPPED
            # Different symlink target.
            if not force and not confirm(f"Overwrite symlink {target}?"):
                log.record(Operation(Op.SKIPPED, "symlink", target, source, reason="user declined"))
                return Op.SKIPPED
            if dry_run:
                log.record(Operation(Op.UPDATED, "symlink", target, source, reason="dry-run"))
                return Op.UPDATED
            target.unlink()
        else:
            # Real file/directory at target.
            if target.is_file() and source.is_file() and trees_equal(source, target):
                log.record(
                    Operation(Op.SKIPPED, "symlink", target, source, reason="identical content")
                )
                return Op.SKIPPED
            if not force and not confirm(f"Replace {target} with symlink to {source}?"):
                log.record(Operation(Op.SKIPPED, "symlink", target, source, reason="user declined"))
                return Op.SKIPPED
            if dry_run:
                log.record(Operation(Op.UPDATED, "symlink", target, source, reason="dry-run"))
                return Op.UPDATED
            if target.is_dir() and not target.is_symlink():
                shutil.rmtree(target)
            else:
                target.unlink()

    if dry_run:
        log.record(Operation(Op.CREATED, "symlink", target, source, reason="dry-run"))
        return Op.CREATED

    ensure_parent(target)
    try:
        os.symlink(link_value, target)
    except OSError as e:
        log.record(Operation(Op.FAILED, "symlink", target, source, reason=str(e)))
        return Op.FAILED

    log.record(Operation(Op.CREATED, "symlink", target, source))
    return Op.CREATED


def copy(
    source: Path,
    target: Path,
    *,
    dry_run: bool = False,
    force: bool = False,
    confirm: ConfirmFn = default_confirm,
    log: OperationLog | None = None,
) -> Op:
    """Copy a file or directory from ``source`` to ``target``.

    Directory copies merge on top of existing contents; identical entries
    are skipped individually.
    """
    source = Path(source)
    target = Path(target)
    log = log if log is not None else OperationLog()

    if not source.exists():
        log.record(Operation(Op.FAILED, "copy", target, source, reason=f"source missing: {source}"))
        return Op.FAILED

    if lexists(target):
        if is_symlink(target):
            if not force and not confirm(f"Replace symlink {target} with a real copy?"):
                log.record(Operation(Op.SKIPPED, "copy", target, source, reason="user declined"))
                return Op.SKIPPED
            if dry_run:
                log.record(Operation(Op.UPDATED, "copy", target, source, reason="dry-run"))
                return Op.UPDATED
            target.unlink()
        elif trees_equal(source, target):
            log.record(Operation(Op.SKIPPED, "copy", target, source, reason="identical content"))
            return Op.SKIPPED
        else:
            if not force and not confirm(f"Overwrite {target} with copy of {source}?"):
                log.record(Operation(Op.SKIPPED, "copy", target, source, reason="user declined"))
                return Op.SKIPPED
            if dry_run:
                log.record(Operation(Op.UPDATED, "copy", target, source, reason="dry-run"))
                return Op.UPDATED
            if target.is_dir():
                shutil.rmtree(target)
            else:
                target.unlink()

    if dry_run:
        log.record(Operation(Op.CREATED, "copy", target, source, reason="dry-run"))
        return Op.CREATED

    ensure_parent(target)
    try:
        if source.is_dir():
            shutil.copytree(source, target, symlinks=True)
        else:
            shutil.copy2(source, target, follow_symlinks=True)
    except OSError as e:
        log.record(Operation(Op.FAILED, "copy", target, source, reason=str(e)))
        return Op.FAILED

    log.record(Operation(Op.CREATED, "copy", target, source))
    return Op.CREATED


def remove(
    target: Path,
    *,
    dry_run: bool = False,
    force: bool = False,
    confirm: ConfirmFn = default_confirm,
    log: OperationLog | None = None,
) -> Op:
    """Delete a file, directory, or symlink. Prompts unless ``force``."""
    log = log if log is not None else OperationLog()
    if not lexists(target):
        log.record(Operation(Op.SKIPPED, "delete", target, reason="not present"))
        return Op.SKIPPED

    if not force and not confirm(f"Delete {target}?"):
        log.record(Operation(Op.SKIPPED, "delete", target, reason="user declined"))
        return Op.SKIPPED

    if dry_run:
        log.record(Operation(Op.DELETED, "delete", target, reason="dry-run"))
        return Op.DELETED

    try:
        if target.is_dir() and not target.is_symlink():
            shutil.rmtree(target)
        else:
            target.unlink()
    except OSError as e:
        log.record(Operation(Op.FAILED, "delete", target, reason=str(e)))
        return Op.FAILED

    log.record(Operation(Op.DELETED, "delete", target))
    return Op.DELETED


def write_text(
    path: Path,
    content: str,
    *,
    dry_run: bool = False,
    force: bool = False,
    confirm: ConfirmFn = default_confirm,
    log: OperationLog | None = None,
) -> Op:
    """Create a new text file. Prompts before overwriting."""
    log = log if log is not None else OperationLog()
    if lexists(path):
        try:
            existing = path.read_text()
        except OSError:
            existing = None
        if existing == content:
            log.record(Operation(Op.SKIPPED, "write", path, reason="identical content"))
            return Op.SKIPPED
        if not force and not confirm(f"Overwrite {path}?"):
            log.record(Operation(Op.SKIPPED, "write", path, reason="user declined"))
            return Op.SKIPPED
        if dry_run:
            log.record(Operation(Op.UPDATED, "write", path, reason="dry-run"))
            return Op.UPDATED
    elif dry_run:
        log.record(Operation(Op.CREATED, "write", path, reason="dry-run"))
        return Op.CREATED

    ensure_parent(path)
    try:
        path.write_text(content)
    except OSError as e:
        log.record(Operation(Op.FAILED, "write", path, reason=str(e)))
        return Op.FAILED

    log.record(Operation(Op.CREATED, "write", path))
    return Op.CREATED
