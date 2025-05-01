export ROOT=$(realpath $(dir $(firstword $(MAKEFILE_LIST))))
export BIN=$(ROOT)/bin
export GOBIN?=$(BIN)
export GO=$(shell which go)
export PACKAGE_NAME=github.com/mrjosh/helm-ls
export GOLANG_CROSS_VERSION=v1.22.7
export CGO_ENABLED=1

$(eval GIT_COMMIT=$(shell git rev-parse --short HEAD))
$(eval BRANCH_NAME=$(shell git symbolic-ref -q --short HEAD || git describe --tags --exact-match))
$(eval COMPILED_BY=$(shell hostname))
$(eval BUILD_TIME=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p'))

GO_LDFLAGS :=" -X main.Version=${BRANCH_NAME} -X main.CompiledBy=${COMPILED_BY} -X main.GitCommit=${GIT_COMMIT} -X main.Branch=${BRANCH_NAME} -X main.BuildTime=${BUILD_TIME}"

export TEST_RUNNER=$(GOBIN)/gotestsum
export LINTER=$(GOBIN)/golangci-lint
export LINTERCMD=run --no-config -v \
	--print-linter-name \
	--skip-files ".*.gen.go" \
	--skip-files ".*_test.go" \
	--sort-results \
	--disable-all \
	--enable=gocyclo \
	--enable=ineffassign \
	--enable=revive \
	--enable=errcheck \
	--enable=goconst \
	--enable=megacheck \
	--enable=misspell \
	--enable=unused \
	--enable=typecheck \
	--enable=staticcheck \
	--enable=govet \
	--enable=gosimple

all:
	@$(GO) build -ldflags ${GO_LDFLAGS} -o bin/helm_ls .

# lint runs vet plus a number of other checkers, it is more comprehensive, but louder
lint:
	@LINTER_BIN=$$(command -v $(LINTER)) || { echo "golangci-lint command not found! Installing..." && $(MAKE) install-metalinter; };
	@$(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/ \
		| xargs $(LINTER) $(LINTERCMD) ./...; if [ $$? -eq 1 ]; then \
			echo ""; \
			echo "Lint found suspicious constructs. Please check the reported constructs"; \
			echo "and fix them if necessary before submitting the code for reviewal."; \
		fi

# for ci jobs, runs lint against the changed packages in the commit
ci-lint:
	@$(LINTER) $(LINTERCMD) --deadline 10m ./...

# Check if golangci-lint not exists, then install it
install-metalinter:
	@$(GO) get -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.2
	@$(GO) install -v github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.2

install-testrunner:
	@$(GO) install -v gotest.tools/gotestsum@latest

install-yamlls:
	npm install --global yaml-language-server@1.15

integration-test-deps:
	@YAMLLS_BIN=$$(command -v yaml-language-server) || { echo "yaml-language-server command not found! Installing..." && $(MAKE) install-yamlls; };
	git submodule init
	git submodule update --depth 1

define run-tests
	@$(TEST_RUNNER) ./... -v -race -tags=integration || { \
		echo "gotestsum command not found or failed! Falling back to 'go test'..."; \
		$(GO) test ./... -v -race -tags=integration; \
	}
endef

test:
	$(MAKE) integration-test-deps
	$(call run-tests)

test-update-snaps: 
	$(MAKE) integration-test-deps
	UPDATE_SNAPS=true $(call run-tests)

coverage:
	@$(GO) test -coverprofile=.coverage -tags=integration -coverpkg=./internal/... ./internal/... && go tool cover -html=.coverage

.PHONY: build-release
build-release:
	@docker run \
			--rm \
			-e CGO_ENABLED=1 \
			-e COMPILED_BY=$(COMPILED_BY) \
			-e VERSION=$(BRANCH_NAME) \
			-e BRANCH_NAME=$(BRANCH_NAME) \
			-e BUILD_TIME=$(BUILD_TIME) \
			-e GIT_COMMIT=$(GIT_COMMIT) \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v `pwd`:/go/src/$(PACKAGE_NAME) \
			-w /go/src/$(PACKAGE_NAME) \
			ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
			--clean --skip=validate,publish --snapshot
