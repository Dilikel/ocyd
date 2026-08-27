BINARY_NAME=ocyd
CLI_NAME=ocydctl
BUILD_DIR=bin
MAIN_DAEMON=cmd/ocyd/main.go
MAIN_CLI=cmd/ocydctl/main.go

.DEFAULT_GOAL := build

.PHONY: all build fmt vet lint test clean

all: fmt vet lint test build

build: vet fmt 
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_DAEMON)

run: build 
	./bin/ocyd

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)


