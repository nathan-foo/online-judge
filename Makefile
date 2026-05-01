include .env
export BASE_URL

.DEFAULT_GOAL := setup

COMPOSE := docker compose

.PHONY: setup client-setup python-setup docker-up docker-down docker-ngrok build logs client

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

docker-up:
	$(COMPOSE) up -d --build

docker-down:
	$(COMPOSE) down

docker-ngrok:
	$(COMPOSE) up -d --build && ngrok http --url=$(BASE_URL) 8080

build:
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f

client:
	cd client && npm run dev
