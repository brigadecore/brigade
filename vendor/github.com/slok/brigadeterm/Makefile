

# Name of this service/application
SERVICE_NAME := brigadeterm

# Path of the go service inside docker
DOCKER_GO_SERVICE_PATH := /src

# Shell to use for running scripts
SHELL := $(shell which bash)

# Get OS
OSTYPE := $(shell uname)

# Get docker path or an empty string
DOCKER := $(shell command -v docker)

# Get the main unix group for the user running make (to be used by docker-compose later)
GID := $(shell id -g)

# Get the unix user id for the user running make (to be used by docker-compose later)
UID := $(shell id -u)

# Version from Git.
VERSION=$(shell git describe --tags --always)

# cmds
UNIT_TEST_CMD := ./hack/scripts/unit-test.sh
INTEGRATION_TEST_CMD := ./hack/scripts/integration-test.sh
MOCKS_CMD := ./hack/scripts/mockgen.sh
DOCKER_RUN_CMD := docker run --env ostype=$(OSTYPE) -v ${PWD}:$(DOCKER_GO_SERVICE_PATH) --rm -it $(SERVICE_NAME)
BUILD_BINARY_CMD := VERSION=${VERSION} ./hack/scripts/build.sh
BUILD_IMAGE_CMD := IMAGE_VERSION=${VERSION} ./hack/scripts/build-image.sh
CI_RELEASE_CMD := ./hack/scripts/travis-release.sh
DEPS_CMD := GO111MODULE=on go mod tidy && GO111MODULE=on go mod vendor
K8S_VERSION := 1.10.9
SET_K8S_DEPS_CMD := GO111MODULE=on go mod edit \
	-require=k8s.io/client-go@kubernetes-${K8S_VERSION} \
	-require=k8s.io/apimachinery@kubernetes-${K8S_VERSION} \
	-require=k8s.io/api@kubernetes-${K8S_VERSION} &&  \
	$(DEPS_CMD)

# environment dirs
DEV_DIR := docker/dev

IMAGE := $(shell docker images -q brigadeterm 2> /dev/null)

# The default action of this Makefile is to build the development docker image
.PHONY: default
default: build

# Test if the dependencies we need to run this Makefile are installed
.PHONY: deps-development
deps-development:
ifndef DOCKER
	@echo "Docker is not available. Please install docker"
	@exit 1
endif
ifndef IMAGE
	@echo "Docker Image not available, Building..."
	docker build -t $(SERVICE_NAME) \
	--build-arg uid=$(UID) \
	--build-arg gid=$(GID) \
	--build-arg ostype=$(OSTYPE) -f ./docker/dev/Dockerfile .
endif

# Build the development docker image
.PHONY: build
build:
	docker build -t $(SERVICE_NAME) \
	--build-arg uid=$(UID) \
	--build-arg gid=$(GID) \
	--build-arg ostype=$(OSTYPE) -f ./docker/dev/Dockerfile .

# Shell the development docker image
.PHONY: build
shell: build
	$(DOCKER_RUN_CMD) /bin/bash



# Build production stuff.
build-binary: deps-development
	$(DOCKER_RUN_CMD) /bin/sh -c '$(BUILD_BINARY_CMD)'

.PHONY: build-image
build-image:
	$(BUILD_IMAGE_CMD)

# Test stuff in dev
.PHONY: unit-test
unit-test: build
	$(DOCKER_RUN_CMD) /bin/sh -c '$(UNIT_TEST_CMD)'
.PHONY: integration-test
integration-test: build
	$(DOCKER_RUN_CMD) /bin/sh -c '$(INTEGRATION_TEST_CMD)'
.PHONY: test
test: integration-test

# Test stuff in ci
.PHONY: ci-unit-test
ci-unit-test:
	$(UNIT_TEST_CMD)
.PHONY: ci-integration-test
ci-integration-test:
	$(INTEGRATION_TEST_CMD)
.PHONY: ci
ci: ci-integration-test

.PHONY: ci-release
ci-release:
	$(CI_RELEASE_CMD)

# Mocks stuff in dev
.PHONY: mocks
mocks: build
	#$(DOCKER_RUN_CMD) /bin/sh -c '$(MOCKS_CMD)'
	$(MOCKS_CMD)

# Dependencies stuff.
.PHONY: set-k8s-deps
set-k8s-deps:
	$(SET_K8S_DEPS_CMD)

.PHONY: deps
deps:
	$(DEPS_CMD)
