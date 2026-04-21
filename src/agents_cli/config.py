"""Project and user configuration.

Users optionally drop a ``.agents.toml`` (project root) or
``~/.config/agents-cli/config.toml`` (user) to encode their preferences:

    source = "opencode"
    targets = ["claude", "gemini", "droid"]
    sync_method = "symlink"  # or "copy"

    [methods]
    skill = "symlink"
    command = "copy"

Everything is optional; anything not configured falls back to the defaults
below (which happen to be "opencode" source and "symlink" sync method —
matching the original author's use case but easy to override).
"""

from __future__ import annotations

import os
import sys
from dataclasses import dataclass, field
from pathlib import Path

if sys.version_info >= (3, 11):
    import tomllib
else:  # pragma: no cover
    import tomli as tomllib

from agents_cli.harnesses import ResourceType, get

PROJECT_CONFIG_FILES = (".agents.toml", "agents.toml")
USER_CONFIG_FILE = "~/.config/agents-cli/config.toml"

DEFAULT_SOURCE = "opencode"
DEFAULT_METHOD = "symlink"


@dataclass(frozen=True)
class AgentsConfig:
    source: str = DEFAULT_SOURCE
    targets: tuple[str, ...] = ()
    sync_method: str = DEFAULT_METHOD
    per_type_method: dict[ResourceType, str] = field(default_factory=dict)

    def method_for(self, rtype: ResourceType) -> str:
        return self.per_type_method.get(rtype, self.sync_method)


def _read_toml(path: Path) -> dict:
    try:
        with path.open("rb") as f:
            return tomllib.load(f)
    except FileNotFoundError:
        return {}
    except tomllib.TOMLDecodeError as e:
        from agents_cli.ui import warn

        warn(f"Failed to parse {path}: {e}")
        return {}


def _merge(base: dict, override: dict) -> dict:
    merged = dict(base)
    for k, v in override.items():
        if isinstance(v, dict) and isinstance(merged.get(k), dict):
            merged[k] = _merge(merged[k], v)
        else:
            merged[k] = v
    return merged


def _find_project_config(project_root: Path | None) -> Path | None:
    if project_root is None:
        return None
    for name in PROJECT_CONFIG_FILES:
        p = project_root / name
        if p.is_file():
            return p
    return None


def load(project_root: Path | None = None) -> AgentsConfig:
    data: dict = {}

    user_path = Path(os.path.expanduser(USER_CONFIG_FILE))
    if user_path.is_file():
        data = _merge(data, _read_toml(user_path))

    proj_path = _find_project_config(project_root)
    if proj_path is not None:
        data = _merge(data, _read_toml(proj_path))

    source = str(data.get("source") or DEFAULT_SOURCE).lower()
    targets = tuple(str(t).lower() for t in data.get("targets", []))
    sync_method = str(data.get("sync_method") or DEFAULT_METHOD).lower()

    per_type: dict[ResourceType, str] = {}
    methods = data.get("methods") or {}
    if isinstance(methods, dict):
        for key, val in methods.items():
            try:
                rtype = ResourceType(str(key).lower())
            except ValueError:
                continue
            per_type[rtype] = str(val).lower()

    return AgentsConfig(
        source=source,
        targets=targets,
        sync_method=sync_method,
        per_type_method=per_type,
    )


def project_config_path(project_root: Path) -> Path:
    """Where ``agents init`` will write the project config."""
    return project_root / PROJECT_CONFIG_FILES[0]


def validate(cfg: AgentsConfig) -> list[str]:
    """Return a list of human-readable validation errors (empty = ok)."""
    errors: list[str] = []
    try:
        get(cfg.source)
    except KeyError:
        errors.append(f"Unknown source harness: {cfg.source!r}")
    for t in cfg.targets:
        try:
            get(t)
        except KeyError:
            errors.append(f"Unknown target harness: {t!r}")
    if cfg.sync_method not in {"symlink", "copy"}:
        errors.append(f"sync_method must be 'symlink' or 'copy' (got {cfg.sync_method!r})")
    for rtype, method in cfg.per_type_method.items():
        if method not in {"symlink", "copy"}:
            errors.append(f"methods.{rtype.value} must be 'symlink' or 'copy' (got {method!r})")
    return errors
