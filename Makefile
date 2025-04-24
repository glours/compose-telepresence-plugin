BINARY_NAME = telepresence
BINARY_DIR = bin
HOME_BIN_DIR = $(HOME)/bin
VERSION = $(shell git describe --tags --always)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT)"

.PHONY: all
all: build

.PHONY: build
build:
	mkdir -p $(BINARY_DIR)
	go build $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)

.PHONY: install
install: build
	mkdir -p $(HOME_BIN_DIR)
	cp $(BINARY_DIR)/$(BINARY_NAME) $(HOME_BIN_DIR)/

.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build   - Build the binary in ./bin"
	@echo "  install - Install the binary to ~/bin"
	@echo "  clean   - Remove the binary and bin directory"
