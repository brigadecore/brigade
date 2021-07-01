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
		-e GITHUB_TOKEN=$${GITHUB_TOKEN} \
		-e GOCACHE=/workspaces/brigade/.gocache \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(GO_DEV_IMAGE)

	JS_DEV_IMAGE := node:14.16.0-stretch

	JS_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e NPM_TOKEN=$${NPM_TOKEN} \
		-e SKIP_DOCKER=true \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(JS_DEV_IMAGE)

	KANIKO_IMAGE := brigadecore/kaniko:v0.2.0

	KANIKO_DOCKER_CMD := docker run \
		-it \
		--rm \
		-e SKIP_DOCKER=true \
		-e DOCKER_PASSWORD=$${DOCKER_PASSWORD} \
		-v $(PROJECT_ROOT):/workspaces/brigade \
		-w /workspaces/brigade \
		$(KANIKO_IMAGE)

	HELM_IMAGE := brigadecore/helm-tools:v0.1.0

	HELM_DOCKER_CMD := docker run \
	  -it \
		--rm \
		-e SKIP_DOCKER=true \
		-e HELM_PASSWORD=$${HELM_PASSWORD} \
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

ifdef HELM_REGISTRY
	HELM_REGISTRY := $(HELM_REGISTRY)/
endif

ifdef HELM_ORG
	HELM_ORG := $(HELM_ORG)/
endif

HELM_CHART_PREFIX := $(HELM_REGISTRY)$(HELM_ORG)

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
		cd ../brigadier-polyfill && \
		yarn install && \
		yarn lint && \
		cd ../worker && \
		yarn install && \
		yarn lint \
	'

.PHONY: yarn-audit
yarn-audit:
	$(JS_DOCKER_CMD) sh -c ' \
		cd v2/brigadier && \
		yarn install && \
		yarn audit && \
		cd ../brigadier-polyfill && \
		yarn install && \
		yarn audit && \
		cd ../worker && \
		yarn install && \
		yarn audit \
	'

.PHONY: clean-js
clean-js:
	$(JS_DOCKER_CMD) sh -c ' \
		rm -rf \
			v2/brigadier/dist \
			v2/brigadier/node_modules \
			v2/brigadier-polyfill/dist \
			v2/brigadier-polyfill/node_modules \
			v2/worker/dist \
			v2/worker/node_modules \
	'

.PHONY: test-unit-js
test-unit-js:
	$(JS_DOCKER_CMD) sh -c ' \
		cd v2/brigadier && \
		yarn install && \
		yarn test && \
		yarn build && \
		cd ../brigadier-polyfill && \
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

APISERVER_ADDRESS ?= "https://localhost:7000"
APISERVER_ROOT_PASSWORD ?= "F00Bar!!!"

.PHONY: test-integration
test-integration: hack-expose-apiserver
	@export VERSION="$(VERSION)" \
          APISERVER_ADDRESS="$(APISERVER_ADDRESS)" \
          APISERVER_ROOT_PASSWORD="$(APISERVER_ROOT_PASSWORD)" && \
		cd v2 && \
		go test \
			-v \
			--count=1 \
			-timeout=10m \
			-tags=integration \
			./tests/... || (cd - && $(MAKE) hack-unexpose-apiserver && exit 1)
	@$(MAKE) hack-unexpose-apiserver

# Validates the schemas in the v2/apiserver/schemas dir
#
# Adds references to any schema that are themselves $ref'd
# in any of the others.
.PHONY: validate-schemas
validate-schemas:
	$(JS_DOCKER_CMD) sh -c ' \
		npm install -g ajv-cli@3.3.0 && \
		for schema in $$(ls v2/apiserver/schemas/*.json); do \
			ajv compile -s $$schema \
				-r v2/apiserver/schemas/common.json \
				-r v2/apiserver/schemas/source-state.json ; \
		done \
	'

# Validates the examples in the examples/ dir
#
# Currently, they are project-specific;
# we can add event, job, etc. examples and add validation here.
.PHONY: validate-examples
validate-examples:
	$(JS_DOCKER_CMD) sh -c ' \
		npm install -g ajv-cli@3.3.0 && \
		echo "Validating example projects..." && \
		for project in $$(ls examples/*/project.yaml); do \
			ajv validate -d $$project \
				-s v2/apiserver/schemas/project.json \
				-r v2/apiserver/schemas/common.json ; \
		done \
	'

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

# Using docker directly here (kaniko doesn't support Windows container builds)
# and this does require a Windows build machine, but we're placing this here
# to make use of the image prefix and tag logic.
# See the GitHub action(s) yaml in .github/ for use in CI.
.PHONY: build-logger-windows
build-logger-windows:
	docker build \
		-f v2/logger/Dockerfile.winserv-2019 \
		-t $(DOCKER_IMAGE_PREFIX)logger-windows:$(IMMUTABLE_DOCKER_TAG) \
		-t $(DOCKER_IMAGE_PREFIX)logger-windows:$(MUTABLE_DOCKER_TAG) v2/logger

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

.PHONY: publish
publish: publish-brigadier push-images publish-chart publish-cli

.PHONY: publish-brigadier
publish-brigadier: build-brigadier
	$(JS_DOCKER_CMD) sh -c ' \
		cd v2/brigadier && \
		echo '//registry.npmjs.org/:_authToken=${NPM_TOKEN}' > .npmrc && \
		yarn publish \
			--new-version $$(printf $(VERSION) | cut -c 2- ) \
			--access public \
			--no-git-tag-version \
	'

.PHONY: push-images
push-images: push-apiserver push-scheduler push-observer push-logger-linux push-git-initializer push-worker

.PHONY: push-logger-linux
push-logger-linux:
	$(KANIKO_DOCKER_CMD) sh -c ' \
		docker login $(DOCKER_REGISTRY) -u $(DOCKER_USERNAME) -p $${DOCKER_PASSWORD} && \
		kaniko \
			--dockerfile /workspaces/brigade/v2/logger/Dockerfile.linux \
			--context dir:///workspaces/brigade/logger \
			--destination $(DOCKER_IMAGE_PREFIX)logger-linux:$(IMMUTABLE_DOCKER_TAG) \
			--destination $(DOCKER_IMAGE_PREFIX)logger-linux:$(MUTABLE_DOCKER_TAG) \
	'

.PHONY: push-logger-windows
push-logger-windows:
	docker push $(DOCKER_IMAGE_PREFIX)logger-windows:$(IMMUTABLE_DOCKER_TAG)
	docker push $(DOCKER_IMAGE_PREFIX)logger-windows:$(MUTABLE_DOCKER_TAG)

.PHONY: push-%
push-%:
	$(KANIKO_DOCKER_CMD) sh -c ' \
		docker login $(DOCKER_REGISTRY) -u $(DOCKER_USERNAME) -p $${DOCKER_PASSWORD} && \
		kaniko \
			--build-arg VERSION="$(VERSION)" \
			--build-arg COMMIT="$(GIT_VERSION)" \
			--dockerfile /workspaces/brigade/v2/$*/Dockerfile \
			--context dir:///workspaces/brigade/ \
			--destination $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
			--destination $(DOCKER_IMAGE_PREFIX)$*:$(MUTABLE_DOCKER_TAG) \
	'

.PHONY: publish-chart
publish-chart:
	$(HELM_DOCKER_CMD) sh	-c ' \
		helm registry login $(HELM_REGISTRY) -u $(HELM_USERNAME) -p $${HELM_PASSWORD} && \
		cd charts/brigade && \
		helm dep up && \
		sed -i "s/^version:.*/version: $(VERSION)/" Chart.yaml && \
		sed -i "s/^appVersion:.*/appVersion: $(VERSION)/" Chart.yaml && \
		helm chart save . $(HELM_CHART_PREFIX)brigade:$(VERSION) && \
		helm chart push $(HELM_CHART_PREFIX)brigade:$(VERSION) \
	'

.PHONY: publish-cli
publish-cli: build-cli
	$(GO_DOCKER_CMD) sh -c ' \
		go get github.com/tcnksm/ghr && \
		ghr \
			-u $(GITHUB_ORG) \
			-r $(GITHUB_REPO) \
			-c $$(git rev-parse HEAD) \
			-t $${GITHUB_TOKEN} \
			-n ${VERSION} \
			${VERSION} ./bin \
	'

################################################################################
# Targets to facilitate hacking on Brigade.                                    #
################################################################################

.PHONY: hack-new-kind-cluster
hack-new-kind-cluster:
	hack/kind/new-cluster.sh

.PHONY: hack-build-images
hack-build-images: hack-build-apiserver hack-build-scheduler hack-build-observer hack-build-logger-linux hack-build-git-initializer hack-build-worker

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

IMAGE_PULL_POLICY ?= Always

.PHONY: hack-deploy
hack-deploy:
	kubectl get namespace brigade || kubectl create namespace brigade
	helm dep up charts/brigade && \
	helm upgrade brigade charts/brigade \
		--install \
		--namespace brigade \
		--wait \
		--timeout 600s \
		--set apiserver.image.repository=$(DOCKER_IMAGE_PREFIX)apiserver \
		--set apiserver.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set apiserver.image.pullPolicy=$(IMAGE_PULL_POLICY) \
		--set apiserver.service.type=NodePort \
		--set apiserver.service.nodePort=31600 \
		--set scheduler.image.repository=$(DOCKER_IMAGE_PREFIX)scheduler \
		--set scheduler.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set scheduler.image.pullPolicy=$(IMAGE_PULL_POLICY) \
		--set observer.image.repository=$(DOCKER_IMAGE_PREFIX)observer \
		--set observer.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set observer.image.pullPolicy=$(IMAGE_PULL_POLICY) \
		--set worker.image.repository=$(DOCKER_IMAGE_PREFIX)worker \
		--set worker.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set worker.image.pullPolicy=$(IMAGE_PULL_POLICY) \
		--set gitInitializer.image.repository=$(DOCKER_IMAGE_PREFIX)git-initializer \
		--set gitInitializer.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set gitInitializer.image.pullPolicy=$(IMAGE_PULL_POLICY) \
		--set logger.linux.image.repository=$(DOCKER_IMAGE_PREFIX)logger-linux \
		--set logger.linux.image.tag=$(IMMUTABLE_DOCKER_TAG) \
		--set logger.linux.image.pullPolicy=$(IMAGE_PULL_POLICY)

.PHONY: hack
hack: hack-push-images hack-build-cli hack-deploy

.PHONY: hack-expose-apiserver
hack-expose-apiserver:
	@kubectl --namespace brigade port-forward service/brigade-apiserver 7000:443 &>/dev/null & \
		echo $$! > /tmp/brigade-apiserver.PID

.PHONY: hack-unexpose-apiserver
hack-unexpose-apiserver:
	@kill -TERM $$(cat /tmp/brigade-apiserver.PID)

# Convenience targets for loading images into a KinD cluster
.PHONY: hack-load-images
hack-load-images: load-apiserver load-scheduler load-observer load-logger-linux load-git-initializer load-worker

load-%:
	@echo "Loading $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)"
	@kind load docker-image $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG) \
			|| echo >&2 "kind not installed or error loading image: $(DOCKER_IMAGE_PREFIX)$*:$(IMMUTABLE_DOCKER_TAG)"

docs-stop-preview:
	@docker rm -f brigade-docs &> /dev/null || true

docs-preview: docs-stop-preview
	@docker run -d -v $$PWD:/src -p 1313:1313 --name brigade-docs -w /src/docs \
	klakegg/hugo:0.54.0-ext-alpine server -D -F --noHTTPCache --watch --bind=0.0.0.0
	# Wait for the documentation web server to finish rendering
	@until docker logs brigade-docs | grep -m 1  "Web Server is available"; do : ; done
	@open "http://localhost:1313"