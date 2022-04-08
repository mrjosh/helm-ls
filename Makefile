export ROOT=$(realpath $(dir $(firstword $(MAKEFILE_LIST))))
export GOBIN?=$(BIN)
export BIN=$(ROOT)/bin
export GO=$(shell which go)
export CGO_ENABLED=1
export GOX=$(BIN)/gox

$(eval GIT_COMMIT=$(shell git rev-parse --short HEAD))
$(eval BRANCH_NAME=$(shell git rev-parse --abbrev-ref HEAD))

export GO_LDFLAGS="-X main.CompiledBy=runner@Mac-1640963667289.local -X main.Version=${GIT_COMMIT} -X main.BranchName=${BRANCH_NAME} -X main.BuildTime=`date -u '+%Y-%m-%d_%I:%M:%S%p'`"

export LINTER=$(GOBIN)/golangci-lint
export LINTERCMD=run --no-config -v \
	--print-linter-name \
	--skip-files ".*.gen.go" \
	--skip-files ".*_test.go" \
	--sort-results \
	--disable-all \
	--enable=structcheck \
	--enable=deadcode \
	--enable=gocyclo \
	--enable=ineffassign \
	--enable=revive \
	--enable=goimports \
	--enable=errcheck \
	--enable=varcheck \
	--enable=goconst \
	--enable=megacheck \
	--enable=misspell \
	--enable=unused \
	--enable=typecheck \
	--enable=staticcheck \
	--enable=govet \
	--enable=gosimple

all:
	@$(GO) build -ldflags ${GO_LDFLAGS} -o bin/helm_lint_ls .

# lint runs vet plus a number of other checkers, it is more comprehensive, but louder
lint:
	@LINTER_BIN=$$(command -v $(LINTER)) || { echo "golangci-lint command not found! Installing..." && $(MAKE) install-metalinter; };
	@$(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/ \
		| xargs $(LINTER) $(LINTERCMD) ./...; if [ $$? -eq 1 ]; then \
			echo ""; \
			echo "Lint found suspicious constructs. Please check the reported constructs"; \
			echo "and fix them if necessary before submitting the code for reviewal."; \
		fi

.PHONY: gox
gox:
	@gox -output="dist/homepi_{{.OS}}_{{.Arch}}"

# for ci jobs, runs lint against the changed packages in the commit
ci-lint:
	@$(shell which golangci-lint) $(LINTERCMD) --deadline 10m ./...

# Check if golangci-lint not exists, then install it
install-metalinter:
	@$(GO) get -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1
	@$(GO) install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1

test:
	$(GO) test ./... -v -race

install-gox:
	@$(GO) install github.com/mitchellh/gox@v1.0.1

.PHONY: build-linux
build-linux: install-gox
	@$(GOX) -ldflags ${GO_LDFLAGS} --arch=amd64 --os=linux --output="dist/homepi_{{.OS}}_{{.Arch}}"
	@$(GOX) -ldflags ${GO_LDFLAGS} --arch=arm --os=linux --output="dist/homepi_{{.OS}}_{{.Arch}}"

.PHONY: build-macOS
build-macOS: install-gox
	@$(GOX) -ldflags ${GO_LDFLAGS} --arch=amd64 --os=darwin --output="dist/homepi_{{.OS}}_{{.Arch}}"

.PHONY: build-artifacts
build-artifacts:
	@$(MAKE) build-linux && $(MAKE) build-macOS
