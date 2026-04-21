.PHONY: install dev run test lint format typecheck check clean build tool-install tool-upgrade tool-uninstall

install:
	uv sync

dev:
	uv sync --all-groups

run:
	uv run agents

test:
	uv run pytest

lint:
	uv run ruff check src tests

format:
	uv run ruff format src tests
	uv run ruff check --fix src tests

typecheck:
	uv run mypy src

check: lint typecheck test

clean:
	rm -rf build dist *.egg-info .pytest_cache .mypy_cache .ruff_cache

build:
	uv build

tool-install:
	uv tool install --force --from . agents-cli

tool-upgrade:
	uv tool upgrade agents-cli

tool-uninstall:
	uv tool uninstall agents-cli
