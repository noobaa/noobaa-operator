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

export OPERATOR_SDK_VERSION ?= v0.17.0
export OPERATOR_SDK ?= build/_tools/operator-sdk-$(OPERATOR_SDK_VERSION)

KUBECONFIG ?= build/empty-kubeconfig

#------------#
#- Building -#
#------------#

all: build
	@echo "✅ all"
.PHONY: all

build: cli image gen-olm
	@echo "✅ build"
.PHONY: build

cli: $(OPERATOR_SDK) gen
	$(OPERATOR_SDK) run --kubeconfig="$(KUBECONFIG)" --local --operator-flags "version"
	@echo "✅ cli"
.PHONY: cli

image: $(OPERATOR_SDK) gen
	$(OPERATOR_SDK) build $(IMAGE)
	@echo "✅ image"
.PHONY: image

vendor:
	go mod tidy
	go mod vendor
	@echo "✅ vendor"
.PHONY: vendor

run: $(OPERATOR_SDK) gen
	$(OPERATOR_SDK) run --local --operator-flags "operator run"
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

$(OPERATOR_SDK):
	bash build/install-operator-sdk.sh
	@echo "✅ $(OPERATOR_SDK)"


#------------#
#- Generate -#
#------------#

gen: vendor pkg/bundle/deploy.go
	@echo "✅ gen"
.PHONY: gen

pkg/bundle/deploy.go: pkg/bundler/bundler.go version/version.go $(shell find deploy/ -type f)
	mkdir -p pkg/bundle
	go run pkg/bundler/bundler.go deploy/ pkg/bundle/deploy.go

gen-api: $(OPERATOR_SDK) gen
	$(TIME) $(OPERATOR_SDK) generate k8s
	$(TIME) $(OPERATOR_SDK) generate crds
	@echo "✅ gen-api"
.PHONY: gen-api

gen-api-fail-if-dirty: gen-api
	git diff --exit-code || ( \
		echo "Build failed: gen-api is not up to date."; \
		echo "Run 'make gen-api' and update your PR.";  \
		exit 1; \
	)
.PHONY: gen-api-fail-if-dirty

gen-olm: $(OPERATOR_SDK) gen
	rm -rf $(OLM)
	$(OPERATOR_SDK) run --kubeconfig="$(KUBECONFIG)" --local --operator-flags "olm catalog -n my-noobaa-operator --dir $(OLM)"
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
test-olm: $(OPERATOR_SDK) gen-olm
	$(TIME) ./test/test-olm.sh $(CATALOG_IMAGE)
	@echo "✅ test-olm"
.PHONY: test-olm
