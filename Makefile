export ROOT=$(realpath $(dir $(firstword $(MAKEFILE_LIST))))
export BIN=$(ROOT)/bin
export GOBIN?=$(BIN)
export GO=$(shell which go)
export PACKAGE_NAME=github.com/mrjosh/helm-ls
export GOLANG_CROSS_VERSION=v1.20.6

$(eval GIT_COMMIT=$(shell git rev-parse --short HEAD))
$(eval BRANCH_NAME=$(shell git rev-parse --abbrev-ref HEAD))
$(eval COMPILED_BY=$(shell hostname))
$(eval BUILD_TIME=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p'))

GO_LDFLAGS := -X "main.CompiledBy=${COMPILED_BY}" -X "main.Version=${GIT_COMMIT}" -X "main.BranchName=${BRANCH_NAME}" -X "main.BuildTime=${BUILD_TIME}"

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

test:
	@$(GO) test ./... -v -race

.PHONY: build-release
build-release:
	@docker run \
			--rm \
			-e CGO_ENABLED=1 \
			-e COMPILED_BY=$(COMPILED_BY) \
			-e BRANCH_NAME=$(BRANCH_NAME) \
			-e BUILD_TIME=$(BUILD_TIME) \
			-e GIT_COMMIT=$(GIT_COMMIT) \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v `pwd`:/go/src/$(PACKAGE_NAME) \
			-w /go/src/$(PACKAGE_NAME) \
			ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
			--clean --skip-validate --skip-publish --snapshot
