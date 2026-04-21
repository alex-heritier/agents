from __future__ import annotations

from agents_cli.commands._shared import build_group
from agents_cli.harnesses import ResourceType

app = build_group(
    name="guideline",
    rtype=ResourceType.GUIDELINE,
    help_text=(
        "Agent guideline files — AGENTS.md, CLAUDE.md, GEMINI.md, and friends.\n\n"
        "Discovery walks up from the git root *and* into subdirectories for nested rules."
    ),
)
