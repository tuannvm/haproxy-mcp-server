.PHONY: build test clean run-dev release-snapshot run-docker run docker-compose-up docker-compose-down lint

# Variables
BINARY_NAME ?= $(shell grep -m 1 '^module ' go.mod 2>/dev/null | sed 's/^module github.com\/[^\/]*\/\([^\/]*\).*/\1/' || basename $$(pwd))
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=bin
NPM_VERSION ?= $(shell echo $(VERSION) | sed 's/^v//')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Define OS and architecture combinations
OSES = darwin linux windows
ARCHS = amd64 arm64

# Define clean targets
CLEAN_TARGETS := $(BINARY_NAME)
CLEAN_TARGETS += $(foreach os,$(OSES),$(foreach arch,$(ARCHS),$(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,)))
CLEAN_TARGETS += $(foreach os,$(OSES),$(foreach arch,$(ARCHS),./npm/$(BINARY_NAME)-$(os)-$(arch)/bin/))
CLEAN_TARGETS += ./npm/.npmrc
CLEAN_TARGETS += $(foreach os,$(OSES),$(foreach arch,$(ARCHS),./npm/$(BINARY_NAME)-$(os)-$(arch)/.npmrc))

# Build the application
build:
	mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR) $(CLEAN_TARGETS)

# Run the application in development mode
run-dev:
	go run cmd/server/main.go

# Create a release snapshot using GoReleaser
release-snapshot:
	goreleaser release --snapshot --clean

# Run the application using the built binary
run:
	./$(BUILD_DIR)/$(BINARY_NAME)

# Build and run Docker image
run-docker: build
	docker build -t $(BINARY_NAME):$(VERSION) .
	docker run -p 9097:9097 $(BINARY_NAME):$(VERSION)

# Start the application with Docker Compose
docker-compose-up:
	docker-compose up -d

# Stop Docker Compose services
docker-compose-down:
	docker-compose down

# Run linting checks (same as CI)
lint:
	@echo "Running linters..."
	@go mod tidy
	@if ! git diff --quiet go.mod go.sum; then echo "go.mod or go.sum is not tidy, run 'go mod tidy'"; git diff go.mod go.sum; exit 1; fi
	@if ! command -v golangci-lint &> /dev/null; then echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; fi
	@golangci-lint run --timeout=5m


.PHONY: build-all-platforms
build-all-platforms: clean tidy format ## Build the project for all platforms
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		GOOS=$(os) GOARCH=$(arch) go build $(COMMON_BUILD_ARGS) -o $(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,) ./cmd/$(BINARY_NAME); \
	))

.PHONY: npm
npm: build-all-platforms ## Create the npm packages
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		EXECUTABLE=./$(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,); \
		DIRNAME=$(BINARY_NAME)-$(os)-$(arch); \
		mkdir -p ./npm/$$DIRNAME/bin; \
		cp $$EXECUTABLE ./npm/$$DIRNAME/bin/; \
	))

.PHONY: npm-publish
npm-publish: npm ## Publish the npm packages
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		DIRNAME="$(BINARY_NAME)-$(os)-$(arch)"; \
		cd npm/$$DIRNAME; \
		echo '//registry.npmjs.org/:_authToken=\$(NPM_TOKEN)' >> .npmrc; \
		jq '.version = "$(NPM_VERSION)"' package.json > tmp.json && mv tmp.json package.json; \
		npm publish --access public; \
		cd ../..; \
	))

# Default target
all: clean build
