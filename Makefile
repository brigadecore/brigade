# The Docker registry where images are pushed.
# Note that if you use an org (like on Quay and DockerHub), you should
# include that: quay.io/foo
DOCKER_REGISTRY    ?= deis
DOCKER_BUILD_FLAGS :=
LDFLAGS            :=

BINS        = brigade-api brigade-controller brigade-github-gateway brigade-cr-gateway brigade-vacuum brig
IMAGES      = brigade-api brigade-controller brigade-github-gateway brigade-cr-gateway brigade-vacuum brig brigade-worker git-sidecar

GIT_TAG   = $(shell git describe --tags --always 2>/dev/null)
VERSION   ?= ${GIT_TAG}
IMAGE_TAG ?= ${VERSION}
LDFLAGS   += -X github.com/Azure/brigade/pkg/version.Version=$(VERSION)

CX_OSES = linux windows darwin
CX_ARCHS = amd64

# Build native binaries
.PHONY: build
build: $(BINS)

.PHONY: $(BINS)
$(BINS): vendor
	go build -ldflags '$(LDFLAGS)' -o bin/$@ ./$@/cmd/$@

# Cross-compile for Docker+Linux
build-docker-bins: $(addsuffix -docker-bin,$(BINS))

%-docker-bin: vendor
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$*/rootfs/$* ./$*/cmd/$*

# To use docker-build, you need to have Docker installed and configured. You should also set
# DOCKER_REGISTRY to your own personal registry if you are not pushing to the official upstream.
.PHONY: docker-build
docker-build: build-docker-bins
docker-build: $(addsuffix -image,$(IMAGES))

%-image:
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/$*:$(IMAGE_TAG) $*

# You must be logged into DOCKER_REGISTRY before you can push.
.PHONY: docker-push
docker-push: $(addsuffix -push,$(IMAGES))

%-push:
	docker push $(DOCKER_REGISTRY)/$*:$(IMAGE_TAG)

# Cross-compile binaries for our CX targets.
# Mainly, this is for brig-cross-compile
%-cross-compile: vendor
	@for os in $(CX_OSES); do \
		echo "building $$os"; \
		for arch in $(CX_ARCHS); do \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./bin/$*-$$os-$$arch ./$*/cmd/$*; \
		done;\
	done

.PHONY: build-release
build-release: brig-cross-compile

.PRECIOUS: build-chart
.PHONY: build-chart
build-chart:
	helm package -d docs/ ./charts/brigade
	helm package -d docs/ ./charts/brigade-project
	helm repo index docs/

# All non-functional tests
.PHONY: test
test: test-style
test: test-unit
test: test-js

# Unit tests. Local only.
.PHONY: test-unit
test-unit: vendor
	go test -v ./...

# Functional tests assume access to github.com
# To set this up in your local environment:
# - Make sure kubectl is pointed to the right cluster
# - Create "myvals.yaml" and set to something like this:
#   project: "deis/empty-testbed"
#   repository: "github.com/deis/empty-testbed"
#   secret: "MySecret"
# - Run "helm install ./charts/brigade-project -f myvals.yaml
# - Run "make run" in one terminal
# - Run "make test-functional" in another terminal
#
# This will clone the github.com/deis/empty-testbed repo and run the brigade.js
# file found there.
# Test Repo is https://github.com/deis/empty-testbed
TEST_REPO_COMMIT =  589e15029e1e44dee48de4800daf1f78e64287c0
KUBECONFIG       ?= ${HOME}/.kube/config
.PHONY: test-functional
test-functional: vendor
test-functional:
	go test --tags integration ./tests -kubeconfig $(KUBECONFIG) $(TEST_REPO_COMMIT)

# JS test is local only
.PHONY: test-js
test-js:
	cd brigade-worker && yarn test

.PHONY: test-style
test-style:
	gometalinter --config ./gometalinter.json ./...

.PHONY: format
format: format-go format-js

.PHONY: format-go
format-go:
	go list -f '{{.Dir}}' ./... | xargs goimports -w -local github.com/Azure/brigade

.PHONY: format-js
format-js:
	cd brigade-worker && yarn format

HAS_GOMETALINTER := $(shell command -v gometalinter;)
HAS_DEP          := $(shell command -v dep;)
HAS_GIT          := $(shell command -v git;)

vendor:
ifndef HAS_GIT
	$(error You must install git)
endif
ifndef HAS_DEP
	go get -u github.com/golang/dep/cmd/dep
endif
ifndef HAS_GOMETALINTER
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
endif
	dep ensure

.PHONY: bootstrap
bootstrap: vendor
