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


def prompt_choice(n: int, prompt: str = "Selection") -> int | None:
    """Prompt for a 1-based integer choice from ``n`` items.

    Returns the 0-based index of the chosen item, or ``None`` if the user
    cancels (empty input, ``q``/``quit``, or EOF).  Loops forever on invalid
    input so a TTY user can correct a typo without re-running the command.
    """
    while True:
        try:
            raw = input(f"{prompt} [1-{n}, q to cancel]: ").strip()
        except EOFError:
            return None
        if not raw or raw.lower() in {"q", "quit"}:
            return None
        try:
            idx = int(raw)
        except ValueError:
            err_console.print(
                Text(f"Invalid input — enter a number 1-{n} or q to cancel.", style="warn")
            )
            continue
        if 1 <= idx <= n:
            return idx - 1
        err_console.print(Text(f"Out of range — enter a number 1-{n}.", style="warn"))


def view_text(title: str, subtitle: str | None, content: str) -> None:
    """Display a read-only view of ``content`` with a banner header.

    Uses a Rich pager when stdout is a TTY so long files scroll nicely;
    falls back to plain print otherwise (e.g. piped output or tests).
    """
    import sys

    if sys.stdout.isatty():
        with console.pager():
            banner(title, subtitle)
            console.print(content)
    else:
        banner(title, subtitle)
        console.print(content)


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
