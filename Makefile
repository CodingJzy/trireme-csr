PROJECT_NAME := trireme-csr
VERSION_FILE := ./version/version.go
VERSION := 0.11
REVISION=$(shell git log -1 --pretty=format:"%H")
BUILD_NUMBER := latest
DOCKER_REGISTRY?=aporeto
DOCKER_IMAGE_NAME?=$(PROJECT_NAME)
DOCKER_IMAGE_TAG?=$(BUILD_NUMBER)

codegen:
	echo 'package version' > $(VERSION_FILE)
	echo '' >> $(VERSION_FILE)
	echo '// VERSION is the version of Trireme-CSR' >> $(VERSION_FILE)
	echo 'const VERSION = "$(VERSION)"' >> $(VERSION_FILE)
	echo '' >> $(VERSION_FILE)
	echo '// REVISION is the revision of Trireme-CSR' >> $(VERSION_FILE)
	echo 'const REVISION = "$(REVISION)"' >> $(VERSION_FILE)

build: codegen
	CGO_ENABLED=1 go build -a -v -installsuffix cgo

build_ca-manager: codegen
	cd cmd/ca-manager && CGO_ENABLED=1 go build -a -v -installsuffix cgo

package: build build_ca-manager
	mv trireme-csr docker/trireme-csr
	mv cmd/ca-manager/ca-manager docker/ca-manager

docker_build: package
	docker \
		build \
		-t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) docker

docker_push: docker_build
	docker \
		push \
		$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)
