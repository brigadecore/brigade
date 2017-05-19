REG=technosophos

# For test runs
ZOLVER_EVENT="X-GitHub-Event: push"
ZOLVER_TEST_COMMIT=cbb38c431c40d9168e652f6a43a73a245fb3ef99
TEST_DIR=./_functional_tests

.PHONY: build
build:
	go build -o bin/acid .

.PHONY: build-docker-bin
build-docker-bin:
	GOOS=linux GOARCH=amd64 go build -o chart/rootfs/acid .
	cp runner.js chart/rootfs/

.PHONY: run
run: build
run:
	bin/acid

.PHONY: docker-build
docker-build: build-docker-bin
docker-build:
	docker build -t $(REG)/acid:latest chart/rootfs
	docker build --no-cache -t $(REG)/acid-ubuntu:latest acidic/acid-ubuntu
	docker build --no-cache -t $(REG)/acid-go:latest acidic/acid-go
	docker build --no-cache -t $(REG)/acid-node:latest acidic/acid-node

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
		localhost:7744/webhook/push \
		-vvv -T $(TEST_DIR)/zolver-generated.json

.PHONY: test-js
test-js:
	eslint runner.js
