from __future__ import annotations

from agents_cli.commands._shared import build_group
from agents_cli.harnesses import ResourceType

app = build_group(
    name="subagent",
    rtype=ResourceType.SUBAGENT,
    help_text="Specialized subagents invoked by the primary agent.",
)
