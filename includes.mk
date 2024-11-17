BIN ?= bin
BIN_DIR ?= $(join $(dir $(lastword $(MAKEFILE_LIST))), $(BIN))
LOCAL_BIN := $(CURDIR)/bin

GIT_BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
GIT_HASH ?= $(shell git rev-parse --short HEAD)
GIT_TAG_HASH ?=

VERSION = $(GIT_BRANCH)-$(GIT_HASH)
BUILD_VERSION ?= 

GOLANGCI_LINT_BIN := $(LOCAL_BIN)/golangci-lint
GOLANGCI_LINT_VERSION := v1.62.0

GO = go
GO_FLAGS ?=
GO_LDFLAGS ?= -ldflags="-X 'main.Version=$(BUILD_VERSION)'"

# go_install_util make install a binary from a golang module.
# Parameters:
# 1 - module uri for building;
# 2 - module version in semver format (https://semver.org/) or 'latest';
# 3 - full path to install the binary.
# 4 - build flags (optional)
# It does not work through go install, it is needed to be able to use different versions in different services.
# Checks if binary file exists, creates a temp directory, make a fake module in it, in which it calls installation and building.
define go_install_util
	@[ ! -f $(3)@$(2) ] \
		|| exit 0 \
		&& echo "Installing $(1)@$(2) ..." \
		&& tmp=$$(mktemp -d) \
		&& cd $$tmp \
		&& echo "Module: $(1)" \
		&& echo "Version: $(2)" \
		&& echo "Binary: $(3)" \
		&& go mod init temp && go get -d $(1)@$(2) && go build $(4) -o $(3)@$(2) $(1) \
		&& ln -sf $(3)@$(2) $(3) \
		&& rm -rf $$tmp \
		&& echo "$(3) has been installed!" \
		&& echo "=========================================="
endef