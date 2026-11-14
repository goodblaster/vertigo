GO_BIN=$(shell which go)

# Build metadata baked into the binary via internal/build.
VERSION=$(shell git describe --tags --always)
COMMIT=$(shell git rev-parse --short HEAD)
BUILT_AT=$(shell date +%FT%T%z)
BUILD_PKG=github.com/goodblaster/vertigo/internal/build
LDFLAGS=-X $(BUILD_PKG).Version=$(VERSION) -X $(BUILD_PKG).Commit=$(COMMIT) -X $(BUILD_PKG).Date=$(BUILT_AT)

.PHONY: all build test vet clean amd64 arm64 native

# Release artifacts: linux zips, cross-compiled under dist/ so the native
# ./vertigo binary is never overwritten.
all: amd64 arm64

build:
	$(GO_BIN) build -ldflags="$(LDFLAGS)" -o vertigo .

test:
	$(GO_BIN) test ./...

vet:
	$(GO_BIN) vet ./...

amd64 arm64:
	@echo "Building for linux/$@..."
	CGO_ENABLED=0 GOARCH=$@ GOOS=linux $(GO_BIN) build -ldflags="$(LDFLAGS)" -o dist/linux_$@/vertigo .
	@echo "Creating $@ zip artifact..."
	zip -j vertigo_$(VERSION)_linux_$@.zip dist/linux_$@/vertigo

native: build
	@echo "Creating native zip artifact..."
	zip vertigo.zip vertigo

clean:
	@echo "Cleaning up..."
	rm -rf vertigo vertigo*.zip dist
