"""Interactive TUI browser (``show``) and picker (``edit``).

Both classes are built on Textual and share the same discovery model:
``DiscoveredResource`` objects come in; the Apps return either an int exit
code (``ShowBrowser``) or a ``Path | None`` (``EditPicker``).

Public API
----------
* ``run_show_browser(*, resources, ctx, title) -> int``
* ``run_edit_picker(*, resources, ctx, title) -> Path | None``
* ``filter_resources(resources, query, ctx) -> list[int]``
* ``build_preview(res, ctx) -> str``
* ``format_resource_row(res, ctx, index) -> tuple[str, str, str, str]``
"""

from __future__ import annotations

from pathlib import Path
from typing import ClassVar

from textual import on
from textual.app import App, ComposeResult
from textual.binding import Binding
from textual.containers import Horizontal, Vertical, VerticalScroll
from textual.widgets import DataTable, Footer, Header, Input, Label, Static

from agents_cli.discovery import DiscoveredResource
from agents_cli.paths import Scope, ScopeContext, display_path

# ---------------------------------------------------------------------------
# Pure helper functions (easy to unit-test without Textual)
# ---------------------------------------------------------------------------


def _format_size(size: int) -> str:
    if size < 1024:
        return f"{size}B"
    if size < 1024 * 1024:
        return f"{size // 1024}K"
    return f"{size // (1024 * 1024)}M"


def filter_resources(
    resources: list[DiscoveredResource],
    query: str,
    ctx: ScopeContext,
) -> list[int]:
    """Return indices of resources whose fields contain ``query`` (case-insensitive).

    An empty query matches everything.
    """
    if not query:
        return list(range(len(resources)))
    q = query.lower()
    return [
        i
        for i, r in enumerate(resources)
        if q in r.display_name().lower()
        or q in display_path(r.path, ctx).lower()
        or q in r.scope.value.lower()
        or q in r.harness_id.lower()
    ]


def build_preview(res: DiscoveredResource, ctx: ScopeContext) -> str:
    """Build the right-pane preview text for a resource.

    Returns a plain-text string: metadata header, a rule, then file content.
    OSError on read is shown as an error line rather than crashing.
    """
    scope_label = "project" if res.scope is Scope.PROJECT else "global"
    path_str = display_path(res.entry, ctx)
    size_str = _format_size(res.size)
    header = f"Path: {path_str}    Scope: {scope_label}    Size: {size_str}\n" + "─" * 60 + "\n"
    try:
        body = res.entry.read_text()
    except OSError as exc:
        body = f"[Read error: {exc}]"
    return header + body


def format_resource_row(
    res: DiscoveredResource,
    ctx: ScopeContext,
    index: int,
) -> tuple[str, str, str, str]:
    """Return a ``(index, name, scope, size)`` tuple for list-pane rows."""
    mark = ""
    if res.is_symlink:
        mark = "↪ "
    elif res.is_alias_location:
        mark = "~ "
    scope_label = "project" if res.scope is Scope.PROJECT else "global"
    return (str(index), mark + res.display_name(), scope_label, _format_size(res.size))


# ---------------------------------------------------------------------------
# ShowBrowser — two-pane interactive browser
# ---------------------------------------------------------------------------

_SHOW_CSS = """\
ShowBrowser {
    layout: vertical;
}

#body {
    layout: horizontal;
    height: 1fr;
}

#list-pane {
    width: 38%;
    border: solid $accent;
    padding: 0 1;
    layout: vertical;
}

#list-header {
    height: 1;
    color: $accent;
    text-style: bold;
}

#list-table {
    height: 1fr;
}

#preview-pane {
    width: 1fr;
    border: solid $accent;
    padding: 0 1;
    layout: vertical;
}

#preview-header {
    height: 1;
    color: $accent;
    text-style: bold;
}

#preview-scroll {
    height: 1fr;
}

#preview-content {
    height: auto;
}

#show-filter-bar {
    height: 3;
    display: none;
    border: solid $warning;
    padding: 0 1;
}

#show-filter-bar.active {
    display: block;
}
"""


class ShowBrowser(App[int]):
    """Two-pane interactive browser for ``agents <resource> show``."""

    BINDINGS: ClassVar = [
        Binding("q", "quit_browser", "Quit", show=True),
        Binding("j", "navigate_down", "Down", show=False),
        Binding("k", "navigate_up", "Up", show=False),
        Binding("/", "start_filter", "Filter", show=True),
    ]

    CSS = _SHOW_CSS

    def __init__(
        self,
        resources: list[DiscoveredResource],
        ctx: ScopeContext,
        title: str,
    ) -> None:
        super().__init__()
        self._resources = resources
        self._ctx = ctx
        self._app_title = title
        self._filtered: list[int] = list(range(len(resources)))
        self._filter_active: bool = False
        self._filter_query: str = ""

    def compose(self) -> ComposeResult:
        yield Header(show_clock=False)
        with Horizontal(id="body"):
            with Vertical(id="list-pane"):
                yield Label("", id="list-header")
                yield DataTable(id="list-table", cursor_type="row", zebra_stripes=True)
            with Vertical(id="preview-pane"):
                yield Label("", id="preview-header")
                with VerticalScroll(id="preview-scroll"):
                    yield Static("", id="preview-content", markup=False)
        with Vertical(id="show-filter-bar"):
            yield Input(id="show-filter-input", placeholder="Filter…")
        yield Footer()

    def on_mount(self) -> None:
        self.title = self._app_title
        self.sub_title = "↑/↓  Navigate    /  Filter    q  Quit"
        self._rebuild_table()
        self._refresh_preview()

    # ------------------------------------------------------------------
    # Table helpers
    # ------------------------------------------------------------------

    def _rebuild_table(self) -> None:
        table = self.query_one("#list-table", DataTable)
        table.clear(columns=True)
        table.add_columns("#", "Name", "Scope", "Size")
        for pos, ridx in enumerate(self._filtered, start=1):
            res = self._resources[ridx]
            row = format_resource_row(res, self._ctx, pos)
            table.add_row(*row)
        count = len(self._filtered)
        total = len(self._resources)
        suffix = f" ({self._filter_query})" if self._filter_query else ""
        self.query_one("#list-header", Label).update(f"FILES {count}/{total}{suffix}")
        if self._filtered:
            table.move_cursor(row=0)

    def _refresh_preview(self) -> None:
        if not self._filtered:
            self.query_one("#preview-header", Label).update("(no results)")
            self.query_one("#preview-content", Static).update("")
            return
        table = self.query_one("#list-table", DataTable)
        cursor = table.cursor_row
        if cursor < 0 or cursor >= len(self._filtered):
            cursor = 0
        ridx = self._filtered[cursor]
        res = self._resources[ridx]
        self.query_one("#preview-header", Label).update(f"PREVIEW: {res.display_name()}")
        self.query_one("#preview-content", Static).update(build_preview(res, self._ctx))

    # ------------------------------------------------------------------
    # Cursor sync → preview
    # ------------------------------------------------------------------

    def on_data_table_row_highlighted(self, event: DataTable.RowHighlighted) -> None:
        self._refresh_preview()

    # ------------------------------------------------------------------
    # Key actions
    # ------------------------------------------------------------------

    def action_quit_browser(self) -> None:
        self.exit(0)

    def action_navigate_down(self) -> None:
        table = self.query_one("#list-table", DataTable)
        new_row = min(table.cursor_row + 1, table.row_count - 1)
        table.move_cursor(row=new_row)

    def action_navigate_up(self) -> None:
        table = self.query_one("#list-table", DataTable)
        new_row = max(table.cursor_row - 1, 0)
        table.move_cursor(row=new_row)

    def action_start_filter(self) -> None:
        bar = self.query_one("#show-filter-bar")
        bar.add_class("active")
        self._filter_active = True
        self.query_one("#show-filter-input", Input).focus()

    def key_escape(self) -> None:
        if self._filter_active:
            self._clear_filter()
            self.query_one("#list-table", DataTable).focus()

    def _clear_filter(self) -> None:
        self.query_one("#show-filter-bar").remove_class("active")
        self._filter_active = False
        inp = self.query_one("#show-filter-input", Input)
        inp.value = ""
        self._filter_query = ""
        self._filtered = list(range(len(self._resources)))
        self._rebuild_table()
        self._refresh_preview()

    # ------------------------------------------------------------------
    # Filter input events
    # ------------------------------------------------------------------

    @on(Input.Changed, "#show-filter-input")
    def _on_filter_changed(self, event: Input.Changed) -> None:
        self._filter_query = event.value
        self._filtered = filter_resources(self._resources, event.value, self._ctx)
        self._rebuild_table()
        self._refresh_preview()

    @on(Input.Submitted, "#show-filter-input")
    def _on_filter_submitted(self, event: Input.Submitted) -> None:
        self.query_one("#show-filter-bar").remove_class("active")
        self._filter_active = False
        self.query_one("#list-table", DataTable).focus()


# ---------------------------------------------------------------------------
# EditPicker — single-pane picker that returns a Path
# ---------------------------------------------------------------------------

_EDIT_CSS = """\
EditPicker {
    layout: vertical;
}

#edit-list-pane {
    height: 1fr;
    border: solid $accent;
    padding: 0 1;
    layout: vertical;
}

#edit-list-header {
    height: 1;
    color: $accent;
    text-style: bold;
}

#picker-table {
    height: 1fr;
}

#edit-filter-bar {
    height: 3;
    display: none;
    border: solid $warning;
    padding: 0 1;
}

#edit-filter-bar.active {
    display: block;
}
"""


class EditPicker(App[Path | None]):
    """Single-pane interactive picker for ``agents <resource> edit``."""

    BINDINGS: ClassVar = [
        Binding("q", "cancel", "Cancel", show=True),
        Binding("j", "navigate_down", "Down", show=False),
        Binding("k", "navigate_up", "Up", show=False),
        Binding("/", "start_filter", "Filter", show=True),
        Binding("enter", "select_resource", "Edit in $EDITOR", show=True),
    ]

    CSS = _EDIT_CSS

    def __init__(
        self,
        resources: list[DiscoveredResource],
        ctx: ScopeContext,
        title: str,
    ) -> None:
        super().__init__()
        self._resources = resources
        self._ctx = ctx
        self._app_title = title
        self._filtered: list[int] = list(range(len(resources)))
        self._filter_active: bool = False
        self._filter_query: str = ""

    def compose(self) -> ComposeResult:
        yield Header(show_clock=False)
        with Vertical(id="edit-list-pane"):
            yield Label("", id="edit-list-header")
            yield DataTable(id="picker-table", cursor_type="row", zebra_stripes=True)
        with Vertical(id="edit-filter-bar"):
            yield Input(id="edit-filter-input", placeholder="Filter…")
        yield Footer()

    def on_mount(self) -> None:
        self.title = self._app_title
        self.sub_title = "↑/↓  Navigate    Enter  Edit in $EDITOR    q  Cancel"
        self._rebuild_table()

    # ------------------------------------------------------------------
    # Table helpers
    # ------------------------------------------------------------------

    def _rebuild_table(self) -> None:
        table = self.query_one("#picker-table", DataTable)
        table.clear(columns=True)
        table.add_columns("#", "Name", "Scope", "Path", "Size")
        for pos, ridx in enumerate(self._filtered, start=1):
            res = self._resources[ridx]
            mark = ""
            if res.is_symlink:
                mark = "↪ "
            elif res.is_alias_location:
                mark = "~ "
            scope_label = "project" if res.scope is Scope.PROJECT else "global"
            table.add_row(
                str(pos),
                mark + res.display_name(),
                scope_label,
                display_path(res.path, self._ctx),
                _format_size(res.size),
            )
        count = len(self._filtered)
        total = len(self._resources)
        suffix = f" ({self._filter_query})" if self._filter_query else ""
        self.query_one("#edit-list-header", Label).update(
            f"SELECT A FILE TO EDIT  {count}/{total}{suffix}"
        )
        if self._filtered:
            table.move_cursor(row=0)

    def _selected_entry(self) -> Path | None:
        if not self._filtered:
            return None
        table = self.query_one("#picker-table", DataTable)
        cursor = table.cursor_row
        if cursor < 0 or cursor >= len(self._filtered):
            cursor = 0
        ridx = self._filtered[cursor]
        return self._resources[ridx].entry

    # ------------------------------------------------------------------
    # Key actions
    # ------------------------------------------------------------------

    def action_cancel(self) -> None:
        self.exit(None)

    def action_select_resource(self) -> None:
        path = self._selected_entry()
        if path is not None:
            self.exit(path)

    def action_navigate_down(self) -> None:
        table = self.query_one("#picker-table", DataTable)
        new_row = min(table.cursor_row + 1, table.row_count - 1)
        table.move_cursor(row=new_row)

    def action_navigate_up(self) -> None:
        table = self.query_one("#picker-table", DataTable)
        new_row = max(table.cursor_row - 1, 0)
        table.move_cursor(row=new_row)

    def action_start_filter(self) -> None:
        bar = self.query_one("#edit-filter-bar")
        bar.add_class("active")
        self._filter_active = True
        self.query_one("#edit-filter-input", Input).focus()

    def key_escape(self) -> None:
        if self._filter_active:
            self._clear_filter()
            self.query_one("#picker-table", DataTable).focus()

    def _clear_filter(self) -> None:
        self.query_one("#edit-filter-bar").remove_class("active")
        self._filter_active = False
        inp = self.query_one("#edit-filter-input", Input)
        inp.value = ""
        self._filter_query = ""
        self._filtered = list(range(len(self._resources)))
        self._rebuild_table()

    # ------------------------------------------------------------------
    # DataTable.RowSelected — fired when Enter is pressed on focused table
    # ------------------------------------------------------------------

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        path = self._selected_entry()
        if path is not None:
            self.exit(path)

    # ------------------------------------------------------------------
    # Filter input events
    # ------------------------------------------------------------------

    @on(Input.Changed, "#edit-filter-input")
    def _on_filter_changed(self, event: Input.Changed) -> None:
        self._filter_query = event.value
        self._filtered = filter_resources(self._resources, event.value, self._ctx)
        self._rebuild_table()

    @on(Input.Submitted, "#edit-filter-input")
    def _on_filter_submitted(self, event: Input.Submitted) -> None:
        self.query_one("#edit-filter-bar").remove_class("active")
        self._filter_active = False
        self.query_one("#picker-table", DataTable).focus()


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def run_show_browser(
    *,
    resources: list[DiscoveredResource],
    ctx: ScopeContext,
    title: str,
) -> int:
    """Launch the two-pane browser. Returns exit code (0 = normal quit)."""
    app = ShowBrowser(resources=resources, ctx=ctx, title=title)
    result = app.run()
    return result if isinstance(result, int) else 0


def run_edit_picker(
    *,
    resources: list[DiscoveredResource],
    ctx: ScopeContext,
    title: str,
) -> Path | None:
    """Launch the edit picker. Returns selected ``Path`` or ``None`` on cancel."""
    app = EditPicker(resources=resources, ctx=ctx, title=title)
    result = app.run()
    return result if isinstance(result, Path) else None
