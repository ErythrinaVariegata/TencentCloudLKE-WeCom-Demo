# Binary name
BINARY=lke-wecom-demo

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Build flags
LDFLAGS=-ldflags "-s -w"

# Default OS and ARCH
OS?=$(shell go env GOOS)
ARCH?=$(shell go env GOARCH)

# Output directory
OUT_DIR=build

# Build for current platform
.PHONY: build
build:
	$(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -rf $(OUT_DIR)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Cross compilation targets
.PHONY: build-all
build-all: build-linux build-windows build-darwin

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)-linux-arm64

.PHONY: build-windows
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)-windows-amd64.exe
	GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)-windows-arm64.exe

.PHONY: build-darwin
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)-darwin-arm64

# Custom build target for specific OS and ARCH
.PHONY: build-custom
build-custom:
	GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) $(LDFLAGS) -o $(OUT_DIR)/$(BINARY)-$(OS)-$(ARCH)
