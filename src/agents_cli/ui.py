"""Rich-based UI helpers. Every bit of user-visible output flows through here."""

from __future__ import annotations

from collections.abc import Iterable

from rich.console import Console
from rich.table import Table
from rich.text import Text
from rich.theme import Theme

_theme = Theme(
    {
        "primary": "bold cyan",
        "secondary": "cyan",
        "muted": "dim white",
        "success": "bold green",
        "warn": "yellow",
        "error": "bold red",
        "accent": "magenta",
        "harness": "bold blue",
        "path": "cyan",
        "name": "bold white",
        "kind": "italic dim",
    }
)

console = Console(theme=_theme, highlight=False)
err_console = Console(theme=_theme, stderr=True, highlight=False)


def banner(title: str, subtitle: str | None = None) -> None:
    console.print(Text(title, style="primary"))
    if subtitle:
        console.print(Text(subtitle, style="muted"))


def hint(msg: str) -> None:
    console.print(Text(f"  {msg}", style="muted"))


def ok(msg: str) -> None:
    console.print(Text(msg, style="success"))


def warn(msg: str) -> None:
    err_console.print(Text(msg, style="warn"))


def error(msg: str) -> None:
    err_console.print(Text(msg, style="error"))


def section(title: str) -> None:
    console.print()
    console.print(Text(title, style="primary"))


def table(
    title: str | None,
    columns: Iterable[str],
    rows: Iterable[Iterable[object]],
) -> None:
    t = Table(title=title, show_header=True, header_style="bold", box=None, pad_edge=False)
    for col in columns:
        t.add_column(col, overflow="fold")
    any_row = False
    for row in rows:
        any_row = True
        t.add_row(*(str(c) for c in row))
    if not any_row:
        console.print(Text("  (nothing to show)", style="muted"))
        return
    console.print(t)


def confirm(prompt: str, *, default: bool = False, assume_yes: bool = False) -> bool:
    if assume_yes:
        return True
    import sys

    if not sys.stdin.isatty():
        return default
    suffix = " [Y/n]" if default else " [y/N]"
    try:
        answer = input(prompt + suffix + " ").strip().lower()
    except EOFError:
        return default
    if not answer:
        return default
    return answer in {"y", "yes"}


def format_summary(
    *,
    created: int,
    updated: int,
    skipped: int,
    deleted: int,
    failed: int,
    dry_run: bool,
) -> None:
    parts = []
    if created:
        parts.append(f"[success]+{created} created[/success]")
    if updated:
        parts.append(f"[accent]~{updated} updated[/accent]")
    if skipped:
        parts.append(f"[muted]={skipped} skipped[/muted]")
    if deleted:
        parts.append(f"[warn]-{deleted} deleted[/warn]")
    if failed:
        parts.append(f"[error]!{failed} failed[/error]")
    if not parts:
        parts.append("[muted]no changes[/muted]")
    prefix = "[muted](dry run)[/muted] " if dry_run else ""
    console.print(prefix + "  ".join(parts))
