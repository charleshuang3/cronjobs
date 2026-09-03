# List available recipes
default:
    @just --list

# Build cronjobs binary
build:
    go build -o bin/cronjobs ./cmd/cronjobs

# Lint Go code using golangci-lint
lint:
    golangci-lint run ./...

# Run Go tests
test:
    go test -v ./...

# Format Go code using goimports
fmt:
    goimports -w -local "github.com/charleshuang3/cronjobs" .

# Check Go code formatting using goimports
fmt-check:
    @test -z "$($(go env GOPATH)/bin/goimports -local github.com/charleshuang3/cronjobs -l . 2>/dev/null || goimports -local github.com/charleshuang3/cronjobs -l .)" || (echo "Unformatted Go files found:" && goimports -local github.com/charleshuang3/cronjobs -l . && exit 1)

# Run go mod tidy
tidy:
    go mod tidy

# Update Go dependencies and tidy
update-deps:
    go get -u ./...
    @just tidy

# Clean build artifacts
clean:
    rm -rf bin
