MODULE_NAME := $(shell go list -m)
VERSION_PKG := $(MODULE_NAME)/version

BINARY_NAME := bin/osminokal

BUILD_PATH  := ./main.go

# Get current version, commit, and build date
GIT_TAG := $(shell git describe --tags --always)
GIT_COMMIT := $(shell git rev-parse HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Linker flags to embed version information
LDFLAGS := -X $(VERSION_PKG).Version=$(GIT_TAG)
LDFLAGS += -X $(VERSION_PKG).Commit=$(GIT_COMMIT)
LDFLAGS += -X $(VERSION_PKG).BuildDate=$(BUILD_DATE)

# --- Targets ---

.PHONY: build run clean

build:
	@echo "Building $(BINARY_NAME) (Version: $(GIT_TAG), Commit: $(GIT_COMMIT))"
	go build -ldflags "-w -s $(LDFLAGS)" -o $(BINARY_NAME) $(BUILD_PATH)

run: build
	./$(BINARY_NAME)

clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
