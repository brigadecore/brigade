SHELL ?= /bin/bash

.DEFAULT_GOAL := build

################################################################################
# Version details                                                              #
################################################################################

# This will reliably return the short SHA1 of HEAD or, if the working directory
# is dirty, will return that + "-dirty"
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	PROJECT_ROOT := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
	GO_DEV_IMAGE := brigadecore/go-tools:v0.1.0

	GO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e GOCACHE=/workspaces/brigade/.gocache \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(GO_DEV_IMAGE)

	JS_DEV_IMAGE := node:12.3.1-stretch

	JS_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(JS_DEV_IMAGE)

	KANIKO_IMAGE := brigadecore/kaniko:v0.2.0

	KANIKO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e DOCKER_USERNAME=$${DOCKER_USERNAME} \
		-e DOCKER_PASSWORD=$${DOCKER_PASSWORD} \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(KANIKO_IMAGE)

	HELM_IMAGE := brigadecore/helm-tools:v0.1.0

	HELM_DOCKER_CMD := docker run \
	  -it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(HELM_IMAGE)
endif

################################################################################
# Binaries and Docker images we build and publish                              #
################################################################################

ifdef DOCKER_REGISTRY
	DOCKER_REGISTRY := $(DOCKER_REGISTRY)/
endif

ifdef DOCKER_ORG
	DOCKER_ORG := $(DOCKER_ORG)/
endif

DOCKER_IMAGE_PREFIX := $(DOCKER_REGISTRY)$(DOCKER_ORG)brigade2-

ifdef VERSION
	MUTABLE_DOCKER_TAG := latest
else
	VERSION            := $(GIT_VERSION)
	MUTABLE_DOCKER_TAG := edge
endif

IMMUTABLE_DOCKER_TAG := $(VERSION)

################################################################################
# Tests                                                                        #
################################################################################

.PHONY: lint-go
lint-go:
	$(GO_DOCKER_CMD) sh -c ' \
		cd sdk/v2 && \
		golangci-lint run --config ../../golangci.yaml && \
		cd ../../v2 && \
		golangci-lint run --config ../golangci.yaml \
	'

.PHONY: test-unit-go
test-unit-go:
	$(GO_DOCKER_CMD) sh -c ' \
		cd sdk/v2 && \
		go test \
			-v \
			-timeout=60s \
			-race \
			-coverprofile=coverage.txt \
			-covermode=atomic \
			./... && \
		cd ../../v2 && \
		go test \
			-v \
			-timeout=60s \
			-race \
			-coverprofile=coverage.txt \
			-covermode=atomic \
			./... \
	'

.PHONY: lint-js
lint-js:
	$(JS_DOCKER_CMD) sh -c ' \
		cd v2/brigadier && \
		yarn install && \
		yarn lint && \
		cd ../worker && \
		yarn install && \
		yarn lint \
	'

.PHONY: test-unit-js
test-unit-js:
	$(JS_DOCKER_CMD) sh -c ' \
		cd v2/brigadier && \
		yarn install && \
		yarn test && \
		cd ../worker && \
		yarn install && \
		yarn test \
	'

.PHONY: lint-chart
lint-chart:
	$(HELM_DOCKER_CMD) sh -c ' \
		cd charts/brigade && \
		helm dep up && \
		helm lint . \
	'

.PHONY: test-integration
test-integration: hack-expose-apiserver
	@cd v2 && \
		go test \
			-v \
			-timeout=10m \
			-tags=integration \
			./tests/... || (cd - && $(MAKE) hack-unexpose-apiserver && exit 1)
	@$(MAKE) hack-unexpose-apiserver

################################################################################
# Build                                                                        #
################################################################################

.PHONY: build
build: build-brigadier build-images build-cli

.PHONY: build-brigadier
build-brigadier:
	$(JS_DOCKER_CMD) sh -c ' \
		cd v2/brigadier && \
		yarn install && \
		yarn build \
	'

.PHONY: build-images
build-images: build-apiserver build-scheduler build-observer build-logger-linux build-git-initializer build-worker

.PHONY: build-logger-linux
build-logger-linux:
	$(KANIKO_DOCKER_CMD) kaniko \
		--dockerfile /workspaces/brigade/v2/logger/Dockerfile.linux \
		--context dir:///workspaces/brigade/logger \
		--no-push

.PHONY: build-%
build-%:
	$(KANIKO_DOCKER_CMD) kaniko \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(GIT_VERSION) \
		--dockerfile /workspaces/brigade/v2/$*/Dockerfile \
		--context dir:///workspaces/brigade/ \
		--no-push

.PHONY: build-cli
build-cli:
	$(GO_DOCKER_CMD) sh -c ' \
		cd v2 && \
		VERSION=$(VERSION) \
		COMMIT=$(GIT_VERSION) \
		../scripts/build-cli.sh \
	'

################################################################################
# Publish                                                                      #
################################################################################

.PHONY: push-images
push-images: push-apiserver push-scheduler push-observer push-logger-linux push-git-initializer push-worker

.PHONY: push-logger-linux
push-logger-linux:
	$(KANIKO_DOCKER_CMD) sh -c ' \
		docker login $(DOCKER_REGISTRY) -u $${DOCKER_USERNAME} -p $${DOCKER_PASSWORD} && \
		kaniko \
			--dockerfile /workspaces/brigade/v2/logger/Dockerfile.linux \
			--context dir:///workspaces/brigade/logger \
			--destination $(DOCKER_IMAGE_PREFIX)logger-linux:$(IMMUTABLE_DOCKER_TAG) \
			--destination $(DOCKER_IMAGE_PREFIX)logger-linux:$(MUTABLE_DOCKER_TAG) \
	'

.PHONY: push-%
push-%:
	$(KANIKO_DOCKER_CMD) sh -c ' \
		docker login $(DOCKER_REGISTRY) -u $${DOCKER_USERNAME} -p $${DOCKER_PASSWORD} && \
		kaniko \
			--build-arg VERSION="$(VERSION)" \
			--build-arg COMMIT="$(GIT_VERSION)" \
			--dockerfile /workspaces/brigade/v2/$*/Dockerfile \
			--context dir:///workspaces/brigade/ \
			--destination $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
			--destination $(DOCKER_IMAGE_PREFIX)$*:$(MUTABLE_DOCKER_TAG) \
	'

################################################################################
# Targets to facilitate hacking on Brigade.                                    #
################################################################################

.PHONY: hack-build-logger-linux
hack-build-logger-linux:
	docker build \
		-f v2/logger/Dockerfile.linux \
		-t $(DOCKER_IMAGE_PREFIX)logger-linux:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		v2/logger

.PHONY: hack-build-%
hack-build-%:
	docker build \
		-f v2/$*/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.

.PHONY: hack-build-cli
hack-build-cli:
	$(GO_DOCKER_CMD) sh -c ' \
		cd v2 && \
		OSES=$(shell go env GOOS) \
		ARCHS=$(shell go env GOARCH) \
		VERSION="$(VERSION)" \
		COMMIT="$(GIT_VERSION)" \
		../scripts/build-cli.sh \
	'

.PHONY: hack-push-images
hack-push-images: hack-push-apiserver hack-push-scheduler hack-push-observer hack-push-logger-linux hack-push-git-initializer hack-push-worker

.PHONY: hack-push-%
hack-push-%: hack-build-%
	docker push $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)

.PHONY: hack
hack: hack-push-images hack-build-cli
	kubectl get namespace brigade || kubectl create namespace brigade
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--set apiserver.image.repository=$(DOCKER_IMAGE_PREFIX)apiserver \
		--set apiserver.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set apiserver.image.pullPolicy=Always \
		--set apiserver.service.type=NodePort \
		--set apiserver.service.nodePort=31600 \
		--set scheduler.image.repository=$(DOCKER_IMAGE_PREFIX)scheduler \
		--set scheduler.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set scheduler.image.pullPolicy=Always \
		--set observer.image.repository=$(DOCKER_IMAGE_PREFIX)observer \
		--set observer.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set observer.image.pullPolicy=Always \
		--set worker.image.repository=$(DOCKER_IMAGE_PREFIX)worker \
		--set worker.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set worker.image.pullPolicy=Always \
		--set gitInitializer.image.repository=$(DOCKER_IMAGE_PREFIX)git-initializer \
		--set gitInitializer.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set gitInitializer.image.pullPolicy=Always \
		--set logger.linux.image.repository=$(DOCKER_IMAGE_PREFIX)logger-linux \
		--set logger.linux.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set logger.linux.image.pullPolicy=Always

.PHONY: hack-expose-apiserver
hack-expose-apiserver:
	@kubectl --namespace brigade port-forward service/brigade-apiserver 7000:443 &>/dev/null & \
		echo $$! > /tmp/brigade-apiserver.PID

.PHONY: hack-unexpose-apiserver
hack-unexpose-apiserver:
	@kill -TERM $$(cat /tmp/brigade-apiserver.PID)
