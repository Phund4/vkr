.DEFAULT_GOAL := help

# Каталог infra/ (где лежит этот Makefile и docker-compose.yml).
INFRA_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
# Корень монорепозитория.
REPO_ROOT := $(abspath $(INFRA_DIR)/..)

COMPOSE ?= docker compose

# Сервисы из ../services/ (профиль apps). Для up нужны и infra, и apps — иначе depends_on
# на kafka-topics-init, postgres и т.д. даёт «undefined service».
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
infra-up: ## Поднять инфраструктуру (postgres, kafka, clickhouse, minio, ELK, prometheus, …)
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra up -d

.PHONY: infra-down
infra-down: ## Остановить сервисы профиля infra
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra down

.PHONY: infra-logs
infra-logs: ## Логи профиля infra
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra logs -f

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
sync-ml-artifacts: ## Скопировать best.pt из $(REPO_ROOT)/artifacts в .data/artifacts для volume ml-serving
	@set -e; \
	a="$(REPO_ROOT)/artifacts/accident/baseline-cnn/best.pt"; \
	c="$(REPO_ROOT)/artifacts/congestion/tiny-cnn/best.pt"; \
	test -f "$$a" || (echo "sync-ml-artifacts: нет файла $$a" >&2; exit 1); \
	test -f "$$c" || (echo "sync-ml-artifacts: нет файла $$c" >&2; exit 1); \
	mkdir -p "$(REPO_ROOT)/.data/artifacts/accident/baseline-cnn" "$(REPO_ROOT)/.data/artifacts/congestion/tiny-cnn"; \
	cp -f "$$a" "$(REPO_ROOT)/.data/artifacts/accident/baseline-cnn/"; \
	cp -f "$$c" "$(REPO_ROOT)/.data/artifacts/congestion/tiny-cnn/"; \
	echo "sync-ml-artifacts: OK → $(REPO_ROOT)/.data/artifacts"

.PHONY: stack-up
stack-up: ## Весь стек: infra + apps (--build для приложений)
	cd $(INFRA_DIR) && $(COMPOSE) --profile infra --profile apps up -d --build

.PHONY: stack-down
stack-down: ## Остановить весь compose-проект
	cd $(INFRA_DIR) && $(COMPOSE) down

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
