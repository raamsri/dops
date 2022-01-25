.DEFAULT_GOAL := build

.PHONY: go-lint

# Run the golangci-lint tool
go-lint:
	golangci-lint run --timeout=15m ./...

.PHONY: lint

# Run all the linters
lint: go-lint

# The build targets allow to build the binary and docker image
.PHONY: build build.docker build.mini

BINARY        ?= dops
SOURCES        = $(shell find . -name '*.go')
IMAGE_STAGING  = 707342285240.dkr.ecr.ap-south-1.amazonaws.com/$(BINARY)-staging
IMAGE         ?= 707342285240.dkr.ecr.ap-south-1.amazonaws.com/$(BINARY)
VERSION       ?= $(shell git describe --tags --always --dirty)
BUILD_FLAGS   ?= -v
LDFLAGS       ?= -X github.com/toppr-systems/dops/dops.Version=$(VERSION) -w -s
ARCHS         = amd64
SHELL         = /bin/bash


build: build/$(BINARY)

build/$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build.push/multiarch:
	arch_specific_tags=()
	for arch in $(ARCHS); do \
		image="$(IMAGE):$(VERSION)-$${arch}" ;\
		docker pull $${arch}/alpine:3.14 ;\
		docker pull golang:1.16 ;\
		DOCKER_BUILDKIT=1 docker build --rm --tag $${image} --build-arg VERSION="$(VERSION)" --build-arg ARCH="$${arch}" . ;\
		docker push $${image} ;\
		arch_specific_tags+=( "--amend $${image}" ) ;\
	done ;\
	DOCKER_CLI_EXPERIMENTAL=enabled docker manifest create "$(IMAGE):$(VERSION)" $${arch_specific_tags[@]} ;\
	for arch in $(ARCHS); do \
		DOCKER_CLI_EXPERIMENTAL=enabled docker manifest annotate --arch $${arch} "$(IMAGE):$(VERSION)" "$(IMAGE):$(VERSION)-$${arch}" ;\
	done;\
	DOCKER_CLI_EXPERIMENTAL=enabled docker manifest push "$(IMAGE):$(VERSION)" \

build.push: build.docker
	docker push "$(IMAGE):$(VERSION)"

build.arm64v8:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build.amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build.darwin.amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build.arm32v7:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o build/$(BINARY) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" .

build.docker:
	docker build --rm --tag "$(IMAGE):$(VERSION)" --build-arg VERSION="$(VERSION)" --build-arg ARCH="amd64" .

build.mini:
	docker build --rm --tag "$(IMAGE):$(VERSION)-mini" --build-arg VERSION="$(VERSION)" -f Dockerfile.mini .

clean:
	@rm -rf build

 # Builds and push container images to the staging bucket.
.PHONY: release.staging

release.staging:
	IMAGE=$(IMAGE_STAGING) $(MAKE) build.push/multiarch

release.prod:
	$(MAKE) build.push/multiarch
