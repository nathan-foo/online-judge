include .env
export BASE_URL

.DEFAULT_GOAL := setup

COMPOSE := docker compose

.PHONY: setup client-setup python-setup docker-up docker-down docker-ngrok k8s-up k8s-down k8s-ngrok k8s-logs k8s-destroy build logs client

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

secrets/rabbitmq.conf: secrets/rabbitmq_password.txt
	printf 'default_user = online_judge\ndefault_pass = %s\n' "$$(cat $<)" > $@

docker-up: secrets/rabbitmq.conf
	$(COMPOSE) up -d --build

docker-down:
	$(COMPOSE) down

docker-ngrok: secrets/rabbitmq.conf
	$(COMPOSE) up -d --build && ngrok http --url=$(BASE_URL) 8080

define k8s_deploy
	minikube start
	eval $$(minikube docker-env) && \
		docker build -t online-judge/gateway:dev ./gateway && \
		docker build -t online-judge/user-service:dev ./services/user-service && \
		docker build -t online-judge/quiz-service:dev ./services/quiz-service && \
		docker build -t online-judge/attempt-service:dev ./services/attempt-service && \
		docker build -t online-judge/client:dev \
			--build-arg NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=$(NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY) ./client
	kubectl kustomize --load-restrictor LoadRestrictionsNone k8s/overlays/local | kubectl apply -f -
	kubectl rollout status statefulset/postgres -n online-judge --timeout=120s
	kubectl rollout status statefulset/rabbitmq -n online-judge --timeout=120s
	kubectl rollout status deployment/redis -n online-judge --timeout=120s
	kubectl rollout status deployment/user-service -n online-judge --timeout=120s
	kubectl rollout status deployment/quiz-service -n online-judge --timeout=120s
	kubectl rollout status deployment/attempt-service -n online-judge --timeout=120s
	kubectl rollout status deployment/gateway -n online-judge --timeout=120s
	kubectl rollout status deployment/client -n online-judge --timeout=120s
endef

k8s-up: secrets/rabbitmq.conf
	$(k8s_deploy)
	kubectl port-forward -n online-judge svc/client 3000:3000 & \
		kubectl port-forward -n online-judge svc/gateway 8080:8080 & \
		trap 'kill $$(jobs -p) 2>/dev/null' EXIT INT TERM; \
		wait

k8s-down:
	kubectl kustomize --load-restrictor LoadRestrictionsNone k8s/overlays/local | kubectl delete --ignore-not-found -f -

k8s-ngrok: secrets/rabbitmq.conf
	$(k8s_deploy)
	kubectl port-forward -n online-judge svc/client 3000:3000 & \
		kubectl port-forward -n online-judge svc/gateway 8080:8080 & \
		trap 'kill $$(jobs -p) 2>/dev/null' EXIT INT TERM; \
		ngrok http --url=$(BASE_URL) 8080

k8s-logs:
	kubectl logs -n online-judge -l app --all-containers --prefix -f --tail=100 --max-log-requests=20

k8s-destroy:
	minikube delete

build:
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f

client:
	cd client && npm run dev
