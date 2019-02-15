######################################################################
## Brigade makefile
## 
## Target dependency graph
##
## echo-images: list all image names
##
## bootstrap: setup environment
##
## build: build binaries for the go projects
## -> build-$(binary): build specific binary
##    -> build-$(binary)-$(os): build specific binary for a single os
##       -> build-$(binary)-$(os)-$(arch): build for a single arch
##
## build-image-bins: build binaries for docker images
## -> build-image-bins-$(binary): build image bins for specific image
##    -> build-image-bins-$(binary)-$(arch): single arch
##
## build-images: build all images
## -> build-$(image)-images: build all images for a single repo
##    -> build-$(image)-$(arch)-images: build image for a single repo and arch
##
## push-images: push all images
## -> build-images
## -> push-$(image)-images: push all images for a single repo
##    -> build-$(image)-images
##    -> push-$(image)-$(arch)-images: push images for a single repo and arch
##       -> build-$(image)-$(arch)-images
##
## create-manifests: creates and pushes all manifests
## -> push-images
## -> create-$(image)-manifests: create and push manifest for a single repo

######################################################################
## ARGUMENTS AND VARIABLES
######################################################################

# Enable debugging
DEBUG_MAKE ?=

# Skip building docker files (only do this if you know the right versions are built)
SKIP_DOCKER_BUILD ?=
SKIP_DOCKER_PUSH ?=

# The Docker registry where images are pushed.
# Note that if you use an org (like on Quay and DockerHub), you should
# include that: quay.io/foo
DOCKER_REGISTRY    ?= deis
DOCKER_BUILD_FLAGS :=
LDFLAGS            :=

# Set to "amd64 arm64 arm32" to build all archs
ARCHS ?= amd64
OSES ?= linux windows darwin

# Helm chart/release defaults
BRIGADE_RELEASE                 ?= brigade-server
BRIGADE_NAMESPACE               ?= default
BRIGADE_GITHUB_GW_SERVICE       := $(BRIGADE_RELEASE)-brigade-github-app
BRIGADE_GITHUB_GW_INTERNAL_PORT := 80
BRIGADE_GITHUB_GW_EXTERNAL_PORT := 7744

BINS        = brigade-api brigade-controller brigade-cr-gateway brigade-generic-gateway brigade-vacuum brig
IMAGES      = brigade-api brigade-controller brigade-cr-gateway brigade-generic-gateway brigade-vacuum brig brigade-worker git-sidecar

GIT_TAG   = $(shell git describe --tags --always)
VERSION   ?= ${GIT_TAG}
IMAGE_TAG ?= ${VERSION}
LDFLAGS   += -X github.com/Azure/brigade/pkg/version.Version=$(VERSION)


amd64_GOARCH := amd64
arm64_GOARCH := arm64
arm32_GOARCH := arm

amd64_QEMU_IMPORT :=
arm64_QEMU_IMPORT := ADD https://github.com/multiarch/qemu-user-static/releases/download/v3.1.0-2/x86_64_qemu-aarch64-static.tar.gz /usr/bin
arm32_QEMU_IMPORT := ADD https://github.com/multiarch/qemu-user-static/releases/download/v3.1.0-2/x86_64_qemu-arm-static.tar.gz /usr/bin

amd64_DOCKER_ARCH := amd64
arm64_DOCKER_ARCH := arm64v8
arm32_DOCKER_ARCH := arm32v6

amd64_DOCKER_ANNOTATIONS := --os linux --arch amd64
arm64_DOCKER_ANNOTATIONS := --os linux --arch arm64 --variant v8
arm32_DOCKER_ANNOTATIONS := --os linux --arch arm

######################################################################
## MAIN TARGETS
######################################################################

.PHONY: echo-images
echo-images:
	$(call echo_invocation)
	@echo $(IMAGES)

.PHONY: bootstrap
bootstrap: _vendor _bootstrap-js
	$(call echo_invocation)

.PHONY: build
build: _vendor
	$(call echo_invocation)

.PHONY: build-image-bins
build-image-bins: _vendor
	$(call echo_invocation)

.PHONY: build-images
build-images: _vendor
	$(call echo_invocation)

.PHONY: push-images
ifndef SKIP_DOCKER_BUILD
push-images: build-images
else
push-images:
endif
	$(call echo_invocation)

.PHONY: create-manifests
ifndef SKIP_DOCKER_PUSH
create-manifests: push-images
else
create-manifests:
endif
	$(call echo_invocation)

######################################################################
## RELEASE TARGETS
######################################################################

.PHONY: build-release
build-release: brig-cross-compile
	$(call echo_invocation)

.PHONY: helm-install
helm-install: helm-upgrade
	$(call echo_invocation)

.PHONY: helm-upgrade
helm-upgrade:
	$(call echo_invocation)
	helm upgrade --install $(BRIGADE_RELEASE) brigade/brigade --namespace $(BRIGADE_NAMESPACE) \
		--set brigade-github-app.enabled=true \
		--set controller.tag=$(VERSION) \
		--set api.tag=$(VERSION) \
		--set worker.tag=$(VERSION) \
		--set cr.tag=$(VERSION) \
		--set vacuum.tag=$(VERSION)

######################################################################
## TEST TARGETS
######################################################################

# All non-functional tests
.PHONY: test
test: test-style test-unit test-js
	$(call echo_invocation)

# Unit tests. Local only.
.PHONY: test-unit
test-unit: vendor
	$(call echo_invocation)
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
TEST_REPO_COMMIT :=  589e15029e1e44dee48de4800daf1f78e64287c0
KUBECONFIG       ?= ${HOME}/.kube/config
.PHONY: test-functional
test-functional:
	$(call echo_invocation)
	@kubectl port-forward service/$(BRIGADE_GITHUB_GW_SERVICE) $(BRIGADE_GITHUB_GW_EXTERNAL_PORT):80 &>/dev/null & \
		echo $$! > /tmp/$(BRIGADE_GITHUB_GW_SERVICE).PID
	go test --tags integration ./tests -kubeconfig $(KUBECONFIG) $(TEST_REPO_COMMIT)
	@kill -TERM $$(cat /tmp/$(BRIGADE_GITHUB_GW_SERVICE).PID)

# JS test is local only
.PHONY: test-js
test-js:
	$(call echo_invocation)
	cd brigade-worker && yarn build && KUBECONFIG="./test/fake_kubeconfig.yaml" yarn test

.PHONY: test-style
test-style:
	$(call echo_invocation)
	golangci-lint run --config ./golangci.yml

######################################################################
## TEST TARGETS
######################################################################

.PHONY: format
format: format-go format-js
	$(call echo_invocation)

.PHONY: format-go
format-go:
	$(call echo_invocation)
	go list -f '{{.Dir}}' ./... | xargs goimports -w -local github.com/Azure/brigade

.PHONY: format-js
format-js:
	$(call echo_invocation)
	cd brigade-worker && yarn format

######################################################################
## PRIVATE TARGETS
######################################################################

HAS_GOLANGCI     := $(shell command -v golangci-lint;)
HAS_DEP          := $(shell command -v dep;)
HAS_GIT          := $(shell command -v git;)
HAS_YARN         := $(shell command -v yarn;)
HAS_GOMPLATE     := $(shell command -v gomplate;)

.PHONY: _bootstrap-js
_bootstrap-js:
	$(call echo_invocation)
ifndef HAS_YARN
	$(error You must install yarn)
endif
	cd brigade-worker && yarn install

.PHONY: _vendor
_vendor:
	$(call echo_invocation)
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
ifndef HAS_GOMPLATE
	go get github.com/hairyhenderson/gomplate/cmd/gomplate
endif
	@dep ensure

######################################################################
## TARGET TEMPLATES
######################################################################

# Targets for a single binary
# $1: binary name (brigade-api)
#
# Targets:
# build-$(binary)
define BIN_TARGETS
.PHONY build: build-$1
build-$1:
	$$(call echo_invocation)

$(foreach os,$(OSES),$(eval $(call BIN_OS_TARGETS,$1,$(os))))
endef

# Targets for a single binary on a given os
# $1: binary name (brigade-api)
# $2: os (windows)
#
# Targets:
# build-$(binary)-$(os)
define BIN_OS_TARGETS
.PHONY build-$1: build-$1-$2
build-$1-$2:
	$$(call echo_invocation)

$(foreach arch,$(ARCHS),$(eval $(call BIN_OS_ARCH_TARGETS,$1,$2,$(arch))))
endef

# Targets for a single binary on a given os and arch
# $1: binary name (brigade-api)
# $2: os (windows)
# $3: arch (amd64)
#
# Targets:
# build-$(binary)-$(os)-$(arch) [optional]
define BIN_OS_ARCH_TARGETS
ifeq ($2,windows)
ifeq ($3,amd64)
.PHONY build-$1-$2: build-$1-$2-$3
build-$1-$2-$3: bin/$1-$2-$3.exe _vendor
	$$(call echo_invocation)

bin/$1-$2-$3.exe: $(shell find $1 -type f -name '*.go') _vendor
	$$(call echo_invocation)
	GOOS=$2 GOARCH=$$($3_GOARCH) CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$$@ ./$1/cmd/$1
endif
endif

ifeq ($2,darwin)
ifeq ($3,amd64)
.PHONY build-$1-$2: build-$1-$2-$3
build-$1-$2-$3: bin/$1-$2-$3 _vendor
	$$(call echo_invocation)

bin/$1-$2-$3: $(shell find $1 -type f -name '*.go') _vendor
	$$(call echo_invocation)
	GOOS=$2 GOARCH=$$($3_GOARCH) CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$$@ ./$1/cmd/$1
endif
endif

ifeq ($2,linux)
.PHONY build-$1-$2: build-$1-$2-$3
build-$1-$2-$3: bin/$1-$2-$3 _vendor
	$$(call echo_invocation)

bin/$1-$2-$3: $(shell find $1 -type f -name '*.go') _vendor
	$$(call echo_invocation)
	GOOS=$2 GOARCH=$$($3_GOARCH) CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$$@ ./$1/cmd/$1
endif
endef

# Targets for binaries for a single image
# $1: image name (brigade-api)
#
# Targets:
# build-image-bins-$(binary)
define IMAGE_BIN_TARGETS
.PHONY build-image-bins: build-image-bins-$1
build-image-bins-$1:
	$$(call echo_invocation)

$(foreach arch,$(ARCHS),$(eval $(call IMAGE_BIN_ARCH_TARGETS,$1,$(arch))))
endef

# Targets for binaries for a single image arch
# $1: image name (brigade-api)
# $2: arch (amd64)
#
# Targets:
# build-image-bins-$(binary)-$(arch)
define IMAGE_BIN_ARCH_TARGETS
.PHONY build-image-bins-$1: build-image-bins-$1-$2
build-image-bins-$1-$2: $1/rootfs/$2/$1 _vendor
	$$(call echo_invocation)

$1/rootfs/$2/$1: $(shell find $1 -type f -name '*.go') _vendor
	$$(call echo_invocation)
	GOOS=linux GOARCH=$$($2_GOARCH) CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o ./$$@ ./$1/cmd/$1
endef

# Targets for a single image
# $1: image name (git-sidecar)
#
# Targets:
# build-$(binary)-image
# push-$(binary)-image
# create-$(binary)-manifest
define IMAGE_TARGETS
.PHONY build-images: build-$1-images
.PHONY push-images: push-$1-images
.PHONY create-manifests: create-$1-manifests

build-$1-images:
	$$(call echo_invocation)

ifndef SKIP_DOCKER_BUILD
push-$1-images: build-$1-images
else
push-$1-images:
endif
	$$(call echo_invocation)

ifndef SKIP_DOCKER_PUSH
create-$1-manifests: push-$1-images
else
create-$1-manifests:
endif
	$$(call echo_invocation)
	docker manifest create --amend $(DOCKER_REGISTRY)/$1:$(IMAGE_TAG) $(foreach arch,$(ARCHS),$(DOCKER_REGISTRY)/$1:$(arch)-$(IMAGE_TAG))
	$(foreach arch,$(ARCHS),docker manifest annotate $$($(arch)_DOCKER_ANNOTATIONS) $(DOCKER_REGISTRY)/$1:$(IMAGE_TAG) $(DOCKER_REGISTRY)/$1:$(arch)-$(IMAGE_TAG)
	)
	docker manifest push $(DOCKER_REGISTRY)/$1:$(IMAGE_TAG)

$(foreach arch,$(ARCHS),$(eval $(call IMAGE_ARCH_TARGETS,$1,$(arch))))
endef

# Targets for a single image
# $1: image name (git-sidecar)
# $2: arch (amd64)
#
# Targets:
# build-$(binary)-$(arch)-image
# push-$(binary)-$(arch)-image
define IMAGE_ARCH_TARGETS
build-$1-images: build-$1-$2-images
push-$1-images: push-$1-$2-images

.PHONY: build-$1-$2-images
build-$1-$2-images: $1/Dockerfile.$2
	$$(call echo_invocation)
	docker build $(DOCKER_BUILD_FLAGS) -t $(DOCKER_REGISTRY)/$1:$2-$(IMAGE_TAG) -f $1/Dockerfile.$2 $1

ifndef SKIP_DOCKER_BUILD
push-$1-$2-images: build-$1-$2-images 
else
push-$1-$2-images: 
endif
	$$(call echo_invocation)
	docker push $(DOCKER_REGISTRY)/$1:$2-$(IMAGE_TAG)

$1/Dockerfile.$2: $1/Dockerfile _vendor
	$$(call echo_invocation)
	@echo "arch=\"$2\"\ndocker=\"$$($2_DOCKER_ARCH)\"\nqemu=\"$$($2_QEMU_IMPORT)\"" | gomplate -f $1/Dockerfile -c .=stdin:///context.env > ./$$@
endef

$(foreach bin,$(BINS),$(eval $(call BIN_TARGETS,$(bin))))
$(foreach bin,$(BINS),$(eval $(call IMAGE_BIN_TARGETS,$(bin))))
$(foreach img,$(IMAGES),$(eval $(call IMAGE_TARGETS,$(img))))

######################################################################
## HELPERS
######################################################################

_COLOR_RED    := $(shell echo "\033[31m")
_COLOR_GREEN  := $(shell echo "\033[32m")
_COLOR_YELLOW := $(shell echo "\033[33m")
_COLOR_CYAN   := $(shell echo "\033[35m")
_COLOR_GRAY   := $(shell echo "\033[30;1m")
_COLOR_RESET  := $(shell echo "\033[0m")

# Target name
_TARGET_NAME        = $(_COLOR_CYAN)$@$(_COLOR_RESET)
_DEPENDENCIES       = $(_COLOR_GREEN)$(filter-out $?,$^)$(_COLOR_RESET)
_NEWER_TRIGGERS     = $(if $?, $(_COLOR_RED)$?$(_COLOR_RESET))
_INVOCATION         = $(_COLOR_GRAY)[$(_COLOR_RESET)$(_TARGET_NAME)$(_COLOR_GRAY)]$(_COLOR_RESET)
define echo_invocation
$(if $(DEBUG_MAKE),@echo "$(_INVOCATION) $(_DEPENDENCIES)$(_NEWER_TRIGGERS)")
endef

ifdef DEBUG_MAKE
#OLD_SHELL := $(SHELL)
#SHELL = $(warning [$@ ($^) ($?) | ])$(OLD_SHELL)
endif

######################################################################
## DEBUG
######################################################################

.PHONY: list-targets
list-targets:
	$(call echo_invocation)
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | sort | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | xargs -0