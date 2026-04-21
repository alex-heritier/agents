from __future__ import annotations

from agents_cli.commands._shared import build_group
from agents_cli.harnesses import ResourceType

app = build_group(
    name="command",
    rtype=ResourceType.COMMAND,
    help_text="Custom slash commands (/review, /ship, …).",
)
