"""Harness knowledge base.

This module encodes the file/directory conventions for every AI coding
harness we know about. Every path below is verified against each tool's
official documentation and/or source. Paths in ``{}`` braces are
interpolated at runtime:

    {name}  -> the resource slug (skill name, command name, subagent name)
    {home}  -> os.path.expanduser("~")

For resources whose "primary location" can legitimately be more than one
canonical path (e.g. Claude Code's ``./CLAUDE.md`` *or* ``./.claude/CLAUDE.md``),
we pick the one the harness writes to by default and list the alternative
under ``project_aliases`` so we still *discover* resources written there.

Guideline files are a special case: the value of ``project_path`` is a
single file (no ``{name}`` because there is only one canonical file per
directory). Nested discovery (walking up or down the tree) is handled
separately by the discovery layer.
"""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum


class ResourceType(str, Enum):
    """Kinds of artifacts a harness can consume."""

    GUIDELINE = "guideline"
    SKILL = "skill"
    COMMAND = "command"
    SUBAGENT = "subagent"


class Kind(str, Enum):
    """Whether a resource is stored as a single file or as a directory."""

    FILE = "file"
    DIRECTORY = "directory"


class Format(str, Enum):
    MARKDOWN = "markdown"
    TOML = "toml"


@dataclass(frozen=True)
class ResourceSpec:
    """Describes where a single harness stores one resource type.

    Attributes:
        project_path: canonical relative path from project root. Use
            ``{name}`` where the resource slug appears. For guidelines,
            this is a single file name with no ``{name}``.
        global_path: canonical absolute path from ``~``. May be ``None``
            if the harness has no documented user-scope path.
        project_aliases: additional paths that are also read by the
            harness. We write to ``project_path`` but discover resources
            at any of these too.
        global_aliases: same, but for global scope.
        kind: FILE (single file) or DIRECTORY (contains SKILL.md etc.).
        format: file format for FILE resources. DIRECTORY resources are
            always a directory on disk and always contain at least one
            markdown file (SKILL.md).
        entry_file: For DIRECTORY resources, the canonical entry file
            inside the directory (e.g. ``SKILL.md``).
    """

    project_path: str
    global_path: str | None = None
    project_aliases: tuple[str, ...] = ()
    global_aliases: tuple[str, ...] = ()
    kind: Kind = Kind.FILE
    format: Format = Format.MARKDOWN
    entry_file: str | None = None


@dataclass(frozen=True)
class Harness:
    """A single AI coding harness (opencode, Claude Code, Gemini CLI, ...)."""

    id: str
    """Stable identifier used on the CLI: ``--opencode``, ``--claude``."""

    display_name: str
    """Human-facing name: ``OpenCode``."""

    aliases: tuple[str, ...] = ()
    """Extra names accepted on the CLI (e.g. ``cc`` for claude)."""

    guideline: ResourceSpec | None = None
    skill: ResourceSpec | None = None
    command: ResourceSpec | None = None
    subagent: ResourceSpec | None = None

    notes: str = ""
    """Notes surfaced by ``agents doctor``."""

    def spec(self, rtype: ResourceType) -> ResourceSpec | None:
        return {
            ResourceType.GUIDELINE: self.guideline,
            ResourceType.SKILL: self.skill,
            ResourceType.COMMAND: self.command,
            ResourceType.SUBAGENT: self.subagent,
        }[rtype]

    def supports(self, rtype: ResourceType) -> bool:
        return self.spec(rtype) is not None


# -----------------------------------------------------------------------
# OpenCode — https://opencode.ai/docs (verified 2026-04)
# -----------------------------------------------------------------------
# - Plural dirs are canonical (agents/, commands/, skills/), singular aliases
#   ("agent/", "command/", "skill/") are accepted for backwards compat.
# - Global config dir is ~/.config/opencode (NOT XDG-aware, not ~/.opencode).
# - Also reads Claude-compat paths: ~/.claude/CLAUDE.md, .claude/skills/.
OPENCODE = Harness(
    id="opencode",
    display_name="OpenCode",
    aliases=("oc",),
    guideline=ResourceSpec(
        project_path="AGENTS.md",
        global_path="~/.config/opencode/AGENTS.md",
        project_aliases=("CLAUDE.md",),
        global_aliases=("~/.claude/CLAUDE.md",),
    ),
    skill=ResourceSpec(
        project_path=".opencode/skills/{name}/SKILL.md",
        global_path="~/.config/opencode/skills/{name}/SKILL.md",
        project_aliases=(
            ".opencode/skill/{name}/SKILL.md",
            ".claude/skills/{name}/SKILL.md",
            ".agents/skills/{name}/SKILL.md",
        ),
        global_aliases=(
            "~/.claude/skills/{name}/SKILL.md",
            "~/.agents/skills/{name}/SKILL.md",
        ),
        kind=Kind.DIRECTORY,
        entry_file="SKILL.md",
    ),
    command=ResourceSpec(
        project_path=".opencode/commands/{name}.md",
        global_path="~/.config/opencode/commands/{name}.md",
        project_aliases=(".opencode/command/{name}.md",),
    ),
    subagent=ResourceSpec(
        project_path=".opencode/agents/{name}.md",
        global_path="~/.config/opencode/agents/{name}.md",
        project_aliases=(".opencode/agent/{name}.md",),
    ),
    notes="Plural dir names are canonical; set OPENCODE_DISABLE_CLAUDE_CODE=1 to ignore ~/.claude/.",
)


# -----------------------------------------------------------------------
# Claude Code — https://docs.claude.com/en/docs/claude-code (verified 2026-04)
# -----------------------------------------------------------------------
# - Guidelines: ./CLAUDE.md (preferred) or .claude/CLAUDE.md.
# - Skills directory is the canonical way to ship custom functionality;
#   .claude/commands/*.md still works (commands were merged into skills).
CLAUDE = Harness(
    id="claude",
    display_name="Claude Code",
    aliases=("cc",),
    guideline=ResourceSpec(
        project_path="CLAUDE.md",
        global_path="~/.claude/CLAUDE.md",
        project_aliases=(".claude/CLAUDE.md", "CLAUDE.local.md"),
    ),
    skill=ResourceSpec(
        project_path=".claude/skills/{name}/SKILL.md",
        global_path="~/.claude/skills/{name}/SKILL.md",
        kind=Kind.DIRECTORY,
        entry_file="SKILL.md",
    ),
    command=ResourceSpec(
        project_path=".claude/commands/{name}.md",
        global_path="~/.claude/commands/{name}.md",
    ),
    subagent=ResourceSpec(
        project_path=".claude/agents/{name}.md",
        global_path="~/.claude/agents/{name}.md",
    ),
    notes="Commands have been merged into skills; both paths still work.",
)


# -----------------------------------------------------------------------
# Gemini CLI — https://geminicli.com/docs (verified 2026-04)
# -----------------------------------------------------------------------
# - Default context filename is GEMINI.md (configurable in settings.json).
# - Custom commands are TOML, not markdown.
# - Skills follow the .agents/ cross-tool standard.
GEMINI = Harness(
    id="gemini",
    display_name="Gemini CLI",
    guideline=ResourceSpec(
        project_path="GEMINI.md",
        global_path="~/.gemini/GEMINI.md",
    ),
    skill=ResourceSpec(
        project_path=".gemini/skills/{name}/SKILL.md",
        global_path="~/.gemini/skills/{name}/SKILL.md",
        project_aliases=(".agents/skills/{name}/SKILL.md",),
        global_aliases=("~/.agents/skills/{name}/SKILL.md",),
        kind=Kind.DIRECTORY,
        entry_file="SKILL.md",
    ),
    command=ResourceSpec(
        project_path=".gemini/commands/{name}.toml",
        global_path="~/.gemini/commands/{name}.toml",
        format=Format.TOML,
    ),
    subagent=ResourceSpec(
        project_path=".gemini/agents/{name}.md",
        global_path="~/.gemini/agents/{name}.md",
    ),
    notes="Custom commands are TOML (prompt + description). Subagents require experimental.enableAgents=true.",
)


# -----------------------------------------------------------------------
# Droid (Factory AI) — https://docs.factory.ai (verified 2026-04)
# -----------------------------------------------------------------------
# - Config root is .factory/ (project) and ~/.factory/ (user).
# - Subagents are called "droids" and live at .factory/droids/*.md.
# - Skills are the Anthropic-style SKILL.md directories.
DROID = Harness(
    id="droid",
    display_name="Droid (Factory)",
    aliases=("factory",),
    guideline=ResourceSpec(
        project_path="AGENTS.md",
        global_path="~/.factory/AGENTS.md",
    ),
    skill=ResourceSpec(
        project_path=".factory/skills/{name}/SKILL.md",
        global_path="~/.factory/skills/{name}/SKILL.md",
        project_aliases=(".agent/skills/{name}/SKILL.md",),
        kind=Kind.DIRECTORY,
        entry_file="SKILL.md",
    ),
    command=ResourceSpec(
        project_path=".factory/commands/{name}.md",
        global_path="~/.factory/commands/{name}.md",
    ),
    subagent=ResourceSpec(
        project_path=".factory/droids/{name}.md",
        global_path="~/.factory/droids/{name}.md",
    ),
    notes="Subagents are 'custom droids'. Uses the open AGENTS.md standard.",
)


# -----------------------------------------------------------------------
# Kilo Code CLI — https://kilo.ai/docs (verified 2026-04, v1.0+)
# -----------------------------------------------------------------------
# - New platform uses .kilo/ (singular, no "code"). Legacy was .kilocode/.
# - Global config is ~/.config/kilo/.
# - Skills follow the Agent Skills standard.
KILO = Harness(
    id="kilo",
    display_name="Kilo Code",
    aliases=("kilocode",),
    guideline=ResourceSpec(
        project_path="AGENTS.md",
        global_path="~/.config/kilo/AGENTS.md",
        project_aliases=("AGENT.md", "CLAUDE.md", "CONTEXT.md"),
    ),
    skill=ResourceSpec(
        project_path=".kilo/skills/{name}/SKILL.md",
        global_path="~/.kilo/skills/{name}/SKILL.md",
        project_aliases=(
            ".claude/skills/{name}/SKILL.md",
            ".agents/skills/{name}/SKILL.md",
        ),
        kind=Kind.DIRECTORY,
        entry_file="SKILL.md",
    ),
    command=ResourceSpec(
        project_path=".kilo/commands/{name}.md",
        global_path="~/.config/kilo/commands/{name}.md",
        project_aliases=(".kilocode/workflows/{name}.md",),
        global_aliases=("~/.kilocode/workflows/{name}.md",),
    ),
    subagent=ResourceSpec(
        project_path=".kilo/agents/{name}.md",
        global_path="~/.config/kilo/agent/{name}.md",
        project_aliases=(".kilo/agent/{name}.md", ".opencode/agents/{name}.md"),
    ),
    notes="CLI binary is 'kilo'. Built on a fork of OpenCode.",
)


# -----------------------------------------------------------------------
# Codex (OpenAI) — https://developers.openai.com/codex
# -----------------------------------------------------------------------
CODEX = Harness(
    id="codex",
    display_name="Codex",
    guideline=ResourceSpec(
        project_path="AGENTS.md",
        global_path="~/.codex/AGENTS.md",
        project_aliases=("AGENTS.override.md",),
    ),
    skill=ResourceSpec(
        project_path=".codex/skills/{name}/SKILL.md",
        global_path="~/.codex/skills/{name}/SKILL.md",
        kind=Kind.DIRECTORY,
        entry_file="SKILL.md",
    ),
    notes="Codex loads AGENTS.md. Slash commands and subagents not documented publicly.",
)


# -----------------------------------------------------------------------
# Cursor — https://cursor.com/docs
# -----------------------------------------------------------------------
CURSOR = Harness(
    id="cursor",
    display_name="Cursor",
    guideline=ResourceSpec(
        project_path="AGENTS.md",
        # Cursor's user/team rules live in the app settings UI, not on disk.
        global_path=None,
        project_aliases=(".cursorrules", ".cursor/rules/agents.md"),
    ),
    command=ResourceSpec(
        project_path=".cursor/commands/{name}.md",
        global_path="~/.cursor/commands/{name}.md",
    ),
    notes="Rules: .cursor/rules/<rule>/RULE.md or AGENTS.md. User rules managed via Settings UI.",
)


# -----------------------------------------------------------------------
# Amp Code — https://ampcode.com/manual
# -----------------------------------------------------------------------
AMP = Harness(
    id="amp",
    display_name="Amp Code",
    guideline=ResourceSpec(
        project_path="AGENTS.md",
        global_path="~/.config/amp/AGENTS.md",
        project_aliases=("AGENT.md", "CLAUDE.md"),
        global_aliases=("~/.config/AGENTS.md",),
    ),
    skill=ResourceSpec(
        project_path=".agents/skills/{name}/SKILL.md",
        global_path="~/.config/amp/skills/{name}/SKILL.md",
        project_aliases=(".claude/skills/{name}/SKILL.md",),
        kind=Kind.DIRECTORY,
        entry_file="SKILL.md",
    ),
    command=ResourceSpec(
        project_path=".agents/commands/{name}.md",
        global_path="~/.config/amp/commands/{name}.md",
    ),
)


# -----------------------------------------------------------------------
# Qwen Code CLI — https://github.com/QwenLM/qwen-code
# -----------------------------------------------------------------------
# Qwen-Code follows the Gemini CLI layout closely (fork-of-fork lineage).
QWEN = Harness(
    id="qwen",
    display_name="Qwen Code",
    guideline=ResourceSpec(
        project_path="QWEN.md",
        global_path="~/.qwen/QWEN.md",
        project_aliases=("AGENTS.md",),
    ),
    command=ResourceSpec(
        project_path=".qwen/commands/{name}.toml",
        global_path="~/.qwen/commands/{name}.toml",
        format=Format.TOML,
    ),
    notes="Follows Gemini CLI conventions.",
)


# -----------------------------------------------------------------------
# GitHub Copilot — has .github/copilot-instructions.md as the canonical guideline.
# -----------------------------------------------------------------------
COPILOT = Harness(
    id="copilot",
    display_name="GitHub Copilot",
    guideline=ResourceSpec(
        project_path=".github/copilot-instructions.md",
        global_path=None,
        project_aliases=("COPILOT.md",),
    ),
    notes="Copilot reads .github/copilot-instructions.md.",
)


# Order matters for documentation: most commonly-used first.
ALL_HARNESSES: tuple[Harness, ...] = (
    OPENCODE,
    CLAUDE,
    GEMINI,
    DROID,
    KILO,
    CODEX,
    CURSOR,
    AMP,
    QWEN,
    COPILOT,
)


_BY_ID: dict[str, Harness] = {}
for _h in ALL_HARNESSES:
    _BY_ID[_h.id] = _h
    for _alias in _h.aliases:
        _BY_ID[_alias] = _h


def get(harness_id: str) -> Harness:
    try:
        return _BY_ID[harness_id.lower()]
    except KeyError as exc:
        known = ", ".join(h.id for h in ALL_HARNESSES)
        raise KeyError(f"unknown harness {harness_id!r}. Known: {known}") from exc


def try_get(harness_id: str) -> Harness | None:
    return _BY_ID.get(harness_id.lower())


def iter_supporting(rtype: ResourceType) -> list[Harness]:
    """All harnesses that support a given resource type."""
    return [h for h in ALL_HARNESSES if h.supports(rtype)]


# Optional metadata about the resource types themselves (used by help text).
RESOURCE_DESCRIPTIONS: dict[ResourceType, str] = {
    ResourceType.GUIDELINE: "Primary agent instructions (AGENTS.md / CLAUDE.md / GEMINI.md / …).",
    ResourceType.SKILL: "Reusable capability (Anthropic-style SKILL.md folder).",
    ResourceType.COMMAND: "Custom slash command (e.g. /review, /ship).",
    ResourceType.SUBAGENT: "Specialized subagent invoked by the primary agent.",
}


@dataclass(frozen=True)
class ResolvedPath:
    """A concrete, interpolated filesystem path produced by ``render``."""

    raw: str
    path: str  # Absolute or project-relative depending on scope.
    is_alias: bool = False


def _fmt(template: str, name: str) -> str:
    # Simple interpolation — avoids accidentally interpreting braces in paths.
    return template.replace("{name}", name) if "{name}" in template else template
