# The Docker registry where images are pushed.
# Note that if you use an org (like on Quay and DockerHub), you should
# include that: quay.io/foo
DOCKER_REGISTRY ?= acidic.azurecr.io
DOCKER_BUILD_FLAGS :=

# Test Repo is https://github.com/deis/empty-testbed
TEST_REPO_EVENT="X-GitHub-Event: push"
TEST_REPO_COMMIT=6f24c2fd0f13dc95f9056a6448f16607b8d1fa6e

# The location of the functional tests.
TEST_DIR=./_functional_tests

.PHONY: build
build: generate
build:
	go build -o bin/acid .
	go build -o bin/vcs-sidecar vcs-sidecar/cmd/vcs-sidecar/main.go

# Cross-compile for Docker+Linux
.PHONY: build-docker-bin
build-docker-bin: generate
build-docker-bin:
	GOOS=linux GOARCH=amd64 go build -o rootfs/usr/bin/acid .
	GOOS=linux GOARCH=amd64 go build -o vcs-sidecar/rootfs/vcs-sidecar vcs-sidecar/cmd/vcs-sidecar/main.go

.PHONY: run
run: build
run:
	bin/acid

# To use docker-build, you need to have Docker installed and configured. You should also set
# DOCKER_REGISTRY to your own personal registry if you are not pushing to the official upstream.
.PHONY: docker-build
docker-build: build-docker-bin
docker-build:
	docker build -t $(DOCKER_REGISTRY)/acid:latest .
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/vcs-sidecar:latest vcs-sidecar

# You must be logged into DOCKER_REGISTRY before you can push.
.PHONY: docker-push
docker-push:
	docker push $(DOCKER_REGISTRY)/acid
	docker push $(DOCKER_REGISTRY)/vcs-sidecar

# Unit tests. Local only.
.PHONY: test-unit
test-unit: generate
test-unit:
	go test -v . ./pkg/...

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
test-functional:
	-kubectl delete pod test-repo-$(TEST_REPO_COMMIT)
	-kubectl delete cm  test-repo-$(TEST_REPO_COMMIT) && sleep 10
	go run $(TEST_DIR)/generate.go $(TEST_REPO_COMMIT) && curl -X POST \
		-H $(TEST_REPO_EVENT) \
		-H "X-Hub-Signature: $(shell cat $(TEST_DIR)/test-repo-generated.hash)" \
		localhost:7744/events/github \
		-vvv -T $(TEST_DIR)/test-repo-generated.json

# JS test is local only
.PHONY: test-js
test-js:
	eslint js/*.js acid.js

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
