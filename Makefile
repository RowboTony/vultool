# Makefile for vultool

.PHONY: build test clean install fixtures help

# Default target
help:	## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

build:	## Build the vultool binary
	go build -ldflags "-X main.version=$$(cat VERSION)" -o vultool ./cmd/vultool

install:	## Install vultool to GOPATH/bin
	go install -ldflags "-X main.version=$$(cat VERSION)" ./cmd/vultool

test:	## Run all tests
	go test ./...

fixtures:	## Initialize/update test fixtures submodule
	git submodule init
	git submodule update

demo:	## Run a demo with test fixtures (requires fixtures)
	@echo "Testing vultool with GG20 fixture..."
	./vultool inspect -f test/fixtures/testGG20-part1of2.vult --summary
	@echo "\nTesting vultool with DKLS fixture..."
	./vultool inspect -f test/fixtures/testDKLS-1of2.vult --summary

validate:	## Validate all test fixtures
	@echo "Validating all test fixtures..."
	@for file in test/fixtures/*.vult; do \
		echo "Validating $$file..."; \
		case "$$file" in \
			*qa-fast-share2of2.vult) \
				./vultool inspect -f "$$file" --validate --password "vulticli01" || exit 1 ;; \
			*) \
				./vultool inspect -f "$$file" --validate || exit 1 ;; \
		esac; \
	done
	@echo "All fixtures validated successfully!"

clean:	## Clean build artifacts
	rm -f vultool

deps:	## Download and tidy Go dependencies
	go mod download
	go mod tidy

format:	## Format Go code
	go fmt ./...

lint:	## Run Go linter (requires golangci-lint)
	golangci-lint run

# Development workflow
dev: fixtures deps build demo	## Full development setup: fixtures + deps + build + demo

# CI/CD related targets
.PHONY: ci-local security-scan build-all-platforms setup-hooks validate-local coverage benchmark release-check clean-ci

ci-local: deps format lint test security-scan build validate ## Run full CI suite locally

security-scan: ## Run security scans (gosec + govulncheck)
	@echo "Running gosec security scan..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest; \
	fi
	gosec -fmt json -out gosec-report.json ./... || echo "gosec scan completed with warnings"
	
	@echo "Running govulncheck..."
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "Installing govulncheck..."; \
		GOPROXY=direct go install golang.org/x/vuln/cmd/govulncheck@latest; \
	fi
	govulncheck ./... || echo "govulncheck scan completed with warnings"

build-all-platforms: ## Build for all supported platforms
	@echo "Building for all platforms..."
	@mkdir -p dist
	@VERSION=$$(cat VERSION) && \
	for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ] && [ "$$arch" = "arm64" ]; then continue; fi; \
			echo "Building for $$os/$$arch..."; \
			GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build \
				-ldflags "-X main.version=$$VERSION -s -w" \
				-o "dist/vultool-$$VERSION-$$os-$$arch$$([ "$$os" = "windows" ] && echo .exe)" \
				./cmd/vultool; \
		done; \
	done

setup-hooks: ## Install pre-commit hooks
	@echo "Setting up pre-commit hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/bash' > .git/hooks/pre-commit
	@echo 'set -e' >> .git/hooks/pre-commit
	@echo 'echo "Running pre-commit checks..."' >> .git/hooks/pre-commit
	@echo 'make format lint test' >> .git/hooks/pre-commit
	@echo 'echo "Pre-commit checks passed!"' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hooks installed successfully!"

validate-local: build ## Validate build matches CI
	@echo "Running local validation..."
	@./vultool --version
	@if [ -d "test/fixtures" ] && [ "$$(ls -A test/fixtures 2>/dev/null || true)" ]; then \
		echo "Running fixture validation..."; \
		make validate; \
	else \
		echo "No fixtures available, skipping fixture validation"; \
	fi
	@echo "Local validation complete!"

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | tail -1

benchmark: ## Run benchmark tests
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@make ci-local
	@git status --porcelain | grep -q . && echo "❌ Working directory not clean" && exit 1 || true
	@echo "✅ Ready for release"

clean-ci: ## Clean CI artifacts
	@echo "Cleaning CI artifacts..."
	@rm -f coverage.out coverage.html gosec-report.json
	@rm -rf dist/
	@echo "CI artifacts cleaned"
