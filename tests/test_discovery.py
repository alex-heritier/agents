"""Tests for resource discovery."""

from __future__ import annotations

from pathlib import Path

from agents_cli.discovery import discover
from agents_cli.harnesses import ResourceType, get
from agents_cli.paths import Scope, ScopeContext


def _write(path: Path, content: str = "") -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)


def test_discover_guideline_walks_subdirs(project: Path, ctx: ScopeContext) -> None:
    _write(project / "AGENTS.md", "# root")
    _write(project / "packages" / "api" / "AGENTS.md", "# api")
    found = discover(ResourceType.GUIDELINE, Scope.PROJECT, ctx, harnesses=(get("opencode"),))
    paths = {str(r.path) for r in found}
    assert str(project / "AGENTS.md") in paths
    assert str(project / "packages" / "api" / "AGENTS.md") in paths


def test_discover_guideline_ignores_noise(project: Path, ctx: ScopeContext) -> None:
    _write(project / "node_modules" / "foo" / "AGENTS.md", "noise")
    _write(project / ".venv" / "AGENTS.md", "noise")
    _write(project / "AGENTS.md", "real")
    found = discover(ResourceType.GUIDELINE, Scope.PROJECT, ctx, harnesses=(get("opencode"),))
    paths = {str(r.path) for r in found}
    assert str(project / "node_modules" / "foo" / "AGENTS.md") not in paths
    assert str(project / ".venv" / "AGENTS.md") not in paths
    assert str(project / "AGENTS.md") in paths


def test_discover_skill_finds_directory_with_entry_file(project: Path, ctx: ScopeContext) -> None:
    _write(project / ".opencode" / "skills" / "my-skill" / "SKILL.md", "---\nname: x\n---\n")
    found = discover(ResourceType.SKILL, Scope.PROJECT, ctx, harnesses=(get("opencode"),))
    assert [r.name for r in found] == ["my-skill"]
    assert found[0].path == project / ".opencode" / "skills" / "my-skill"


def test_discover_skill_ignores_dirs_without_entry_file(project: Path, ctx: ScopeContext) -> None:
    (project / ".opencode" / "skills" / "empty").mkdir(parents=True)
    found = discover(ResourceType.SKILL, Scope.PROJECT, ctx, harnesses=(get("opencode"),))
    assert found == []


def test_discover_command_and_subagent(project: Path, ctx: ScopeContext) -> None:
    _write(project / ".opencode" / "commands" / "commit.md", "cmd")
    _write(project / ".opencode" / "agents" / "reviewer.md", "agent")
    cmds = discover(ResourceType.COMMAND, Scope.PROJECT, ctx, harnesses=(get("opencode"),))
    assert [c.name for c in cmds] == ["commit"]
    agts = discover(ResourceType.SUBAGENT, Scope.PROJECT, ctx, harnesses=(get("opencode"),))
    assert [a.name for a in agts] == ["reviewer"]


def test_discover_follows_symlinked_skill_directories(project: Path, ctx: ScopeContext) -> None:
    # Simulate a synced skill: .claude/skills/x is a symlink to .opencode/skills/x
    real = project / ".opencode" / "skills" / "helper"
    real.mkdir(parents=True)
    (real / "SKILL.md").write_text("---\nname: helper\n---\n")
    (project / ".claude" / "skills").mkdir(parents=True)
    (project / ".claude" / "skills" / "helper").symlink_to(real)

    found = discover(ResourceType.SKILL, Scope.PROJECT, ctx, harnesses=(get("claude"),))
    assert [r.name for r in found] == ["helper"]
