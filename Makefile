.PHONY: build run test lint clean install

BINARY_NAME=anitui
CMD_DIR=./cmd/anitui
BUILD_DIR=./build

# Use PowerShell for cross-platform compatibility if needed, 
# but simple go commands work fine.
# For 'clean', we use a conditional to handle Windows/Unix.

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(CMD_DIR)

run:
	go run $(CMD_DIR)

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	golangci-lint run ./...

clean:
	@if exist $(BUILD_DIR) (rd /s /q $(BUILD_DIR))

install:
	go install $(CMD_DIR)

build-linux:
	set GOOS=linux&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)

build-windows:
	set GOOS=windows&& set GOARCH=amd64&& go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)

build-all: build-linux build-windows
