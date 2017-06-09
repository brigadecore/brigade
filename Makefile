# The Docker registry where images are pushed.
# Note that if you use an org (like on Quay and DockerHub), you should
# include that: quay.io/foo
DOCKER_REGISTRY ?= acidic.azurecr.io
DOCKER_BUILD_FLAGS :=

# Zolver is a repo that we use for some testing. Really, we need to replace this
# whole thing with something better.
ZOLVER_EVENT="X-GitHub-Event: push"
ZOLVER_TEST_COMMIT=cbb38c431c40d9168e652f6a43a73a245fb3ef99

# The location of the functional tests.
TEST_DIR=./_functional_tests

.PHONY: build
build: generate
build:
	go build -o bin/acid .

# Cross-compile for Docker+Linux
.PHONY: build-docker-bin
build-docker-bin: generate
build-docker-bin:
	GOOS=linux GOARCH=amd64 go build -o chart/rootfs/acid .

.PHONY: run
run: build
run:
	bin/acid

# To use docker-build, you need to have Docker installed and configured. You should also set
# DOCKER_REGISTRY to your own personal registry if you are not pushing to the official upstream.
.PHONY: docker-build
docker-build: build-docker-bin
docker-build:
	docker build -t $(DOCKER_REGISTRY)/acid:latest chart/rootfs
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/acid-ubuntu:latest acidic/acid-ubuntu
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/acid-go:latest acidic/acid-go
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/acid-node:latest acidic/acid-node

# You must be logged into DOCKER_REGISTRY before you can push.
.PHONY: docker-push
docker-push:
	docker push $(DOCKER_REGISTRY)/acid
	docker push $(DOCKER_REGISTRY)/acid-go
	docker push $(DOCKER_REGISTRY)/acid-node
	docker push $(DOCKER_REGISTRY)/acid-ubuntu

# docker-test attempts to fetch known commits from a live repo, and then perform some basic tests.
# This, too, should be replaced by something more robust
.PHONY: docker-test
docker-test: docker-build
docker-test:
	docker run \
		-e CLONE_URL=https://github.com/technosophos/zolver.git \
		-e HEAD_COMMIT_ID=$(ZOLVER_TEST_COMMIT) \
		$(DOCKER_REGISTRY)/acid-go:latest
	docker run \
		-e CLONE_URL=https://github.com/technosophos/zolver.git \
		-e HEAD_COMMIT_ID=$(ZOLVER_TEST_COMMIT) \
		$(DOCKER_REGISTRY)/acid-node:latest
	docker run \
		-e CLONE_URL=https://github.com/technosophos/zolver.git \
		-e HEAD_COMMIT_ID=$(ZOLVER_TEST_COMMIT) \
		$(DOCKER_REGISTRY)/acid-ubuntu:latest

# Unit tests. Local only.
.PHONY: test-unit
test-unit: generate
test-unit:
	go test -v . ./pkg/...

# Functional tests assume access to github.com
.PHONY: test-functional
test-functional:
	-kubectl delete pod test-zolver-$(ZOLVER_TEST_COMMIT)
	-kubectl delete cm  test-zolver-$(ZOLVER_TEST_COMMIT) && sleep 10
	go run $(TEST_DIR)/generate.go $(ZOLVER_TEST_COMMIT)
	curl -X POST \
		-H $(ZOLVER_EVENT) \
		-H "X-Hub-Signature: $(shell cat $(TEST_DIR)/zolver-generated.hash)" \
		localhost:7744/events/github \
		-vvv -T $(TEST_DIR)/zolver-generated.json

# JS test is local only
.PHONY: test-js
test-js:
	eslint js/runner.js

# Compile the JS into the Go
# We don't call `go generate` anymore because it is a redundant abstraction.
generate:
	go-bindata --pkg lib --nometadata -o pkg/js/lib/generated.go js

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
