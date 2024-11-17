include includes.mk

APP = trace-sidecar

.PHONY: golangci-lint-install lint install help run build build-image
.DEFAULT_GOAL := help

golangci-lint-install: ## Install golangci-lint
	$(call go_install_util,github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION),$(GOLANGCI_LINT_BIN))

lint: golangci-lint-install ## run linter
	$(GOLANGCI_LINT_BIN) run ./...

install: ## download dependencies
	@$(GO) mod download > /dev/null >&1

help:
	@grep -hE '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

run: ## run application
	@echo "=> run $(APP) $(VERSION)"
	@$(GO) run $(CURDIR)/cmd/$(APP)/

build: ## build application
	@echo "=> building $(APP) $(BUILD_VERSION)"
	@$(GO_FLAGS) $(GO) build $(GO_LDFLAGS) -o $(BIN_DIR)/$(APP) $(CURDIR)/cmd/$(APP)/

build-image:
	docker build -t $(APP) --build-arg BUILD_VERSION=$(VERSION) .