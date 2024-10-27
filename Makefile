# Required build-tools:
#   - go
#   - git
#   - python3
#   - minikube
#   - docker

export GO111MODULE=on
export GOPROXY:=https://proxy.golang.org

TIME ?= time -p
ARCH ?= $(shell uname -m)

VERSION ?= $(shell go run cmd/version/main.go)
IMAGE ?= noobaa/noobaa-operator:$(VERSION)
DEV_IMAGE ?= noobaa/noobaa-operator-dev:$(VERSION)
REPO ?= github.com/noobaa/noobaa-operator
CATALOG_IMAGE ?= noobaa/noobaa-operator-catalog:$(VERSION)
BUNDLE_IMAGE ?= noobaa/noobaa-operator-bundle:$(VERSION)
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

CONTROLLER_GEN_VERSION=v0.16.3
DEEPCOPY_GEN_VERSION=v0.29.3

GO_LINUX ?= GOOS=linux GOARCH=amd64
GOHOSTOS ?= $(shell go env GOHOSTOS)

MK_PATH:=$(dir $(realpath $(lastword $(MAKEFILE_LIST))))
MK_PARENT:=$(realpath $(MK_PATH)../)
OUTPUT ?= build/_output
BIN ?= $(OUTPUT)/bin
OLM ?= $(OUTPUT)/olm
MANIFESTS ?= $(OUTPUT)/manifests
obc-crd ?= required
cosi-sidecar-image ?= "gcr.io/k8s-staging-sig-storage/objectstorage-sidecar/objectstorage-sidecar:v20221117-v0.1.0-22-g0e67387"
VENV ?= $(OUTPUT)/venv
CMD_MANAGER ?= cmd/manager/main.go

export NOOBAA_OPERATOR_LOCAL ?= $(BIN)/noobaa-operator-local
# OPERATOR_SDK is to install olm only.
export OPERATOR_SDK_VERSION ?= v0.17.2
export OPERATOR_SDK ?= build/_tools/operator-sdk-$(OPERATOR_SDK_VERSION)

#------------#
#- Building -#
#------------#

all: build
	@echo "✅ all"
.PHONY: all

build: cli image gen-olm
	@echo "✅ build"
.PHONY: build

cli: gen
	go build -o $(NOOBAA_OPERATOR_LOCAL) -mod=vendor $(CMD_MANAGER)
	$(BIN)/noobaa-operator-local version
	@echo "✅ cli"
.PHONY: cli

image: $(docker) gen
	GOOS=linux CGO_ENABLED=$(or ${CGO_ENABLED},0) go build -o $(BIN)/noobaa-operator -gcflags all=-trimpath="$(MK_PARENT)" -asmflags all=-trimpath="$(MK_PARENT)" -mod=vendor $(CMD_MANAGER)
	docker build -f build/Dockerfile -t $(IMAGE) .
	@echo "✅ image"
.PHONY: image

dev-image: $(docker) gen
	go build -o $(BIN)/noobaa-operator -trimpath -mod=vendor -gcflags all=-N -gcflags all=-l $(CMD_MANAGER)
	docker build -f build/Dockerfile -t $(IMAGE) .
	docker build -f build/DockerfileDev --build-arg base_image=$(IMAGE) -t $(DEV_IMAGE) .
	@echo "✅ dev image"
.PHONY: dev-image

vendor:
	go mod tidy
	go mod vendor
	@echo "✅ vendor"
.PHONY: vendor

run: gen
	go build -o $(NOOBAA_OPERATOR_LOCAL) -mod=vendor $(CMD_MANAGER)
	$(BIN)/noobaa-operator-local operator run
.PHONY: run

clean:
	rm -rf $(OUTPUT)
	rm -rf vendor/
	@echo "✅ clean"
.PHONY: clean

$(OPERATOR_SDK):
	bash build/install-operator-sdk.sh
	@echo "✅ $(OPERATOR_SDK)"

release-docker:
	docker push $(IMAGE)
	docker push $(CATALOG_IMAGE)
	@echo "✅ docker push"
.PHONY: release-docker

release-cli:
	mkdir -p build-releases
	cp build/_output/bin/noobaa-operator build-releases/noobaa-linux-v$(VERSION)
	@echo "✅ build-releases/noobaa-linux-v$(VERSION)"
	cp build/_output/bin/noobaa-operator-local build-releases/noobaa-mac-v$(VERSION)
	@echo "✅ build-releases/noobaa-mac-v$(VERSION)"
.PHONY: release-cli

release: release-docker release-cli
.PHONY: release

#------------#
#- Generate -#
#------------#

gen: vendor pkg/bundle/deploy.go
	@echo "✅ gen"
.PHONY: gen

pkg/bundle/deploy.go: pkg/bundler/bundler.go version/version.go $(shell find deploy/ -type f)
	mkdir -p pkg/bundle
	go run pkg/bundler/bundler.go deploy/ pkg/bundle/deploy.go

gen-api: controller-gen deepcopy-gen gen
	$(TIME) $(DEEPCOPY_GEN) --go-header-file="build/hack/boilerplate.go.txt" --input-dirs="./pkg/apis/noobaa/v1alpha1/..." --output-file-base="zz_generated.deepcopy"
	$(TIME) $(CONTROLLER_GEN) paths=./... crd:generateEmbeddedObjectMeta=true output:crd:artifacts:config=deploy/crds/
	@echo "✅ gen-api"
.PHONY: gen-api

gen-api-fail-if-dirty: gen-api
	git diff --exit-code || ( \
		echo "Build failed: gen-api is not up to date."; \
		echo "Run 'make gen-api' and update your PR.";  \
		exit 1; \
	)
.PHONY: gen-api-fail-if-dirty

gen-olm: gen
	rm -rf $(OLM)
	go build -o $(NOOBAA_OPERATOR_LOCAL) -mod=vendor $(CMD_MANAGER)
	$(NOOBAA_OPERATOR_LOCAL) olm catalog -n my-noobaa-operator --dir $(OUTPUT)/olm
	python3 -m venv $(VENV) && \
		. $(VENV)/bin/activate && \
		pip3 install --upgrade pip && \
		pip3 install operator-courier==2.1.11 && \
		operator-courier --verbose verify --ui_validate_io $(OLM)
	docker build -t $(CATALOG_IMAGE) -f build/catalog-source.Dockerfile .
	@echo "✅ gen-olm"
.PHONY: gen-olm

gen-odf-package: cli
	rm -rf $(MANIFESTS)
	MANIFESTS="$(MANIFESTS)" CSV_NAME="$(csv-name)" SKIP_RANGE="$(skip-range)" REPLACES="$(replaces)" CORE_IMAGE="$(core-image)" DB_IMAGE="$(db-image)" OPERATOR_IMAGE="$(operator-image)" COSI_SIDECAR_IMAGE="$(cosi-sidecar-image)" OBC_CRD="$(obc-crd)" PSQL_12_IMAGE="$(psql-12-image)" build/gen-odf-package.sh
	@echo "✅ gen-odf-package"
.PHONY: gen-odf-package

bundle-image: gen-odf-package
	docker build -t $(BUNDLE_IMAGE) -f build/bundle/Dockerfile .

#-----------#
#- Testing -#
#-----------#

test: lint test-go
	@echo "✅ test"
.PHONY: test

golangci-lint: gen
	golangci-lint run --disable-all -E varcheck,structcheck,typecheck,errcheck,gosimple,unused,deadcode,ineffassign,staticcheck --timeout=10m
	@echo "✅ golangci-lint"
.PHONY: golangci-lint

lint: gen
	@echo ""
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run --config .golangci.yml
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

test-cli-flow-dev:
	$(TIME) ./test/cli/test_cli_flow.sh --dev
	@echo "✅ test-cli-flow-dev"
.PHONY: test-cli-flow-dev

test-core-config-map-flow:
	$(TIME) ./test/cli/test_cli_flow.sh --check_core_config_map
	@echo "✅ test-core-config-map-flow"
.PHONY: test-core-config-map-flow

# test-olm runs tests for the OLM package
test-olm: $(OPERATOR_SDK) gen-olm
	$(TIME) ./test/test-olm.sh $(CATALOG_IMAGE)
	@echo "✅ test-olm"
.PHONY: test-olm

test-hac: vendor
	ginkgo -v pkg/controller/ha
	@echo "✅ test-hac"
.PHONY: test-hac

test-kms-dev: vendor
	ginkgo -v pkg/util/kms/test/dev
	@echo "✅ test-kms-dev"
.PHONY: test-kms-dev

test-kms-key-rotate: vendor
	ginkgo -v pkg/util/kms/test/rotate
	@echo "✅ test-kms-key-rotate"
.PHONY: test-kms-key-rotate

test-kms-tls-sa: vendor
	ginkgo -v pkg/util/kms/test/tls-sa
	@echo "✅ test-kms-tls-sa"
.PHONY: test-kms-tls-sa

test-kms-tls-token: vendor
	ginkgo -v pkg/util/kms/test/tls-token
	@echo "✅ test-kms-tls-token"
.PHONY: test-kms-tls-token

test-kms-azure-vault: vendor
	ginkgo -v pkg/util/kms/test/azure-vault
	@echo "✅ test-kms-azure-vault"
.PHONY: test-kms-azure-vault

test-kms-ibm-kp: vendor
	ginkgo -v pkg/util/kms/test/ibm-kp
	@echo "✅ test-kms-ibm-kp"
.PHONY: test-kms-ibm-kp

test-kms-kmip: vendor
	ginkgo -v pkg/util/kms/test/kmip
	@echo "✅ test-kms-kmip"
.PHONY: test-kms-kmip

test-admission: vendor
	ginkgo -v pkg/admission/test/integ
	@echo "✅ test-admission"
.PHONY: test-admission

test-cosi: vendor
	ginkgo -v pkg/cosi
	@echo "✅ test-cosi"
.PHONY: test-cosi

test-bucketclass: vendor
	ginkgo -v pkg/bucketclass pkg/controller/bucketclass
	@echo "✅ test-bucketclass"
.PHONY: test-bucketclass

test-obc: vendor
	ginkgo -v pkg/obc
	@echo "✅ test-obc"
.PHONY: test-obc

test-operator: vendor
	ginkgo -v pkg/operator
	@echo "✅ test-operator"
.PHONY: test-operator

test-util: vendor
	ginkgo -v pkg/util
	@echo "✅ test-util"
.PHONY: test-util

test-validations: 
	ginkgo -v pkg/validations
	@echo "✅ test-validations"
.PHONY: test-validations

# find or download controller-gen if necessary
controller-gen:
ifneq ($(CONTROLLER_GEN_VERSION), $(shell controller-gen --version | awk -F ":" '{print $2}'))
	@{ \
	echo "Installing controller-gen@$(CONTROLLER_GEN_VERSION)" ;\
	set -e ;\
	go install -mod=readonly sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION) ;\
	echo "Installed controller-gen@$(CONTROLLER_GEN_VERSION)" ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

deepcopy-gen:
	@{ \
	echo "Installing deepcopy-gen@$(DEEPCOPY_GEN_VERSION)" ;\
	set -e ;\
	go install -mod=readonly k8s.io/code-generator/cmd/deepcopy-gen@$(DEEPCOPY_GEN_VERSION) ;\
	echo "Installed deepcopy-gen@$(DEEPCOPY_GEN_VERSION)" ;\
	}
DEEPCOPY_GEN=$(GOBIN)/deepcopy-gen
.PHONY: deepcopy-gen

