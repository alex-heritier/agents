"""Templates for new resources.

When the user runs ``agents skill new my-skill``, we drop a correctly
shaped file so they can start editing immediately. Templates are minimal
and idiomatic to each resource type.
"""

from __future__ import annotations

from textwrap import dedent

from agents_cli.harnesses import Format, Harness, ResourceType


def _slugify_title(name: str) -> str:
    parts = name.replace("-", " ").replace("_", " ").split()
    return " ".join(p.capitalize() for p in parts) if parts else name


def guideline(name: str) -> str:
    return dedent(
        f"""\
        # {_slugify_title(name) or "Agent Guidelines"}

        Describe how agents should behave in this repo / directory.

        ## Project context

        A one-paragraph orientation: what this codebase is, the primary
        language(s), and any high-level architectural notes worth knowing
        before editing anything.

        ## Commands

        - `make test` — run the test suite.
        - `make lint` — run linters.
        - `make format` — auto-format.

        ## Conventions

        - Add bullet-point rules here (style, naming, testing expectations).
        - Keep this file short; link out rather than inlining docs.
        """
    )


def skill(name: str) -> str:
    return dedent(
        f"""\
        ---
        name: {name}
        description: One-sentence summary of when to use this skill.
        ---

        # {_slugify_title(name)}

        Describe the skill's purpose and scope in 1-3 paragraphs.

        ## When to use

        - Bullet list of specific triggers.

        ## How to use

        1. Step-by-step instructions for the agent.
        2. Include any required commands, files, or references.

        ## Notes

        Additional context, edge cases, and caveats.
        """
    )


def command_markdown(name: str) -> str:
    return dedent(
        f"""\
        ---
        description: One-line description shown in the slash menu.
        argument-hint: "<optional hint about args>"
        ---

        # /{name}

        Instructions for the agent when this command is invoked.

        Use `$ARGUMENTS` to reference user input. You can also embed
        file contents with `@path/to/file.md` and run shell commands
        with `!`<cmd>`` (harness-dependent).
        """
    )


def command_toml(name: str) -> str:
    return dedent(
        f"""\
        description = "One-line description shown in the slash menu."
        prompt = \"\"\"
        Instructions for the agent when /{name} is invoked.

        Use {{{{args}}}} to reference user input.
        \"\"\"
        """
    )


def subagent(name: str) -> str:
    return dedent(
        f"""\
        ---
        name: {name}
        description: Specialized subagent for <task>. Use proactively when <trigger>.
        model: inherit
        ---

        You are a specialized subagent focused on <task>.

        ## Responsibilities

        - Bullet list of what this subagent does.

        ## Constraints

        - What the subagent should avoid.
        """
    )


def render(rtype: ResourceType, harness: Harness, name: str) -> str:
    if rtype is ResourceType.GUIDELINE:
        return guideline(name)
    if rtype is ResourceType.SKILL:
        return skill(name)
    if rtype is ResourceType.COMMAND:
        fmt = harness.command.format if harness.command is not None else Format.MARKDOWN
        return command_markdown(name) if fmt is Format.MARKDOWN else command_toml(name)
    if rtype is ResourceType.SUBAGENT:
        return subagent(name)
    raise ValueError(f"no template for {rtype}")
