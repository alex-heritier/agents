"""Shared pytest fixtures.

All tests are hermetic: they operate on ``tmp_path`` and never touch the
real ``$HOME``.
"""

from __future__ import annotations

import subprocess
from pathlib import Path

import pytest

from agents_cli.paths import ScopeContext


@pytest.fixture
def project(tmp_path: Path) -> Path:
    """Initialise a git repo in ``tmp_path`` and return its root."""
    subprocess.run(["git", "init", "-q"], cwd=tmp_path, check=True)
    return tmp_path


@pytest.fixture
def fake_home(tmp_path_factory: pytest.TempPathFactory) -> Path:
    """Return a disposable ``$HOME`` directory, never the real one."""
    return tmp_path_factory.mktemp("home")


@pytest.fixture
def ctx(project: Path, fake_home: Path) -> ScopeContext:
    """A ``ScopeContext`` rooted in ``project`` with a disposable ``$HOME``."""
    return ScopeContext(project_root=project, home=fake_home)
