GO_BIN=$(shell which go)

# Build metadata baked into the binary via internal/build.
VERSION=$(shell git describe --tags --always)
COMMIT=$(shell git rev-parse --short HEAD)
BUILT_AT=$(shell date +%FT%T%z)
BUILT_BY=$(USER)
BUILT_ON=$(shell hostname)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_PKG=github.com/goodblaster/vertigo/internal/build
LDFLAGS=-X $(BUILD_PKG).Version=$(VERSION) -X $(BUILD_PKG).Branch=$(BRANCH) -X $(BUILD_PKG).Commit=$(COMMIT) -X $(BUILD_PKG).Date=$(BUILT_AT) -X $(BUILD_PKG).User=$(BUILT_BY) -X $(BUILD_PKG).Host=$(BUILT_ON)

.PHONY: all build test vet clean amd64 arm64 native

all: amd64 arm64

build:
	$(GO_BIN) build -ldflags="$(LDFLAGS)" -o vertigo .

test:
	$(GO_BIN) test ./...

vet:
	$(GO_BIN) vet ./...

amd64:
	@echo "Building for amd64..."
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux $(GO_BIN) build -ldflags="$(LDFLAGS)" -o vertigo .
	@echo "Creating amd64 zip artifact..."
	zip vertigo_$(VERSION)_linux_amd64.zip vertigo

arm64:
	@echo "Building for arm64..."
	CGO_ENABLED=0 GOARCH=arm64 GOOS=linux $(GO_BIN) build -ldflags="$(LDFLAGS)" -o vertigo .
	@echo "Creating arm64 zip artifact..."
	zip vertigo_$(VERSION)_linux_arm64.zip vertigo

native:
	@echo "Building for native..."
	$(GO_BIN) build -ldflags="$(LDFLAGS)" -o vertigo .
	@echo "Creating native zip artifact..."
	zip vertigo.zip vertigo

clean:
	@echo "Cleaning up..."
	rm -f vertigo vertigo*.zip
