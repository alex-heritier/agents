"""Tests for browser.py — pure helpers and Textual TUI smoke tests.

Pure helper tests are synchronous and have no Textual dependency.
Textual pilot tests use ``asyncio.run`` so no extra pytest plugins are needed.
"""

from __future__ import annotations

import asyncio
from pathlib import Path

from agents_cli.browser import (
    EditPicker,
    ShowBrowser,
    build_preview,
    filter_resources,
    format_resource_row,
)
from agents_cli.discovery import DiscoveredResource
from agents_cli.harnesses import ResourceType
from agents_cli.paths import Scope, ScopeContext

# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


def _make_res(
    name: str,
    content: str,
    tmp_path: Path,
    *,
    is_symlink: bool = False,
    is_alias_location: bool = False,
    scope: Scope = Scope.PROJECT,
    harness_id: str = "opencode",
) -> DiscoveredResource:
    path = tmp_path / f"{name}.md"
    path.write_text(content)
    return DiscoveredResource(
        harness_id=harness_id,
        rtype=ResourceType.COMMAND,
        scope=scope,
        name=name,
        path=path,
        entry=path,
        is_symlink=is_symlink,
        symlink_target=None,
        size=len(content.encode()),
        is_alias_location=is_alias_location,
    )


# ---------------------------------------------------------------------------
# filter_resources
# ---------------------------------------------------------------------------


def test_filter_resources_empty_query_matches_all(
    project: Path, fake_home: Path, tmp_path: Path
) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [_make_res("alpha", "a", tmp_path), _make_res("beta", "b", tmp_path)]
    assert filter_resources(resources, "", ctx) == [0, 1]


def test_filter_resources_matches_name(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [_make_res("alpha", "a", tmp_path), _make_res("beta", "b", tmp_path)]
    assert filter_resources(resources, "beta", ctx) == [1]


def test_filter_resources_case_insensitive(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [_make_res("AlPhA", "a", tmp_path)]
    assert filter_resources(resources, "alph", ctx) == [0]


def test_filter_resources_matches_harness(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [
        _make_res("foo", "a", tmp_path, harness_id="opencode"),
        _make_res("bar", "b", tmp_path, harness_id="claude"),
    ]
    assert filter_resources(resources, "claude", ctx) == [1]


def test_filter_resources_matches_scope(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [
        _make_res("x", "a", tmp_path, scope=Scope.PROJECT),
        _make_res("y", "b", tmp_path, scope=Scope.GLOBAL),
    ]
    assert filter_resources(resources, "global", ctx) == [1]


def test_filter_resources_no_match(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [_make_res("alpha", "a", tmp_path)]
    assert filter_resources(resources, "zzznomatch", ctx) == []


# ---------------------------------------------------------------------------
# build_preview
# ---------------------------------------------------------------------------


def test_build_preview_contains_content(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    res = _make_res("hello", "hello world content", tmp_path)
    preview = build_preview(res, ctx)
    assert "hello world content" in preview


def test_build_preview_contains_metadata(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    res = _make_res("hello", "body", tmp_path, scope=Scope.GLOBAL)
    preview = build_preview(res, ctx)
    assert "global" in preview
    assert "─" in preview  # separator line present


def test_build_preview_oserror(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    res = _make_res("x", "data", tmp_path)
    # Remove the file so reading fails.
    res.path.unlink()
    preview = build_preview(res, ctx)
    assert "Read error" in preview or "error" in preview.lower()


# ---------------------------------------------------------------------------
# format_resource_row
# ---------------------------------------------------------------------------


def test_format_resource_row_normal(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    res = _make_res("deploy", "content", tmp_path)
    row = format_resource_row(res, ctx, 3)
    assert row[0] == "3"
    assert "deploy" in row[1]
    assert row[2] in ("project", "global")


def test_format_resource_row_symlink_prefix(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    res = _make_res("linked", "x", tmp_path, is_symlink=True)
    row = format_resource_row(res, ctx, 1)
    assert row[1].startswith("↪ ")


def test_format_resource_row_alias_prefix(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    res = _make_res("aliased", "x", tmp_path, is_alias_location=True)
    row = format_resource_row(res, ctx, 2)
    assert row[1].startswith("~ ")


def test_format_resource_row_size(project: Path, fake_home: Path, tmp_path: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    res = _make_res("big", "x" * 1500, tmp_path)
    row = format_resource_row(res, ctx, 1)
    assert "K" in row[3]  # size column shows kilobytes


# ---------------------------------------------------------------------------
# Textual pilot — ShowBrowser
# ---------------------------------------------------------------------------


def test_show_browser_quit_returns_0(project: Path, fake_home: Path, tmp_path: Path) -> None:
    """Pressing q exits ShowBrowser with code 0."""
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [
        _make_res("alpha", "content A", tmp_path),
        _make_res("beta", "content B", tmp_path),
    ]

    async def _run() -> int | None:
        app = ShowBrowser(resources=resources, ctx=ctx, title="Test Browser")
        async with app.run_test(headless=True) as pilot:
            await pilot.press("q")
        return app.return_value

    code = asyncio.run(_run())
    assert code == 0


def test_show_browser_navigate_j_k(project: Path, fake_home: Path, tmp_path: Path) -> None:
    """j/k keys change the cursor row in the list table."""
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [
        _make_res("alpha", "A", tmp_path),
        _make_res("beta", "B", tmp_path),
        _make_res("gamma", "C", tmp_path),
    ]

    async def _run() -> tuple[int, int]:
        app = ShowBrowser(resources=resources, ctx=ctx, title="Nav Test")
        async with app.run_test(headless=True) as pilot:
            from textual.widgets import DataTable

            table = app.query_one("#list-table", DataTable)
            initial_row = table.cursor_row
            await pilot.press("j")
            after_j = table.cursor_row
            await pilot.press("k")
            after_k = table.cursor_row
            await pilot.press("q")
        return (after_j - initial_row, after_k)

    delta_j, after_k = asyncio.run(_run())
    assert delta_j == 1  # j moved down
    assert after_k == 0  # k moved back up to 0


def test_show_browser_filter_narrows_list(project: Path, fake_home: Path, tmp_path: Path) -> None:
    """Typing into the filter input reduces visible rows."""
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [
        _make_res("alpha", "A", tmp_path),
        _make_res("beta", "B", tmp_path),
        _make_res("gamma", "C", tmp_path),
    ]

    async def _run() -> int:
        app = ShowBrowser(resources=resources, ctx=ctx, title="Filter Test")
        async with app.run_test(headless=True) as pilot:
            from textual.widgets import DataTable

            await pilot.press("/")
            await pilot.press("b", "e", "t")
            await pilot.pause()
            table = app.query_one("#list-table", DataTable)
            count = table.row_count
            await pilot.press("q")
        return count

    count = asyncio.run(_run())
    assert count == 1


# ---------------------------------------------------------------------------
# Textual pilot — EditPicker
# ---------------------------------------------------------------------------


def test_edit_picker_cancel_returns_none(project: Path, fake_home: Path, tmp_path: Path) -> None:
    """Pressing q exits EditPicker with None."""
    ctx = ScopeContext(project_root=project, home=fake_home)
    resources = [_make_res("alpha", "A", tmp_path), _make_res("beta", "B", tmp_path)]

    async def _run() -> Path | None:
        app = EditPicker(resources=resources, ctx=ctx, title="Pick Test")
        async with app.run_test(headless=True) as pilot:
            await pilot.press("q")
        return app.return_value

    result = asyncio.run(_run())
    assert result is None


def test_edit_picker_enter_returns_path(project: Path, fake_home: Path, tmp_path: Path) -> None:
    """Pressing Enter on the first row returns that resource's entry path."""
    ctx = ScopeContext(project_root=project, home=fake_home)
    res_a = _make_res("alpha", "A", tmp_path)
    res_b = _make_res("beta", "B", tmp_path)
    resources = [res_a, res_b]

    async def _run() -> Path | None:
        app = EditPicker(resources=resources, ctx=ctx, title="Select Test")
        async with app.run_test(headless=True) as pilot:
            await pilot.press("enter")
        return app.return_value

    result = asyncio.run(_run())
    assert result is not None
    assert result == res_a.entry


def test_edit_picker_navigate_then_select(project: Path, fake_home: Path, tmp_path: Path) -> None:
    """Navigate to the second row with j then Enter returns the second resource."""
    ctx = ScopeContext(project_root=project, home=fake_home)
    res_a = _make_res("alpha", "A", tmp_path)
    res_b = _make_res("beta", "B", tmp_path)
    resources = [res_a, res_b]

    async def _run() -> Path | None:
        app = EditPicker(resources=resources, ctx=ctx, title="Nav Select Test")
        async with app.run_test(headless=True) as pilot:
            await pilot.press("j")
            await pilot.press("enter")
        return app.return_value

    result = asyncio.run(_run())
    assert result == res_b.entry
