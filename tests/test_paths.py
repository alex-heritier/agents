"""Tests for path resolution and scope detection."""

from __future__ import annotations

from pathlib import Path

from agents_cli.harnesses import ResourceType, get
from agents_cli.paths import Scope, ScopeContext, display_path, render_paths


def test_render_paths_project_opencode_guideline(ctx: ScopeContext) -> None:
    paths = render_paths(get("opencode"), ResourceType.GUIDELINE, "AGENTS.md", Scope.PROJECT, ctx)
    assert paths[0].path == str(ctx.project_root / "AGENTS.md")
    # OpenCode treats CLAUDE.md as an alias.
    assert any(p.path == str(ctx.project_root / "CLAUDE.md") for p in paths)


def test_render_paths_project_skill_has_name_substituted(ctx: ScopeContext) -> None:
    paths = render_paths(get("opencode"), ResourceType.SKILL, "my-skill", Scope.PROJECT, ctx)
    assert paths[0].path.endswith(".opencode/skills/my-skill/SKILL.md")


def test_render_paths_global_uses_expanduser(ctx: ScopeContext, tmp_path: Path) -> None:
    paths = render_paths(get("opencode"), ResourceType.GUIDELINE, "AGENTS.md", Scope.GLOBAL, ctx)
    # Primary should end with .config/opencode/AGENTS.md and be absolute.
    assert paths[0].path.endswith(".config/opencode/AGENTS.md")
    assert paths[0].path.startswith("/")


def test_render_paths_returns_empty_when_unsupported(ctx: ScopeContext) -> None:
    # Copilot does not define subagents.
    assert render_paths(get("copilot"), ResourceType.SUBAGENT, "x", Scope.PROJECT, ctx) == []


def test_display_path_relative_to_project(ctx: ScopeContext) -> None:
    assert ctx.project_root is not None
    p = ctx.project_root / "AGENTS.md"
    assert display_path(p, ctx) == "./AGENTS.md"


def test_display_path_inside_home(ctx: ScopeContext) -> None:
    p = ctx.home / "foo" / "bar.md"
    assert display_path(p, ctx) == "~/foo/bar.md"


def test_display_path_does_not_follow_symlinks(tmp_path: Path) -> None:
    target = tmp_path / "real.md"
    target.write_text("hi")
    link = tmp_path / "link.md"
    link.symlink_to(target)
    ctx = ScopeContext(project_root=tmp_path, home=tmp_path)
    # Must show the symlink's own path, not the resolved target.
    assert display_path(link, ctx) == "./link.md"
