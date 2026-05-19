.DEFAULT_GOAL := help

REPO_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
DEPLOY_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
export PIIS_REPO_ROOT := $(REPO_ROOT)

.PHONY: help
help: ## Команды деплоя в локальный Kubernetes (Rancher Desktop)
	@grep -E '^[a-zA-Z0-9_.-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

.PHONY: build-images
build-images: ## docker build traffic/*:local + load в k8s.io (containerd)
	bash $(DEPLOY_DIR)/scripts/build-images.sh

.PHONY: load-images-k8s
load-images-k8s: ## docker save → nerdctl -n k8s.io load (если ErrImagePull)
	bash $(DEPLOY_DIR)/scripts/load-images-k8s.sh

.PHONY: sync-ml-artifacts
sync-ml-artifacts: ## Веса ML → .data/artifacts (как infra/Makefile)
	$(MAKE) -C $(REPO_ROOT)/infra sync-ml-artifacts

.PHONY: apply
apply: ## kubectl apply -k deploy/k8s
	bash $(DEPLOY_DIR)/scripts/render-apply.sh

.PHONY: wait
wait: ## Дождаться rollout и init jobs
	bash $(DEPLOY_DIR)/scripts/wait-rollout.sh

.PHONY: deploy
deploy: build-images sync-ml-artifacts apply wait ## Полный деплой

.PHONY: delete
delete: ## Удалить namespace traffic
	kubectl delete namespace traffic --ignore-not-found

.PHONY: pods
pods: ## Статус подов
	kubectl -n traffic get pods -o wide

.PHONY: urls
urls: ## EXTERNAL-IP всех LoadBalancer-сервисов (UI, API, Grafana, …)
	bash $(DEPLOY_DIR)/scripts/print-external-urls.sh traffic
