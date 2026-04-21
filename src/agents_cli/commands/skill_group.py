from __future__ import annotations

from agents_cli.commands._shared import build_group
from agents_cli.harnesses import ResourceType

app = build_group(
    name="skill",
    rtype=ResourceType.SKILL,
    help_text="Anthropic-style skills (``<name>/SKILL.md`` directories).",
)
