.DEFAULT_GOAL := setup

COMPOSE := docker compose

.PHONY: setup client-setup python-setup up down build logs client

setup: client-setup python-setup

client-setup:
	cd client && npm install && npm run build

python-setup:
	@set -e; \
	for req in services/*/requirements.txt; do \
		[ -f "$$req" ] || continue; \
		dir=$${req%/requirements.txt}; \
		echo "Setting up $$dir"; \
		python3 -m venv "$$dir/.venv"; \
		"$$dir/.venv/bin/python" -m pip install -r "$$req"; \
	done

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

build:
	${COMPOSE} build

logs:
	${COMPOSE} logs -f

client:
	cd client && npm run dev
