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
	# https://github.com/krancour/go-tools
	# https://hub.docker.com/repository/docker/krancour/go-tools
	GO_DEV_IMAGE := krancour/go-tools:v0.4.0

	GO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e GOCACHE=/workspaces/brigade/.gocache \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(GO_DEV_IMAGE)
endif

# Allow for users to supply a different helm cli name,
# for instance, if one has helm v3 as `helm3` and helm v2 as `helm`
HELM ?= helm

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

################################################################################
# Build / Publish                                                              #
################################################################################

.PHONY: build
build: build-images xbuild-cli

.PHONY: build-images
build-images: build-apiserver build-scheduler build-observer

.PHONY: build-logger-linux
build-logger-linux:
	docker build \
		-t $(DOCKER_IMAGE_PREFIX)logger-linux:$(IMMUTABLE_DOCKER_TAG) \
		- < v2/logger/Dockerfile.linux
	docker tag $(DOCKER_IMAGE_PREFIX)logger-linux:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)logger-linux:$(MUTABLE_DOCKER_TAG)

.PHONY: build-%
build-%:
	docker build \
		-f v2/$*/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg VERSION='$(VERSION)' \
		--build-arg COMMIT='$(GIT_VERSION)' \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)$*:$(MUTABLE_DOCKER_TAG)

.PHONY: build-cli
build-cli:
	$(GO_DOCKER_CMD) sh -c ' \
		cd v2 && \
		OSES=$(shell go env GOOS) \
		ARCHS=$(shell go env GOARCH) \
		VERSION="$(VERSION)" \
		COMMIT="$(GIT_VERSION)" \
		../scripts/build-cli.sh \
	'

.PHONY: xbuild-cli
xbuild-cli:
	$(GO_DOCKER_CMD) sh -c ' \
		cd v2 && \
		VERSION="$(VERSION)" \
		COMMIT="$(GIT_VERSION)" \
		../scripts/build-cli.sh \
	'

.PHONY: push-images
push-images: push-apiserver push-scheduler push-observer push-logger-linux

.PHONY: push-%
push-%: build-%
	docker push $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)$*:$(MUTABLE_DOCKER_TAG)

################################################################################
# Temporary hacks to facilitate early development.                             #
# Longer term, we'll devise an improved workflow.                              #
################################################################################

.PHONY: hack
hack: push-images build-cli
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
		--set logger.linux.image.repository=$(DOCKER_IMAGE_PREFIX)logger-linux \
		--set logger.linux.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set logger.linux.image.pullPolicy=Always
