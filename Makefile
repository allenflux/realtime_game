SHELL := /bin/bash

TAG ?= latest
REGISTRY ?= flxu
IMAGE_NAME ?= realtime_game

FRONTEND_IMAGE := $(REGISTRY)/$(IMAGE_NAME):frontend-$(TAG)
API_IMAGE := $(REGISTRY)/$(IMAGE_NAME):api-$(TAG)
WORKER_IMAGE := $(REGISTRY)/$(IMAGE_NAME):worker-$(TAG)

build-frontend:
	docker build --target frontend -t $(FRONTEND_IMAGE) .

build-api:
	docker build --target api -t $(API_IMAGE) .

build-worker:
	docker build --target worker -t $(WORKER_IMAGE) .

build-all: build-frontend build-api build-worker

push-frontend:
	docker push $(FRONTEND_IMAGE)

push-api:
	docker push $(API_IMAGE)

push-worker:
	docker push $(WORKER_IMAGE)

push-all: push-frontend push-api push-worker

release-frontend: build-frontend push-frontend
release-api: build-api push-api
release-worker: build-worker push-worker

release-all: build-all push-all

print:
	@echo "FRONTEND_IMAGE=$(FRONTEND_IMAGE)"
	@echo "API_IMAGE=$(API_IMAGE)"
	@echo "WORKER_IMAGE=$(WORKER_IMAGE)"

.PHONY: \
	build-frontend build-api build-worker build-all \
	push-frontend push-api push-worker push-all \
	release-frontend release-api release-worker release-all \
	print