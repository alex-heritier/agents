"""Tests for the filesystem-ops layer."""

from __future__ import annotations

from pathlib import Path

from agents_cli import fs


def _noconfirm(_msg: str) -> bool:
    return False


def test_symlink_creates_relative_link(tmp_path: Path) -> None:
    src = tmp_path / "real.md"
    src.write_text("hello")
    tgt = tmp_path / "sub" / "link.md"
    log = fs.OperationLog()
    result = fs.symlink(src, tgt, log=log)
    assert result is fs.Op.CREATED
    assert tgt.is_symlink()
    # Link points to a relative path, not absolute.
    assert not str(tgt.readlink()).startswith("/")
    assert tgt.read_text() == "hello"
    assert log.created == 1


def test_symlink_idempotent(tmp_path: Path) -> None:
    src = tmp_path / "real.md"
    src.write_text("hi")
    tgt = tmp_path / "link.md"
    fs.symlink(src, tgt)
    log = fs.OperationLog()
    result = fs.symlink(src, tgt, log=log)
    assert result is fs.Op.SKIPPED
    assert log.skipped == 1


def test_symlink_declines_to_clobber_without_force(tmp_path: Path) -> None:
    src = tmp_path / "real.md"
    src.write_text("new")
    tgt = tmp_path / "existing.md"
    tgt.write_text("original")
    log = fs.OperationLog()
    result = fs.symlink(src, tgt, confirm=_noconfirm, log=log)
    assert result is fs.Op.SKIPPED
    assert tgt.read_text() == "original"
    assert not tgt.is_symlink()


def test_symlink_force_clobbers_file(tmp_path: Path) -> None:
    src = tmp_path / "real.md"
    src.write_text("new")
    tgt = tmp_path / "existing.md"
    tgt.write_text("original")
    result = fs.symlink(src, tgt, force=True)
    assert result is fs.Op.CREATED
    assert tgt.is_symlink()
    assert tgt.read_text() == "new"


def test_symlink_dry_run_does_not_touch_disk(tmp_path: Path) -> None:
    src = tmp_path / "real.md"
    src.write_text("x")
    tgt = tmp_path / "out.md"
    log = fs.OperationLog()
    result = fs.symlink(src, tgt, dry_run=True, log=log)
    assert result is fs.Op.CREATED
    assert not tgt.exists()
    assert log.created == 1


def test_copy_creates_file(tmp_path: Path) -> None:
    src = tmp_path / "a.md"
    src.write_text("content")
    tgt = tmp_path / "b.md"
    result = fs.copy(src, tgt)
    assert result is fs.Op.CREATED
    assert tgt.read_text() == "content"
    assert not tgt.is_symlink()


def test_copy_skips_identical_content(tmp_path: Path) -> None:
    src = tmp_path / "a.md"
    src.write_text("same")
    tgt = tmp_path / "b.md"
    tgt.write_text("same")
    result = fs.copy(src, tgt)
    assert result is fs.Op.SKIPPED


def test_delete_removes_file(tmp_path: Path) -> None:
    p = tmp_path / "f.md"
    p.write_text("x")
    log = fs.OperationLog()
    fs.remove(p, force=True, log=log)
    assert not p.exists()
    assert log.deleted == 1
