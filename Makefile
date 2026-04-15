.PHONY: test build clean

# CGO must be enabled for go-sqlite3 FTS5 features
export CGO_ENABLED=1

build:
	@echo "=> Building WaCLI (CGO Enabled)..."
	go build -ldflags="-s -w" -o dist/wacli ./cmd/wacli

test:
	@echo "=> Running Edge Case Suites..."
	go test -v -race -tags "sqlite_fts5" ./...

clean:
	@echo "=> Cleaning artifacts..."
	rm -rf dist/
	go clean -testcache
