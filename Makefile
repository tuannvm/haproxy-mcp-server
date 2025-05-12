.PHONY: build test clean run-dev release-snapshot run-docker run docker-compose-up docker-compose-down lint

# Variables
BINARY_NAME ?= $(shell grep -m 1 '^module ' go.mod 2>/dev/null | sed 's/^module github.com\/[^\/]*\/\([^\/]*\).*/\1/' || basename $$(pwd))
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=bin
NPM_VERSION ?= $(shell echo $(VERSION) | sed 's/^v//')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMON_BUILD_ARGS = -ldflags "-X github.com/tuannvm/haproxy-mcp-server/internal/version.Version=$(VERSION) -X github.com/tuannvm/haproxy-mcp-server/internal/version.CommitHash=$(COMMIT_HASH) -X github.com/tuannvm/haproxy-mcp-server/internal/version.BuildTime=$(BUILD_TIME) -X github.com/tuannvm/haproxy-mcp-server/internal/version.BinaryName=$(BINARY_NAME)"

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
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME) $(foreach os,$(OSES),$(foreach arch,$(ARCHS),$(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,)))
	find ./npm -name ".npmrc" -type f -delete
	find ./npm -name "tmp.json" -type f -delete

.PHONY: format
format: ## Format the code
	go fmt ./...

.PHONY: tidy
tidy: ## Tidy up the go modules
	go mod tidy

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
	@echo "Building for all platforms..."
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		echo "Building for $(os)/$(arch)"; \
		GOOS=$(os) GOARCH=$(arch) go build $(COMMON_BUILD_ARGS) -o $(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,) ./cmd/server; \
	))

.PHONY: npm
npm: build-all-platforms ## Create the npm packages
	# Create binaries directory in the main package
	mkdir -p ./npm/$(BINARY_NAME)/bin/binaries
	
	# Copy all binaries to the main package
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		EXECUTABLE=./$(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,); \
		BINARY_NAME_WITH_SUFFIX=$(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,); \
		cp $$EXECUTABLE ./npm/$(BINARY_NAME)/bin/binaries/$$BINARY_NAME_WITH_SUFFIX; \
	))
	
	# Copy README.md to the npm package
	cp README.md ./npm/$(BINARY_NAME)/
	
	# Ensure index.js is executable
	chmod +x ./npm/$(BINARY_NAME)/bin/index.js

.PHONY: npm-publish
npm-publish: npm ## Publish the npm package
	@if [ -z "$$NPM_TOKEN" ]; then \
		echo "Error: NPM_TOKEN environment variable is not set"; \
		echo "Please set it with: export NPM_TOKEN=your_npm_token"; \
		exit 1; \
	fi
	
	# Set version in package.json
	cd npm/$(BINARY_NAME); \
	echo "//registry.npmjs.org/:_authToken=$$NPM_TOKEN" > .npmrc; \
	jq '.version = "$(NPM_VERSION)"' package.json > tmp.json && mv tmp.json package.json; \
	echo "Publishing version $(NPM_VERSION)..."; \
	npm publish --access public; \
	cd ../..

.PHONY: npm-publish-dry-run
npm-publish-dry-run: npm ## Test npm packaging without actually publishing
	@echo "Performing a dry run of npm packaging process..."
	cd npm/$(BINARY_NAME); \
	jq '.version = "$(NPM_VERSION)"' package.json > tmp.json && mv tmp.json package.json; \
	npm pack; \
	cd ../..

.PHONY: npm-login
npm-login: ## Login to npm registry
	@echo "Logging in to npm registry..."
	@npm login

# Default target
all: clean build
