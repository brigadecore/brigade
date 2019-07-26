SHELL ?= /bin/bash

.DEFAULT_GOAL := build

################################################################################
# Version details                                                              #
################################################################################

# This will reliably return the short SHA1 of HEAD or, if the working directory
# is dirty, will return that + "-dirty"
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty --match=NeVeRmAtCh)

################################################################################
# Go build details                                                             #
################################################################################

BASE_PACKAGE_NAME := github.com/brigadecore/brigade

################################################################################
# Containerized development environment-- or lack thereof                      #
################################################################################

ifneq ($(SKIP_DOCKER),true)
	GO_DEV_IMAGE := quay.io/deis/lightweight-docker-go:v0.7.0
	JS_DEV_IMAGE := node:12.3.1-stretch

	GO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-v $$(pwd):/go/src/$(BASE_PACKAGE_NAME) \
		-w /go/src/$(BASE_PACKAGE_NAME) $(GO_DEV_IMAGE)

	JS_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e KUBECONFIG="/code/$(BASE_PACKAGE_NAME)/brigade-worker/test/fake_kubeconfig.yaml" \
		-v $$(pwd):/code/$(BASE_PACKAGE_NAME) \
		-w /code/$(BASE_PACKAGE_NAME) $(JS_DEV_IMAGE)
endif

# Allow for users to supply a different helm cli name,
# for instance, if one has helm v3 as `helm3` and helm v2 as `helm`
HELM ?= helm

################################################################################
# Binaries and Docker images we build and publish                              #
################################################################################

IMAGES := brigade-api brigade-controller brigade-cr-gateway brigade-generic-gateway brigade-vacuum brig brigade-worker git-sidecar

ifdef DOCKER_REGISTRY
	DOCKER_REGISTRY := $(DOCKER_REGISTRY)/
endif

ifdef DOCKER_ORG
	DOCKER_ORG := $(DOCKER_ORG)/
endif

DOCKER_IMAGE_PREFIX := $(DOCKER_REGISTRY)$(DOCKER_ORG)

ifdef VERSION
	MUTABLE_DOCKER_TAG := latest
else
	VERSION            := $(GIT_VERSION)
	MUTABLE_DOCKER_TAG := edge
endif

LDFLAGS              := -X github.com/brigadecore/brigade/pkg/version.Version=$(VERSION)
IMMUTABLE_DOCKER_TAG := $(VERSION)

################################################################################
# Utility targets                                                              #
################################################################################

.PHONY: dep
dep:
	$(GO_DOCKER_CMD) dep ensure -v

.PHONY: format
format: format-go format-js

.PHONY: format-go
format-go:
	$(GO_DOCKER_CMD) sh -c "go list -f '{{.Dir}}' ./... | xargs goimports -w -local github.com/brigadecore/brigade"

.PHONY: yarn-install
yarn-install:
	$(JS_DOCKER_CMD) sh -c 'cd brigade-worker && yarn install'

.PHONY: format-js
format-js:
	$(JS_DOCKER_CMD) sh -c 'cd brigade-worker && yarn format'

################################################################################
# Tests                                                                        #
################################################################################

# All non-functional tests
.PHONY: test
test: verify-vendored-code lint test-unit verify-vendored-code-js test-js

# Verifies there are no discrepancies between desired dependencies and the
# tracked, vendored dependencies
.PHONY: verify-vendored-code
verify-vendored-code:
	$(GO_DOCKER_CMD) dep check

.PHONY: lint
lint:
	$(GO_DOCKER_CMD) golangci-lint run --config ./golangci.yml

# Unit tests. Local only.
.PHONY: test-unit
test-unit:
	$(GO_DOCKER_CMD) go test -v ./...

# Verifies there are no discrepancies between desired dependencies and the
# tracked, vendored dependencies
.PHONY: verify-vendored-code-js
verify-vendored-code-js:
	$(JS_DOCKER_CMD) sh -c 'cd brigade-worker && yarn check --integrity && yarn check --verify-tree'

# JS test is local only
.PHONY: test-js
test-js:
	$(JS_DOCKER_CMD) sh -c 'cd brigade-worker && yarn build && yarn test'

################################################################################
# Build / Publish                                                              #
################################################################################

build: build-all-images build-brig

.PHONY: build-all-images
build-all-images: $(addsuffix -build-image,$(IMAGES))

%-build-image:
	cp $*/.dockerignore .
	docker build \
		-f $*/Dockerfile \
		-t $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
		--build-arg LDFLAGS='$(LDFLAGS)' \
		.
	docker tag $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) $(DOCKER_IMAGE_PREFIX)$*:$(MUTABLE_DOCKER_TAG)

.PHONY: load-all-images
load-all-images: $(addsuffix -load-image,$(IMAGES))

%-load-image:
	@echo "Loading $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)"
	@kind load docker-image $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
			|| echo >&2 "kind not installed or error loading image: $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)"

# Cross-compile binaries for brig
build-brig:
	$(GO_DOCKER_CMD) bash -c 'LDFLAGS="$(LDFLAGS)" scripts/build-brig.sh'

.PHONY: push
push: push-all-images

# You must be logged into DOCKER_REGISTRY before you can push.
.PHONY: push-all-images
push-all-images: build-all-images
push-all-images: $(addsuffix -push-image,$(IMAGES))

%-push-image:
	docker push $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)$*:$(MUTABLE_DOCKER_TAG)

################################################################################
# Helm and functional test utils                                               #
################################################################################

# Helm chart/release defaults
BRIGADE_RELEASE                 ?= brigade-server
BRIGADE_NAMESPACE               ?= default
BRIGADE_GITHUB_GW_SERVICE       := $(BRIGADE_RELEASE)-brigade-github-app
BRIGADE_GITHUB_GW_INTERNAL_PORT := 80
BRIGADE_GITHUB_GW_EXTERNAL_PORT := 7744
BRIGADE_VERSION := $(VERSION)

# DOCKER_IMAGE_PREFIX would have a trailing slash that we won't want here
# because the charts glue the registry (registry + org, really), image name.
# and tag together themselves and do not account for the possibility of registry
# or org being "". So, we create a new variable here called BRIGADE_REGISTRY
# that is DOCKER_IMAGE_PREFIX minus any trailing slash. If DOCKER_IMAGE_PREFIX
# is the empty string, we default to "brigadecore", since, again, the charts
# cannot currently accommodate a blank registry or org.
ifeq ("$(DOCKER_IMAGE_PREFIX)","")
	BRIGADE_REGISTRY := brigadecore
else
	BRIGADE_REGISTRY := $(shell echo $(DOCKER_IMAGE_PREFIX) | sed 's:/*$$::')
endif

.PHONY: helm-install
helm-install: helm-upgrade

.PHONY: helm-upgrade
helm-upgrade:
	$(HELM) upgrade --install $(BRIGADE_RELEASE) brigade/brigade --namespace $(BRIGADE_NAMESPACE) \
		--set brigade-github-app.enabled=true \
		--set controller.registry=$(BRIGADE_REGISTRY) \
		--set controller.tag=$(BRIGADE_VERSION) \
		--set api.registry=$(BRIGADE_REGISTRY) \
		--set api.tag=$(BRIGADE_VERSION) \
		--set worker.registry=$(BRIGADE_REGISTRY) \
		--set worker.tag=$(BRIGADE_VERSION) \
		--set cr.enabled=true \
		--set cr.registry=$(BRIGADE_REGISTRY) \
		--set cr.tag=$(BRIGADE_VERSION) \
		--set genericGateway.enabled=true \
		--set genericGateway.registry=$(BRIGADE_REGISTRY) \
		--set genericGateway.tag=$(BRIGADE_VERSION) \
		--set vacuum.registry=$(BRIGADE_REGISTRY) \
		--set vacuum.tag=$(BRIGADE_VERSION)

# Functional tests assume access to github.com
# and Brigade chart installed with `--set brigade-github-app.enabled=true`
#
# To set this up in your local environment:
# - Make sure kubectl is pointed to the right cluster
# - Run "helm repo add brigade https://brigadecore.github.io/charts"
# - Run "helm inspect values brigade/brigade-project > myvals.yaml"
# - Set the values in myvalues.yaml to something like this:
#   project: "brigadecore/empty-testbed"
#   repository: "github.com/brigadecore/empty-testbed"
#   secret: "MySecret"
# - Run "helm install brigade/brigade-project -f myvals.yaml"
# - Run "make test-functional"
#
# This will clone the github.com/brigadecore/empty-testbed repo and run the brigade.js
# file found there.
# Test Repo is https://github.com/brigadecore/empty-testbed
TEST_REPO_COMMIT =  589e15029e1e44dee48de4800daf1f78e64287c0
KUBECONFIG       ?= ${HOME}/.kube/config
.PHONY: test-functional
test-functional:
	@kubectl port-forward service/$(BRIGADE_GITHUB_GW_SERVICE) $(BRIGADE_GITHUB_GW_EXTERNAL_PORT):80 &>/dev/null & \
		echo $$! > /tmp/$(BRIGADE_GITHUB_GW_SERVICE).PID
	go test --tags integration ./tests -kubeconfig $(KUBECONFIG) $(TEST_REPO_COMMIT)
	@kill -TERM $$(cat /tmp/$(BRIGADE_GITHUB_GW_SERVICE).PID)

.PHONY: e2e
e2e:
	./e2e/run.sh