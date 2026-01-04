BINARY_NAME=qv
BUILD_DIR=bin

.PHONY: build clean test

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/qv

clean:
	rm -rf $(BUILD_DIR)

test:
	go test -v ./...
