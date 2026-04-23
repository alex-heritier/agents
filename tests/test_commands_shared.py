"""Tests for interactive show/edit and the new ui helpers.

Command-level tests use ``typer.testing.CliRunner`` with ``mix_stderr=False``
so stdout and stderr are captured separately, matching real-world behaviour.

All tests are filesystem-hermetic (``tmp_path`` / ``project`` fixture) and
never touch ``$HOME`` or open a real editor.
"""

from __future__ import annotations

from pathlib import Path
from typing import Any
from unittest.mock import patch

import pytest
from typer.testing import CliRunner

from agents_cli import ui
from agents_cli.commands._shared import (
    _pick_resource_interactive,
    _resolve_scope_for_edit,
    _resolve_scope_for_show,
    build_group,
)
from agents_cli.config import AgentsConfig
from agents_cli.harnesses import ResourceType
from agents_cli.paths import Scope, ScopeContext

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

runner = CliRunner()


def _write(path: Path, content: str = "") -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)


def _skill_app():
    return build_group(name="skill", rtype=ResourceType.SKILL, help_text="Test skill group")


def _cmd_app():
    return build_group(name="command", rtype=ResourceType.COMMAND, help_text="Test command group")


def _default_cfg() -> AgentsConfig:
    return AgentsConfig(source="opencode")


# ---------------------------------------------------------------------------
# Scope resolution helpers
# ---------------------------------------------------------------------------


def test_resolve_scope_for_show_in_project(project: Path, fake_home: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    assert _resolve_scope_for_show(ctx, False, False) is Scope.PROJECT


def test_resolve_scope_for_show_global_flag(project: Path, fake_home: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    assert _resolve_scope_for_show(ctx, False, True) is Scope.GLOBAL


def test_resolve_scope_for_show_no_project(fake_home: Path) -> None:
    ctx = ScopeContext(project_root=None, home=fake_home)
    assert _resolve_scope_for_show(ctx, False, False) is Scope.GLOBAL


def test_resolve_scope_for_edit_global_flag(project: Path, fake_home: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    assert _resolve_scope_for_edit(ctx, False, True) is Scope.GLOBAL


def test_resolve_scope_for_edit_falls_back_when_no_project(fake_home: Path) -> None:
    ctx = ScopeContext(project_root=None, home=fake_home)
    # No project → falls back to GLOBAL even though global_only=False.
    assert _resolve_scope_for_edit(ctx, False, False) is Scope.GLOBAL


def test_resolve_scope_for_edit_project(project: Path, fake_home: Path) -> None:
    ctx = ScopeContext(project_root=project, home=fake_home)
    assert _resolve_scope_for_edit(ctx, False, False) is Scope.PROJECT


# ---------------------------------------------------------------------------
# show exact mode
# ---------------------------------------------------------------------------


def test_show_exact_mode_prints_content(project: Path, fake_home: Path) -> None:
    _write(
        project / ".opencode" / "commands" / "deploy.md",
        "# Deploy command\nRun the deploy script.",
    )
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.ui.view_text") as mock_view,
    ):
        result = runner.invoke(app, ["show", "deploy"])

    assert result.exit_code == 0
    mock_view.assert_called_once()
    title, _subtitle, content = mock_view.call_args.args
    assert "deploy" in title
    assert "Deploy command" in content


def test_show_exact_mode_not_found_exits_1(project: Path, fake_home: Path) -> None:
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
    ):
        result = runner.invoke(app, ["show", "nonexistent"])

    assert result.exit_code == 1


# ---------------------------------------------------------------------------
# edit exact mode
# ---------------------------------------------------------------------------


def test_edit_exact_mode_opens_editor(project: Path, fake_home: Path) -> None:
    _write(project / ".opencode" / "commands" / "build.md", "# Build")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared._open_in_editor", return_value=0) as mock_editor,
    ):
        result = runner.invoke(app, ["edit", "build"])

    assert result.exit_code == 0
    mock_editor.assert_called_once()
    edited_path: Path = mock_editor.call_args.args[0]
    assert edited_path.name == "build.md"


def test_edit_exact_mode_not_found_exits_1(project: Path, fake_home: Path) -> None:
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
    ):
        result = runner.invoke(app, ["edit", "ghost"])

    assert result.exit_code == 1


# ---------------------------------------------------------------------------
# Interactive show — non-TTY guard
# ---------------------------------------------------------------------------


def test_show_interactive_non_tty_exits_2(project: Path, fake_home: Path) -> None:
    _write(project / ".opencode" / "commands" / "build.md", "# Build")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    # CliRunner provides a non-TTY stdin by default.
    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
    ):
        result = runner.invoke(app, ["show"])

    assert result.exit_code == 2
    assert "TTY" in result.output or "TTY" in (result.stderr or "")


def test_edit_interactive_non_tty_exits_2(project: Path, fake_home: Path) -> None:
    _write(project / ".opencode" / "commands" / "build.md", "# Build")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
    ):
        result = runner.invoke(app, ["edit"])

    assert result.exit_code == 2


# ---------------------------------------------------------------------------
# Interactive show — no candidates
# ---------------------------------------------------------------------------


def test_show_interactive_no_candidates_exits_1(project: Path, fake_home: Path) -> None:
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared.sys") as mock_sys,
    ):
        mock_sys.stdin.isatty.return_value = True
        result = runner.invoke(app, ["show"])

    assert result.exit_code == 1


# ---------------------------------------------------------------------------
# Interactive show — one candidate auto-selects
# ---------------------------------------------------------------------------


def test_show_interactive_one_candidate_auto_selects(project: Path, fake_home: Path) -> None:
    _write(project / ".opencode" / "commands" / "greet.md", "# Greet\nHello world.")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared.sys") as mock_sys,
        patch("agents_cli.ui.view_text") as mock_view,
    ):
        mock_sys.stdin.isatty.return_value = True
        result = runner.invoke(app, ["show"])

    # No prompt needed; auto-selected; view shown.
    assert result.exit_code == 0
    mock_view.assert_called_once()
    _, _, content = mock_view.call_args.args
    assert "Hello world" in content


# ---------------------------------------------------------------------------
# Interactive edit — one candidate auto-selects
# ---------------------------------------------------------------------------


def test_edit_interactive_one_candidate_auto_selects(project: Path, fake_home: Path) -> None:
    _write(project / ".opencode" / "commands" / "greet.md", "# Greet")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared.sys") as mock_sys,
        patch("agents_cli.commands._shared._open_in_editor", return_value=0) as mock_editor,
    ):
        mock_sys.stdin.isatty.return_value = True
        result = runner.invoke(app, ["edit"])

    assert result.exit_code == 0
    mock_editor.assert_called_once()


# ---------------------------------------------------------------------------
# Interactive show — multiple candidates, select by index
# ---------------------------------------------------------------------------


def test_show_interactive_multiple_candidates_selects_by_index(
    project: Path, fake_home: Path
) -> None:
    _write(project / ".opencode" / "commands" / "alpha.md", "alpha content")
    _write(project / ".opencode" / "commands" / "beta.md", "beta content")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    viewed_content: list[str] = []

    def capture_view(title: str, subtitle: str | None, content: str) -> None:
        viewed_content.append(content)

    # Simulate user typing "2" (second resource, sorted alpha → beta).
    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared.sys") as mock_sys,
        patch("agents_cli.ui.view_text", side_effect=capture_view),
        patch("builtins.input", return_value="2"),
    ):
        mock_sys.stdin.isatty.return_value = True
        result = runner.invoke(app, ["show"])

    assert result.exit_code == 0
    assert viewed_content == ["beta content"]


# ---------------------------------------------------------------------------
# Interactive edit — multiple candidates, select by index
# ---------------------------------------------------------------------------


def test_edit_interactive_multiple_candidates_selects_by_index(
    project: Path, fake_home: Path
) -> None:
    _write(project / ".opencode" / "commands" / "alpha.md", "alpha")
    _write(project / ".opencode" / "commands" / "beta.md", "beta")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    edited_paths: list[Path] = []

    def capture_editor(path: Path) -> int:
        edited_paths.append(path)
        return 0

    # Simulate user typing "1" (first resource = alpha).
    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared.sys") as mock_sys,
        patch("agents_cli.commands._shared._open_in_editor", side_effect=capture_editor),
        patch("builtins.input", return_value="1"),
    ):
        mock_sys.stdin.isatty.return_value = True
        result = runner.invoke(app, ["edit"])

    assert result.exit_code == 0
    assert len(edited_paths) == 1
    assert edited_paths[0].name == "alpha.md"


# ---------------------------------------------------------------------------
# Interactive — cancel from picker exits 0
# ---------------------------------------------------------------------------


def test_show_interactive_cancel_exits_0(project: Path, fake_home: Path) -> None:
    _write(project / ".opencode" / "commands" / "alpha.md", "alpha")
    _write(project / ".opencode" / "commands" / "beta.md", "beta")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared.sys") as mock_sys,
        patch("builtins.input", return_value="q"),
    ):
        mock_sys.stdin.isatty.return_value = True
        result = runner.invoke(app, ["show"])

    assert result.exit_code == 0


def test_edit_interactive_cancel_exits_0(project: Path, fake_home: Path) -> None:
    _write(project / ".opencode" / "commands" / "alpha.md", "alpha")
    _write(project / ".opencode" / "commands" / "beta.md", "beta")
    app = _cmd_app()
    ctx = ScopeContext(project_root=project, home=fake_home)

    with (
        patch("agents_cli.commands._shared.detect", return_value=ctx),
        patch("agents_cli.commands._shared.cfgmod.load", return_value=_default_cfg()),
        patch("agents_cli.commands._shared.sys") as mock_sys,
        patch("builtins.input", return_value="q"),
    ):
        mock_sys.stdin.isatty.return_value = True
        result = runner.invoke(app, ["edit"])

    assert result.exit_code == 0


# ---------------------------------------------------------------------------
# Alias / symlink row prefixes in picker
# ---------------------------------------------------------------------------


def test_pick_resource_interactive_symlink_prefix(project: Path, fake_home: Path) -> None:
    """Symlink entries get the '↪ ' prefix in the picker table."""
    from agents_cli.discovery import DiscoveredResource

    ctx = ScopeContext(project_root=project, home=fake_home)
    entry = project / "foo.md"
    entry.write_text("x")

    sym_res = DiscoveredResource(
        harness_id="opencode",
        rtype=ResourceType.COMMAND,
        scope=Scope.PROJECT,
        name="foo",
        path=entry,
        entry=entry,
        is_symlink=True,
        symlink_target=Path("/real/foo.md"),
        size=1,
        is_alias_location=False,
    )
    alias_res = DiscoveredResource(
        harness_id="opencode",
        rtype=ResourceType.COMMAND,
        scope=Scope.PROJECT,
        name="bar",
        path=entry,
        entry=entry,
        is_symlink=False,
        symlink_target=None,
        size=1,
        is_alias_location=True,
    )

    rows_seen: list[tuple] = []

    def capture_table(title: Any, *, columns: Any, rows: Any) -> None:
        rows_seen.extend(rows)

    with (
        patch("agents_cli.ui.table", side_effect=capture_table),
        patch("agents_cli.ui.banner"),
        patch("agents_cli.ui.prompt_choice", return_value=None),
    ):
        _pick_resource_interactive(
            resources=[sym_res, alias_res],
            ctx=ctx,
            title="Test",
        )

    names = [r[3] for r in rows_seen]  # Name column is index 3 (after #, Harness, Scope)
    assert names[0].startswith("↪ "), f"symlink prefix missing: {names[0]!r}"
    assert names[1].startswith("~ "), f"alias prefix missing: {names[1]!r}"


# ---------------------------------------------------------------------------
# ui.prompt_choice
# ---------------------------------------------------------------------------


def test_prompt_choice_valid_input() -> None:
    with patch("builtins.input", return_value="2"):
        result = ui.prompt_choice(3)
    assert result == 1  # 0-based


def test_prompt_choice_q_cancels() -> None:
    with patch("builtins.input", return_value="q"):
        result = ui.prompt_choice(5)
    assert result is None


def test_prompt_choice_quit_cancels() -> None:
    with patch("builtins.input", return_value="quit"):
        result = ui.prompt_choice(5)
    assert result is None


def test_prompt_choice_empty_cancels() -> None:
    with patch("builtins.input", return_value=""):
        result = ui.prompt_choice(5)
    assert result is None


def test_prompt_choice_eof_cancels() -> None:
    with patch("builtins.input", side_effect=EOFError):
        result = ui.prompt_choice(3)
    assert result is None


def test_prompt_choice_invalid_then_valid() -> None:
    """Bad input triggers reprompt; second attempt succeeds."""
    inputs = iter(["abc", "99", "1"])
    with patch("builtins.input", side_effect=inputs):
        result = ui.prompt_choice(3)
    assert result == 0


def test_prompt_choice_out_of_range_then_cancel() -> None:
    inputs = iter(["0", "4", "q"])
    with patch("builtins.input", side_effect=inputs):
        result = ui.prompt_choice(3)
    assert result is None


# ---------------------------------------------------------------------------
# ui.view_text
# ---------------------------------------------------------------------------


def test_view_text_non_tty_prints_content(capsys: pytest.CaptureFixture[str]) -> None:
    """Non-TTY stdout: banner + content go through console.print (no pager)."""
    # stdout is not a TTY in pytest, so the plain-print branch runs.
    ui.view_text("My Title", "some/path", "Hello from content")
    captured = capsys.readouterr()
    assert "My Title" in captured.out
    assert "Hello from content" in captured.out


def test_view_text_subtitle_optional(capsys: pytest.CaptureFixture[str]) -> None:
    ui.view_text("Title Only", None, "Body text")
    captured = capsys.readouterr()
    assert "Title Only" in captured.out
    assert "Body text" in captured.out
