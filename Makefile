# The Docker registry where images are pushed.
# Note that if you use an org (like on Quay and DockerHub), you should
# include that: quay.io/foo
DOCKER_REGISTRY    ?= deis
DOCKER_BUILD_FLAGS :=
LDFLAGS            :=

# Helm chart/release defaults
BRIGADE_RELEASE            ?= brigade-server
BRIGADE_NAMESPACE          ?= default
BRIGADE_GITHUB_GW_SERVICE  := $(BRIGADE_RELEASE)-brigade-github-gw
BRIGADE_GITHUB_GW_PORT     := 7744

BIN_NAMES        = brigade-api brigade-controller brigade-github-gateway brigade-cr-gateway brigade-generic-gateway brigade-vacuum brig
IMAGES           = brigade-api brigade-controller brigade-github-gateway brigade-cr-gateway brigade-generic-gateway brigade-vacuum brig brigade-worker git-sidecar

.PHONY: list
list:
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs

#.PHONY: echo-images
echo-images:
	@echo $(IMAGES)

GIT_TAG   = $(shell git describe --tags --always)
VERSION   ?= ${GIT_TAG}
IMAGE_TAG ?= ${VERSION}
LDFLAGS   += -X github.com/Azure/brigade/pkg/version.Version=$(VERSION)

# Set to "amd64 arm64 arm32" to build all archs
ARCHS ?= amd64
CX_OSES = linux windows darwin

amd64_GOARCH = amd64
arm64_GOARCH = arm64
arm32_GOARCH = arm

amd64_QEMU_IMPORT =
arm64_QEMU_IMPORT = ADD https://github.com/multiarch/qemu-user-static/releases/download/v3.1.0-2/x86_64_qemu-aarch64-static.tar.gz /usr/bin
arm32_QEMU_IMPORT = ADD https://github.com/multiarch/qemu-user-static/releases/download/v3.1.0-2/x86_64_qemu-arm-static.tar.gz /usr/bin

amd64_DOCKER_ARCH = amd64
arm64_DOCKER_ARCH = arm64v8
arm32_DOCKER_ARCH = arm32v6

amd64_DOCKER_ANNOTATIONS = --os linux --arch amd64
arm64_DOCKER_ANNOTATIONS = --os linux --arch arm64 --variant v8
arm32_DOCKER_ANNOTATIONS = --os linux --arch arm

# AMD64_BINS = $(shell echo "$(BIN_NAMES)" | sed 's/[a-z-]*/&\/rootfs\/amd64\/&/g')
# ARM64_BINS = $(shell echo "$(BIN_NAMES)" | sed 's/[a-z-]*/&\/rootfs\/arm64\/&/g')
# ARM32_BINS = $(shell echo "$(BIN_NAMES)" | sed 's/[a-z-]*/&\/rootfs\/arm32\/&/g')

.PHONY: build build-images push-images dockerfiles

define GO_CMD_TARGETS
# Build command
$1/rootfs/$2/$1: vendor $(shell find $1 -type f -name '*.go')
	GOOS=linux GOARCH=$$($2_GOARCH) CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$1/rootfs/$2/$1 ./$1/cmd/$1

build-$2 $1-$2-image: $1/rootfs/$2/$1
endef

define DOCKER_IMAGE_TARGETS
# Build image for a single arch
.PHONY: $1-images push-$1-images
$1-$2-image: $1/Dockerfile.$2
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/$1:$2-$(IMAGE_TAG) -f $1/Dockerfile.$2 $1

# Push image for a single arch
push-$1-$2-image: $1-$2-image
	docker push $(DOCKER_REGISTRY)/$1:$2-$(IMAGE_TAG)

# Create arch dockerfile
$1/Dockerfile.$2: $1/Dockerfile
	@cat $1/Dockerfile | ARCH="$2" QEMU_IMPORT="$$($2_QEMU_IMPORT)" DOCKER_ARCH="$$($2_DOCKER_ARCH)" envsubst '$$$${ARCH} $$$${QEMU_IMPORT} $$$${DOCKER_ARCH}' > $1/Dockerfile.$2

dockerfiles: $1/Dockerfile.$2
build-$2-images $1-images: $1-$2-image
push-$2-images push-$1-images: push-$1-$2-image
endef

define DOCKER_MANIFEST_TARGETS
.PHONY: create-$1-manifest $1

$1: $1-images

create-$1-manifest: push-$1-images
	# Remove old manifest if present
	@if [ -d $(HOME)/.docker/manifests/docker.io_$(DOCKER_REGISTRY)_$1-$(IMAGE_VERSION) ]; then \
		rm -rf $(HOME)/.docker/manifests/docker.io_$(DOCKER_REGISTRY)_$1-$(IMAGE_VERSION); \
	fi

	docker manifest create $(DOCKER_REGISTRY)/$1:$(IMAGE_TAG) $(foreach arch,$(ARCHS),$(DOCKER_REGISTRY)/$1:$(arch)-$(IMAGE_TAG))
	$(foreach arch,$(ARCHS),docker manifest annotate $$($(arch)_DOCKER_ANNOTATIONS) $(DOCKER_REGISTRY)/$1:$(IMAGE_TAG) $(DOCKER_REGISTRY)/$1:$(arch)-$(IMAGE_TAG);)
	docker manifest push $(DOCKER_REGISTRY)/$1:$(IMAGE_TAG)
endef

define BRIG_BUILD_TARGETS
ifeq ($1,windows)
ifeq ($2,amd64)
build-brig-static: bin/brig-$1-$2.exe
bin/brig-$1-$2.exe: vendor $(shell find brig -type f -name '*.go')
	GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$$@ ./brig/cmd/brig
endif
endif

ifeq ($1,darwin)
ifeq ($2,amd64)
build-brig-static: bin/brig-$1-$2
bin/brig-$1-$2: vendor $(shell find brig -type f -name '*.go')
	GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$$@ ./brig/cmd/brig
endif
endif

ifeq ($1,linux)
build-brig-static: bin/brig-$1-$2
bin/brig-$1-$2: vendor $(shell find brig -type f -name '*.go')
	GOOS=$1 GOARCH=$$($2_GOARCH) CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$$@ ./brig/cmd/brig
endif
endef

define GO_ARCH_TARGETS
.PHONY: build-$1 build-$1-images push-$1-images build-brig-static
build: build-$1 build-brig-static
build-images: build-$1-images
push-images: push-$1-images

$(foreach os,$(CX_OSES),$(eval $(call BRIG_BUILD_TARGETS,$(os),$1)))
$(foreach cmd,$(BIN_NAMES),$(eval $(call GO_CMD_TARGETS,$(cmd),$1)))
$(foreach img,$(IMAGES),$(eval $(call DOCKER_IMAGE_TARGETS,$(img),$1)))
endef

$(foreach arch,$(ARCHS),$(eval $(call GO_ARCH_TARGETS,$(arch))))
$(foreach img,$(IMAGES),$(eval $(call DOCKER_MANIFEST_TARGETS,$(img))))

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
# and Brigade chart installed with `--set brigade-github-app.enabled=true`
#
# To set this up in your local environment:
# - Make sure kubectl is pointed to the right cluster
# - Run "helm repo add brigade https://azure.github.io/brigade-charts"
# - Run "helm inspect values brigade/brigade-project > myvals.yaml"
# - Set the values in myvalues.yaml to something like this:
#   project: "deis/empty-testbed"
#   repository: "github.com/deis/empty-testbed"
#   secret: "MySecret"
# - Run "helm install brigade/brigade-project -f myvals.yaml"
# - Run "make test-functional"
#
# This will clone the github.com/deis/empty-testbed repo and run the brigade.js
# file found there.
# Test Repo is https://github.com/deis/empty-testbed
TEST_REPO_COMMIT =  589e15029e1e44dee48de4800daf1f78e64287c0
KUBECONFIG       ?= ${HOME}/.kube/config
.PHONY: test-functional
test-functional: vendor
test-functional:
	@kubectl port-forward service/$(BRIGADE_GITHUB_GW_SERVICE) $(BRIGADE_GITHUB_GW_EXTERNAL_PORT):80 &>/dev/null & \
		echo $$! > /tmp/$(BRIGADE_GITHUB_GW_SERVICE).PID
	go test --tags integration ./tests -kubeconfig $(KUBECONFIG) $(TEST_REPO_COMMIT)
	@kill -TERM $$(cat /tmp/$(BRIGADE_GITHUB_GW_SERVICE).PID)

# JS test is local only
.PHONY: test-js
test-js:
	cd brigade-worker && yarn build && KUBECONFIG="./test/fake_kubeconfig.yaml" yarn test

.PHONY: test-style
test-style:
	golangci-lint run --config ./golangci.yml

.PHONY: format
format: format-go format-js

.PHONY: format-go
format-go:
	go list -f '{{.Dir}}' ./... | xargs goimports -w -local github.com/Azure/brigade

.PHONY: format-js
format-js:
	cd brigade-worker && yarn format

HAS_GOLANGCI     := $(shell command -v golangci-lint;)
HAS_DEP          := $(shell command -v dep;)
HAS_GIT          := $(shell command -v git;)
HAS_YARN         := $(shell command -v yarn;)

.PHONY: bootstrap-js
bootstrap-js:
ifndef HAS_YARN
	$(error You must install yarn)
endif
	cd brigade-worker && yarn install

vendor:
ifndef HAS_GIT
	$(error You must install git)
endif
ifndef HAS_DEP
	go get -u github.com/golang/dep/cmd/dep
endif
ifndef HAS_GOLANGCI
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint && \
	cd $(GOPATH)/src/github.com/golangci/golangci-lint/cmd/golangci-lint && \
	go install -ldflags "-X 'main.version=$(git describe --tags)' -X 'main.commit=$(git rev-parse --short HEAD)' -X 'main.date=$(date)'"
endif
	dep ensure

.PHONY: bootstrap
bootstrap: vendor