# The Docker registry where images are pushed.
# Note that if you use an org (like on Quay and DockerHub), you should
# include that: quay.io/foo
DOCKER_REGISTRY    ?= deis
DOCKER_BUILD_FLAGS :=
VERSION            := $(shell git describe --tags --abbrev=0 2>/dev/null)
LDFLAGS            :=

# Test Repo is https://github.com/deis/empty-testbed
TEST_REPO_COMMIT=589e15029e1e44dee48de4800daf1f78e64287c0

# The location of the functional tests.
TEST_DIR=./tests

KUBECONFIG ?= ${HOME}/.kube/config

LDFLAGS += -X github.com/Azure/brigade/pkg/version.Version=${VERSION}

.PHONY: build
build: build-client
build:
	go build -ldflags '$(LDFLAGS)' -o bin/brigade-gateway ./brigade-gateway/cmd/brigade-gateway
	go build -ldflags '$(LDFLAGS)' -o bin/brigade-controller ./brigade-controller/cmd/brigade-controller
	go build -ldflags '$(LDFLAGS)' -o bin/brigade-api ./brigade-api/cmd/brigade-api

.PHONY: build-client
build-client:
	go build -ldflags '$(LDFLAGS)' -o bin/brig ./brigade-client/cmd/brig

# Cross-compile for Docker+Linux
.PHONY: build-docker-bin
build-docker-bin:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./brigade-gateway/rootfs/usr/bin/brigade-gateway ./brigade-gateway/cmd/brigade-gateway
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./brigade-controller/rootfs/brigade-controller ./brigade-controller/cmd/brigade-controller
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./brigade-api/rootfs/brigade-api ./brigade-api/cmd/brigade-api

.PHONY: run
run: build
run:
	bin/brigade -kubeconfig $(KUBECONFIG)

# To use docker-build, you need to have Docker installed and configured. You should also set
# DOCKER_REGISTRY to your own personal registry if you are not pushing to the official upstream.
.PHONY: docker-build
docker-build: build-docker-bin
docker-build:
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/brigade-gateway:$(VERSION) brigade-gateway
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/brigade-controller:$(VERSION) brigade-controller
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/brigade-api:$(VERSION) brigade-api
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/git-sidecar:$(VERSION) git-sidecar
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/brigade-worker:$(VERSION) brigade-worker

# You must be logged into DOCKER_REGISTRY before you can push.
.PHONY: docker-push
docker-push:
	docker push $(DOCKER_REGISTRY)/brigade-gateway
	docker push $(DOCKER_REGISTRY)/brigade-controller
	docker push $(DOCKER_REGISTRY)/brigade-api
	docker push $(DOCKER_REGISTRY)/git-sidecar
	docker push $(DOCKER_REGISTRY)/brigade-worker

.PRECIOUS: build-chart
.PHONY: build-chart
build-chart:
	helm package -d docs/ ./chart/brigade
	helm package -d docs/ ./chart/brigade-project
	helm repo index docs/

# All non-functional tests
.PHONY: test
test: test-style
test: test-unit
test: test-js

# Unit tests. Local only.
.PHONY: test-unit
test-unit:
	go test -v ./pkg/... ./brigade-controller/... ./brigade-gateway/... ./brigade-api/...

# Functional tests assume access to github.com
# To set this up in your local environment:
# - Make sure kubectl is pointed to the right cluster
# - Create "myvals.yaml" and set to something like this:
#   project: "deis/empty-testbed"
#   repository: "github.com/deis/empty-testbed"
#   secret: "MySecret"
# - Run "helm install ./chart/brigade-project -f myvals.yaml
# - Run "make run" in one terminal
# - Run "make test-functional" in another terminal
#
# This will clone the github.com/deis/empty-testbed repo and run the brigade.js
# file found there.
.PHONY: test-functional
test-functional: test-functional-prepare
test-functional:
	go test $(TEST_DIR)

.PHONY: test-functional-prepare
test-functional-prepare:
	go run $(TEST_DIR)/cmd/generate.go -kubeconfig $(KUBECONFIG) $(TEST_REPO_COMMIT)

# JS test is local only
.PHONY: test-js
test-js:
	cd brigade-worker && yarn test

.PHONY: test-style
test-style:
	gometalinter.v1 \
		--disable-all \
		--enable deadcode \
		--severity deadcode:error \
		--enable gofmt \
		--enable ineffassign \
		--enable misspell \
		--enable vet \
		--tests \
		--vendor \
		--deadline 60s \
		./...
	@echo "Recommended style checks ===>"
	gometalinter.v1 \
		--disable-all \
		--enable golint \
		--vendor \
		--skip proto \
		--deadline 60s \
		./... || :


HAS_NPM := $(shell command -v npm;)
HAS_ESLINT := $(shell command -v eslint;)
HAS_GOMETALINTER := $(shell command -v gometalinter;)
HAS_DEP := $(shell command -v dep;)
HAS_GOX := $(shell command -v gox;)
HAS_GIT := $(shell command -v git;)
HAS_BINDATA := $(shell command -v go-bindata;)

.PHONY: bootstrap
bootstrap:
ifndef HAS_NPM
	$(error You must install npm)
endif
ifndef HAS_GIT
	$(error You must install git)
endif
ifndef HAS_ESLINT
	npm install -g eslint
endif
ifndef HAS_GOMETALINTER
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
endif
ifndef HAS_DEP
	go get -u github.com/golang/dep/cmd/dep
endif
ifndef HAS_BINDATA
	go get github.com/jteeuwen/go-bindata/...
endif
	dep ensure
