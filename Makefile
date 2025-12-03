list:
	@echo "base.md"
	@echo "ui.md"

tree:
	@echo "./ -- base.md"
	@echo "|-- AGENTS.md"
	@echo "|-- CLAUDE.md"
	
	@echo "./src/frontend/ -- ui.md"
	@echo "|-- AGENTS.md"
	@echo "|-- CLAUDE.md"

help:
	@echo "agents CLI"
	@echo
	@echo "Usage: agents [options] [command]"
	@echo
	@echo "Commands:"
	@echo
	@echo " list"
	@echo " tree"
