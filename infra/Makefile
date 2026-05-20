.DEFAULT_GOAL := help

# Каталог infra/ (где лежит этот Makefile и docker-compose.yml).
INFRA_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# Корень монорепозитория.
REPO_ROOT := $(abspath $(INFRA_DIR)/..)

COMPOSE ?= docker compose

# Профиль observability: ELK + Prometheus + Grafana (без infra; таргеты экспортёров в Prometheus могут быть DOWN).
OBSERVABILITY_SERVICES := elasticsearch logstash kibana filebeat prometheus grafana

# Сервисы из ../services/ (профиль apps). Зависимости (kafka, postgres, …) тоже в профиле apps.
APPS_SERVICES := coordinator analytics pusher data-service ml-serving router

# Go-модули: абсолютные пути (lint использует ./.golangci.yml в каждом модуле).
GO_MODULES := \
	$(REPO_ROOT)/services/router \
	$(REPO_ROOT)/services/analytics \
	$(REPO_ROOT)/services/coordinator \
	$(REPO_ROOT)/services/pusher \
	$(REPO_ROOT)/services/data-service \
	$(INFRA_DIR)rtsp-generator

.PHONY: help
help: ## Список команд (make из каталога infra или: make -f infra/Makefile из корня репо)
	@grep -E '^[a-zA-Z0-9_.-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

.PHONY: infra-up
infra-up: ## Ядро infra (как в K8s): postgres, kafka, clickhouse, minio, mediamtx, rtsp-generator
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra up -d

.PHONY: infra-down
infra-down: ## Остановить сервисы профиля infra
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra down

.PHONY: infra-logs
infra-logs: ## Логи профиля infra
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra logs -f

.PHONY: observability-up
observability-up: ## Только observability: ELK + Prometheus + Grafana (таргеты compose)
	cd $(INFRA_DIR) && $(COMPOSE) --profile observability up -d $(OBSERVABILITY_SERVICES)

.PHONY: observability-down
observability-down: ## Остановить observability (infra не трогаем)
	cd $(INFRA_DIR) && $(COMPOSE) --profile observability stop $(OBSERVABILITY_SERVICES)

.PHONY: observability-logs
observability-logs: ## Логи observability
	cd $(INFRA_DIR) && $(COMPOSE) --profile observability logs -f $(OBSERVABILITY_SERVICES)

.PHONY: apps-up
apps-up: ## Поднять только приложения из services/ + их зависимости (оба профиля в графе compose)
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra --profile apps up -d --build $(APPS_SERVICES)

.PHONY: apps-down
apps-down: ## Остановить контейнеры приложений (infra не трогаем)
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra --profile apps stop $(APPS_SERVICES)

.PHONY: apps-logs
apps-logs: ## Логи приложений из services/
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra --profile apps logs -f $(APPS_SERVICES)

.PHONY: sync-ml-artifacts
sync-ml-artifacts: ## best.pt → .data/artifacts (источник: artifacts/, .data/artifacts или services/ml-serving/artifacts/)
	@set -e; \
	ROOT="$(REPO_ROOT)"; \
	DACC="$$ROOT/.data/artifacts/accident/baseline-cnn/best.pt"; \
	DCONG="$$ROOT/.data/artifacts/congestion/tiny-cnn/best.pt"; \
	if test -f "$$DACC" && test -f "$$DCONG"; then \
		echo "sync-ml-artifacts: уже есть в $$ROOT/.data/artifacts"; \
		exit 0; \
	fi; \
	mkdir -p "$$ROOT/.data/artifacts/accident/baseline-cnn" "$$ROOT/.data/artifacts/congestion/tiny-cnn"; \
	acc=""; cong=""; \
	for base in "$$ROOT/artifacts" "$$ROOT/services/ml-serving/artifacts"; do \
		if test -z "$$acc" && test -f "$$base/accident/baseline-cnn/best.pt"; then acc="$$base/accident/baseline-cnn/best.pt"; fi; \
		if test -z "$$cong" && test -f "$$base/congestion/tiny-cnn/best.pt"; then cong="$$base/congestion/tiny-cnn/best.pt"; fi; \
	done; \
	if test -z "$$acc" || test -z "$$cong"; then \
		echo "sync-ml-artifacts: нужны оба best.pt. Положите в один из каталогов:" >&2; \
		echo "  $$ROOT/artifacts/accident/baseline-cnn/best.pt" >&2; \
		echo "  $$ROOT/artifacts/congestion/tiny-cnn/best.pt" >&2; \
		echo "  (или уже в $$ROOT/.data/artifacts/...)" >&2; \
		exit 1; \
	fi; \
	cp -f "$$acc" "$$DACC"; cp -f "$$cong" "$$DCONG"; \
	echo "sync-ml-artifacts: OK ← $$(dirname "$$acc")/... → $$ROOT/.data/artifacts"

.PHONY: stack-up
stack-up: ## Контур как в K8s: infra + apps (--build для приложений, без observability)
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra --profile apps up -d --build

.PHONY: stack-full-up
stack-full-up: ## infra + observability + apps (--build)
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra --profile observability --profile apps up -d --build

.PHONY: stack-down
stack-down: ## Остановить весь compose-проект (все профили)
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra --profile observability --profile apps down

.PHONY: stack-logs
stack-logs: ## Логи всех контейнеров проекта
	cd $(INFRA_DIR) && $(COMPOSE) logs -f

.PHONY: compose-ps
compose-ps: ## Статус контейнеров
	cd $(INFRA_DIR) && $(COMPOSE) ps -a

.PHONY: lint-go
lint-go: ## golangci-lint в каждом Go-сервисе (локальный .golangci.yml)
	@set -e; for m in $(GO_MODULES); do \
		echo "==> $$m"; \
		( cd "$$m" && golangci-lint run ./... ); \
	done

.PHONY: test-go
test-go: ## go test во всех модулях
	@set -e; for m in $(GO_MODULES); do \
		echo "==> $$m"; \
		( cd "$$m" && go test ./... ); \
	done

.PHONY: fmt-go
fmt-go: ## go fmt во всех модулях
	@set -e; for m in $(GO_MODULES); do ( cd "$$m" && go fmt ./... ); done

.PHONY: build-go
build-go: ## go build ./... во всех модулях
	@set -e; for m in $(GO_MODULES); do \
		echo "==> $$m"; \
		( cd "$$m" && go build ./... ); \
	done

.PHONY: install-deps
install-deps: ## Локальные инструменты (lint, swag, …)
	-go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	-go install github.com/swaggo/swag/cmd/swag@latest
	-go install golang.org/x/tools/cmd/goimports@latest
