"""Sanity checks on the harness knowledge base."""

from __future__ import annotations

from agents_cli.harnesses import ALL_HARNESSES, ResourceType, get


def test_every_harness_has_unique_id() -> None:
    ids = [h.id for h in ALL_HARNESSES]
    assert len(ids) == len(set(ids))


def test_get_resolves_aliases() -> None:
    assert get("opencode").id == "opencode"
    assert get("oc").id == "opencode"
    assert get("claude").id == "claude"
    assert get("cc").id == "claude"


def test_every_harness_has_a_guideline() -> None:
    """All supported harnesses must at minimum describe a guideline path."""
    for h in ALL_HARNESSES:
        assert h.guideline is not None, h.id
        assert h.guideline.project_path, h.id


def test_path_templates_have_no_stray_placeholders() -> None:
    for h in ALL_HARNESSES:
        for rtype in ResourceType:
            spec = h.spec(rtype)
            if spec is None:
                continue
            all_paths = [spec.project_path, spec.global_path or ""]
            all_paths.extend(spec.project_aliases)
            all_paths.extend(spec.global_aliases)
            for p in all_paths:
                # Only {name} is allowed as a placeholder.
                assert "{" not in p.replace("{name}", ""), (h.id, rtype, p)
