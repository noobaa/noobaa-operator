# Required build-tools:
#   - go
#   - git
#   - python3
#   - minikube
#   - operator-sdk

export GO111MODULE=on
export GOPROXY:=https://proxy.golang.org

TIME ?= time -p

VERSION ?= $(shell go run cmd/version/main.go)
IMAGE ?= noobaa/noobaa-operator:$(VERSION)
REPO ?= github.com/noobaa/noobaa-operator
CATALOG_IMAGE ?= noobaa/noobaa-operator-catalog:$(VERSION)

GO_LINUX ?= GOOS=linux GOARCH=amd64
GOHOSTOS ?= $(shell go env GOHOSTOS)

OUTPUT ?= build/_output
BIN ?= $(OUTPUT)/bin
OLM ?= $(OUTPUT)/olm
VENV ?= $(OUTPUT)/venv


#------------#
#- Building -#
#------------#

all: build
	@echo "✅ all"
.PHONY: all

build: cli image gen-olm
	@echo "✅ build"
.PHONY: build

cli: operator-sdk gen
	operator-sdk up local --operator-flags "version"
	@echo "✅ cli"
.PHONY: cli

image: operator-sdk gen
	operator-sdk build $(IMAGE)
	@echo "✅ image"
.PHONY: image

vendor:
	go mod tidy
	go mod vendor
	@echo "✅ vendor"
.PHONY: vendor

run: operator-sdk gen
	operator-sdk up local --operator-flags "operator run"
.PHONY: run

clean:
	rm -rf $(OUTPUT)
	rm -rf vendor/
	@echo "✅ clean"
.PHONY: clean

release:
	docker push $(IMAGE)
	docker push $(CATALOG_IMAGE)
	@echo "✅ docker push"
	mkdir -p build-releases
	cp build/_output/bin/noobaa-operator build-releases/noobaa-linux-v$(VERSION)
	@echo "✅ build-releases/noobaa-linux-v$(VERSION)"
	cp build/_output/bin/noobaa-operator-local build-releases/noobaa-mac-v$(VERSION)
	@echo "✅ build-releases/noobaa-mac-v$(VERSION)"
.PHONY: release

operator-sdk:
	@echo "checking operator-sdk version"
	operator-sdk version | grep -q "operator-sdk version: v0.10.1, commit: 872e7d997486bb587660fc8d6226eaab8b5c1087"
	@echo "✅ operator-sdk"
.PHONY: operator-sdk

#------------#
#- Generate -#
#------------#

gen: vendor pkg/bundle/deploy.go
	@echo "✅ gen"
.PHONY: gen

pkg/bundle/deploy.go: pkg/bundler/bundler.go $(shell find deploy/ -type f)
	mkdir -p pkg/bundle
	go run pkg/bundler/bundler.go deploy/ pkg/bundle/deploy.go

gen-api: operator-sdk gen
	$(TIME) operator-sdk generate k8s
	$(TIME) operator-sdk generate openapi
	@echo "✅ gen-api"
.PHONY: gen-api

gen-api-fail-if-dirty: gen-api
	git diff --exit-code || ( \
		echo "Build failed: gen-api is not up to date."; \
		echo "Run 'make gen-api' and update your PR.";  \
		exit 1; \
	)
.PHONY: gen-api-fail-if-dirty

gen-olm: operator-sdk gen
	rm -rf $(OLM)
	operator-sdk up local --operator-flags "olm catalog -n my-noobaa-operator --dir $(OLM)"
	python3 -m venv $(VENV) && \
		. $(VENV)/bin/activate && \
		pip3 install --upgrade pip && \
		pip3 install operator-courier==2.1.7 && \
		operator-courier --verbose verify --ui_validate_io $(OLM)
	docker build -t $(CATALOG_IMAGE) -f build/catalog-source.Dockerfile .
	@echo "✅ gen-olm"
.PHONY: gen-olm


#-----------#
#- Testing -#
#-----------#

test: lint test-go
	@echo "✅ test"
.PHONY: test

lint: gen
	$(TIME) go run golang.org/x/lint/golint \
		-set_exit_status=1 \
		$$(go list ./... | cut -d'/' -f5- | sed 's/^\(.*\)$$/\.\/\1\//' | grep -v ./pkg/apis/noobaa/v1alpha1/ | grep -v ./pkg/bundle/)
	@echo
	$(TIME) go run golang.org/x/lint/golint \
		-set_exit_status=1 \
		$$(echo ./pkg/apis/noobaa/v1alpha1/* | tr ' ' '\n' | grep -v '/zz_generated')
	@echo "✅ lint"
.PHONY: lint

test-go: gen cli
	$(TIME) go test ./pkg/... ./cmd/... ./version/...
	@echo "✅ test-go"
.PHONY: test-go

test-cli-flow:
	$(TIME) ./test/cli/test_cli_flow.sh
	@echo "✅ test-cli-flow"
.PHONY: test-cli-flow

# test-olm runs tests for the OLM package
test-olm: operator-sdk gen-olm
	$(TIME) ./test/test-olm.sh $(CATALOG_IMAGE)
	@echo "✅ test-olm"
.PHONY: test-olm
