# The Docker registry where images are pushed.
# Note that if you use an org (like on Quay and DockerHub), you should
# include that: quay.io/foo
DOCKER_REGISTRY ?= acidic.azurecr.io
DOCKER_BUILD_FLAGS :=

# Test Repo is https://github.com/deis/empty-testbed
TEST_REPO_COMMIT=033f10f0d16cf703710b59269a34b8f4ce8537cb

# The location of the functional tests.
TEST_DIR=./tests

KUBECONFIG ?= ${HOME}/.kube/config

.PHONY: build
build:
	go build -o bin/acid .
	go build -o bin/acid-controller ./acid-controller/cmd/acid-controller
	go build -o bin/acid-api ./acid-api/cmd/acid-api
	go build -o bin/vcs-sidecar ./vcs-sidecar/cmd/vcs-sidecar
	go build -o bin/lsd ./acid-client/cmd/lsd

# Cross-compile for Docker+Linux
.PHONY: build-docker-bin
build-docker-bin:
	GOOS=linux GOARCH=amd64 go build -o rootfs/usr/bin/acid .
	GOOS=linux GOARCH=amd64 go build -o ./acid-controller/rootfs/acid-controller ./acid-controller/cmd/acid-controller
	GOOS=linux GOARCH=amd64 go build -o ./acid-api/rootfs/acid-api ./acid-api/cmd/acid-api
	GOOS=linux GOARCH=amd64 go build -o ./vcs-sidecar/rootfs/vcs-sidecar ./vcs-sidecar/cmd/vcs-sidecar

.PHONY: run
run: build
run:
	bin/acid -kubeconfig $(KUBECONFIG)

# To use docker-build, you need to have Docker installed and configured. You should also set
# DOCKER_REGISTRY to your own personal registry if you are not pushing to the official upstream.
.PHONY: docker-build
docker-build: build-docker-bin
docker-build:
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/acid:latest .
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/acid-controller:latest acid-controller
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/acid-api:latest acid-api
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/vcs-sidecar:latest vcs-sidecar
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/acid-worker:latest acid-worker

# You must be logged into DOCKER_REGISTRY before you can push.
.PHONY: docker-push
docker-push:
	docker push $(DOCKER_REGISTRY)/acid
	docker push $(DOCKER_REGISTRY)/acid-controller
	docker push $(DOCKER_REGISTRY)/acid-api
	docker push $(DOCKER_REGISTRY)/vcs-sidecar
	docker push $(DOCKER_REGISTRY)/acid-worker

# All non-functional tests
.PHONY: test
test: test-style
test: test-unit
test: test-js

# Unit tests. Local only.
.PHONY: test-unit
test-unit:
	go test -v . ./pkg/... ./acid-controller/...

# Functional tests assume access to github.com
# To set this up in your local environment:
# - Make sure kubectl is pointed to the right cluster
# - Create "myvals.yaml" and set to something like this:
#   project: "deis/empty-testbed"
#   repository: "github.com/deis/empty-testbed"
#   secret: "MySecret"
# - Run "helm install ./chart/acid-project -f myvals.yaml
# - Run "make run" in one terminal
# - Run "make test-functional" in another terminal
#
# This will clone the github.com/deis/empty-testbed repo and run the acid.js
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
	docker run $(DOCKER_REGISTRY)/acid-worker:latest npm run test

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
HAS_GLIDE := $(shell command -v glide;)
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
ifndef HAS_GLIDE
	go get -u github.com/Masterminds/glide
endif
ifndef HAS_BINDATA
	go get github.com/jteeuwen/go-bindata/...
endif
	glide install --strip-vendor
