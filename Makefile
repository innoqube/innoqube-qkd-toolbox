# The root directory for the source code of the binaries
CMD_DIR=./cmd

# Build directory
BUILD_DIR=./build

# Go build flags for versioning
LDFLAGS=-ldflags "-s -w"

# Find all subdirectories in cmd/ to get the names of the binaries
BINS=$(shell find $(CMD_DIR) -mindepth 1 -maxdepth 1 -type d | xargs -n 1 basename)

# --- Main Targets ---

.PHONY: all
# Default target, builds all binaries
all: $(addprefix build-,$(BINS))
	@echo "All binaries built successfully."

.PHONY: build
# A generic build target that requires a binary name to be passed
build:
ifndef BINARY
	$(error BINARY is not set. Use 'make build BINARY=<binary_name>')
endif
	@echo "Building $(BINARY)..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_DIR)/$(BINARY)
	@echo "Done!"

.PHONY: run
# A generic run target that requires a binary name
run:
ifndef BINARY
	$(error BINARY is not set. Use 'make run BINARY=<binary_name>')
endif
	@echo "Running $(BINARY)..."
	@$(BUILD_DIR)/$(BINARY)

.PHONY: clean
# Cleans all built binaries and the build directory
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)/*
	@echo "Done!"

# --- Specific Targets for Each Binary ---
# These targets are generated dynamically based on the BINS variable.

# Loop through each binary and create a build target for it
$(foreach bin,$(BINS),$(eval .PHONY: build-$(bin)))
$(foreach bin,$(BINS),$(eval build-$(bin): $(BUILD_DIR)/$(bin)))

# Rule to build a specific binary
$(BUILD_DIR)/%: $(CMD_DIR)/%
	@mkdir -p $(BUILD_DIR)
	@echo "Building $< to $@..."
	@go build $(LDFLAGS) -o $@ $(CMD_DIR)/$(notdir $<)

# Rule to build a specific binary
build-%:
	@echo "Building binary: $*"
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$* $(CMD_DIR)/$*

# --- Help Target ---

.PHONY: help
help:
	@echo "Makefile for a multi-binary Go repository"
	@echo ""
	@echo "Available binaries: $(BINS)"
	@echo ""
	@echo "Usage:"
	@echo "  make all                     - Builds all binaries"
	@echo "  make build BINARY=<name>     - Builds a specific binary (e.g., 'make build BINARY=server')"
	@echo "  make run BINARY=<name>       - Runs a specific binary"
	@echo "  make clean                   - Cleans all binaries and the build directory"
	@echo ""