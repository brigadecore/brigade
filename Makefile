REG=technosophos

# For test runs
ZOLVER_EVENT="X-GitHub-Event: push"
ZOLVER_HUB_SIGNATURE="X-Hub-Signature: sha1=206ed654666106fe879a17b171f60dde3661ebb9"
ZOLVER_TEST_COMMIT=d36f0682e3d7d1b619bef04945be8b0062d69841

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
	docker build -t $(REG)/acid-ubuntu:latest acidic/acid-ubuntu
	docker build -t $(REG)/acid-go:latest acidic/acid-go

.PHONY: docker-push
docker-push:
	docker push $(REG)/acid
	docker push $(REG)/acid-go
	docker push $(REG)/acid-ubuntu

.PHONY: docker-test
docker-test: docker-build
docker-test:
	docker run \
		-e CLONE_URL=https://github.com/technosophos/zolver.git \
		-e HEAD_COMMIT_ID=$(ZOLVER_TEST_COMMIT) \
		$(REG)/acid-go:latest

.PHONY: test-unit
test-unit:
	go test -v .

.PHONY: test-functional
test-functional:
	-kubectl delete pod test-zolver-$(ZOLVER_TEST_COMMIT)
	-kubectl delete cm  test-zolver-$(ZOLVER_TEST_COMMIT) && sleep 10
	curl -X POST \
		-H $(ZOLVER_EVENT) \
		-H $(ZOLVER_HUB_SIGNATURE) \
		localhost:7744/webhook/push \
		-vvv -T ./_functional_tests/zolver.json

