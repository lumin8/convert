TIMESTAMP := $(shell date +%s)
BUILD_USER := $(shell id -un)
BUILD_HOST := $(shell hostname)
GIT_SHA := $(shell git rev-parse --short HEAD)

BINARY := dataman_server
VERSION ?= $(TIMESTAMP)_$(GIT_SHA)

PLATFORMS := windows linux darwin
os = $(word 1, $@)

# Usage: $make PLATFORM_NAME (ex $make darwin, $make VERSION=1.0.0 darwin) 
$(PLATFORMS):
	@mkdir -p release
	GOOS=$(os) GOARCH=amd64 go build -ldflags "\
		-X main.BUILDTIMESTAMP=$(TIMESTAMP) \
		-X main.BUILDUSER=$(BUILD_USER) \
		-X main.BUILDHOST=$(BUILD_HOST) \
		-X main.BUILDGITSHA=$(GIT_SHA)" \
		-o ./release/$(BINARY)-$(VERSION)-$(os) ./dataman/

.PHONY: $(PLATFORMS)
