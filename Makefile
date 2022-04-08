export ROOT=$(realpath $(dir $(firstword $(MAKEFILE_LIST))))
export BIN=$(ROOT)/bin
export GO=$(shell which go)

all:
	$(GO) build -o ${BIN}/helm_lint_ls ./...

# lint runs vet plus a number of other checkers, it is more comprehensive, but louder
lint:
	@LINTER_BIN=$$(command -v $(LINTER)) || { echo "golangci-lint command not found! Installing..." && $(MAKE) install-metalinter; };
	@$(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/ \
		| xargs $(LINTER) $(LINTERCMD) ./pkg/...; if [ $$? -eq 1 ]; then \
			echo ""; \
			echo "Lint found suspicious constructs. Please check the reported constructs"; \
			echo "and fix them if necessary before submitting the code for reviewal."; \
		fi

# for ci jobs, runs lint against the changed packages in the commit
ci-lint:
	@$(shell which golangci-lint) $(LINTERCMD) --deadline 10m ./...

# Check if golangci-lint not exists, then install it
install-metalinter:
	@$(GO) get -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1
	@$(GO) install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1

test:
	$(GO) test ./... -v -race
