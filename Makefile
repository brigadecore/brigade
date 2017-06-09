REG=acidic.azurecr.io

# For test runs
ZOLVER_EVENT="X-GitHub-Event: push"
ZOLVER_TEST_COMMIT=cbb38c431c40d9168e652f6a43a73a245fb3ef99
TEST_DIR=./_functional_tests
DOCKER_BUILD_FLAGS :=

.PHONY: build
build: generate
build:
	go build -o bin/acid .

.PHONY: build-docker-bin
build-docker-bin: generate
build-docker-bin:
	GOOS=linux GOARCH=amd64 go build -o chart/rootfs/acid .

.PHONY: run
run: build
run:
	bin/acid

.PHONY: docker-build
docker-build: build-docker-bin
docker-build:
	docker build -t $(REG)/acid:latest chart/rootfs
	docker build $(DOCKER_BUILD_FLAGS) -t $(REG)/acid-ubuntu:latest acidic/acid-ubuntu
	docker build $(DOCKER_BUILD_FLAGS) -t $(REG)/acid-go:latest acidic/acid-go
	docker build $(DOCKER_BUILD_FLAGS) -t $(REG)/acid-node:latest acidic/acid-node

.PHONY: docker-push
docker-push:
	docker push $(REG)/acid
	docker push $(REG)/acid-go
	docker push $(REG)/acid-node
	docker push $(REG)/acid-ubuntu

.PHONY: docker-test
docker-test: docker-build
docker-test:
	docker run \
		-e CLONE_URL=https://github.com/technosophos/zolver.git \
		-e HEAD_COMMIT_ID=$(ZOLVER_TEST_COMMIT) \
		$(REG)/acid-go:latest
	docker run \
		-e CLONE_URL=https://github.com/technosophos/zolver.git \
		-e HEAD_COMMIT_ID=$(ZOLVER_TEST_COMMIT) \
		$(REG)/acid-node:latest
	docker run \
		-e CLONE_URL=https://github.com/technosophos/zolver.git \
		-e HEAD_COMMIT_ID=$(ZOLVER_TEST_COMMIT) \
		$(REG)/acid-ubuntu:latest

.PHONY: test-unit
test-unit: generate
test-unit:
	go test -v . ./pkg/...

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

.PHONY: test-js
test-js:
	eslint js/runner.js

generate:
	go-bindata --pkg lib --nometadata -o pkg/js/lib/generated.go js

HAS_GOMETALINTER := $(shell command -v gometalinter;)
HAS_GLIDE := $(shell command -v glide;)
HAS_GOX := $(shell command -v gox;)
HAS_GIT := $(shell command -v git;)
HAS_BINDATA := $(shell command -v go-bindata;)

.PHONY: bootstrap
bootstrap:
ifndef HAS_GOMETALINTER
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
endif
ifndef HAS_GLIDE
	go get -u github.com/Masterminds/glide
endif
ifndef HAS_GIT
	$(error You must install git)
endif
ifndef HAS_BINDATA
	go get github.com/jteeuwen/go-bindata/...
endif
	glide install --strip-vendor
