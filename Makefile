.PHONY: build test test-integration lint clean install check

# Build the myruntime binary
build:
	go build -o bin/myruntime ./cmd/myruntime

# Run unit tests
test:
	go test -v ./internal/...

# Run integration tests (requires root, Linux)
test-integration:
	sudo go test -v -count=1 ./tests/

# Clean build artifacts
clean:
	rm -rf bin/

# Install to /usr/local/bin (requires root)
install: build
	sudo cp bin/myruntime /usr/local/bin/

# Run linter
lint:
	golangci-lint run ./...

# Verify Linux environment has required kernel features
check:
	@echo "Checking environment..."
	@uname -s | grep -q Linux || (echo "ERROR: Linux required" && exit 1)
	@test -f /sys/fs/cgroup/cgroup.controllers || (echo "ERROR: cgroup v2 not mounted" && exit 1)
	@echo "Kernel: $$(uname -r)"
	@echo "Cgroup controllers: $$(cat /sys/fs/cgroup/cgroup.controllers)"
	@echo "Environment OK"
