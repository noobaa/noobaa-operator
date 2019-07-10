VERSION ?= 0.1.0
IMAGE ?= noobaa/noobaa-operator:$(VERSION)
REPO ?= github.com/noobaa/noobaa-operator

GO_FLAGS ?= CGO_ENABLED=0 GO111MODULE=on
GO_LINUX_FLAGS ?= GOOS=linux GOARCH=amd64

OUTPUT = build/_output
BIN = $(OUTPUT)/bin
BUNDLE = $(OUTPUT)/bundle
VENV = $(OUTPUT)/venv

# Default tasks:

all: build
	@echo "@@@ All Done."
.PHONY: all

# Developer tasks:

lint:
	@echo Linting...
	go get -u golang.org/x/lint/golint
	golint -set_exit_status=1 $(shell go list ./cmd/... ./pkg/...)
.PHONY: lint

build: cli image
	@echo "@@@ Build Done."
.PHONY: build

cli: vendor
	${GO_FLAGS} go generate cmd/cli/cli.go
	${GO_FLAGS} go build -mod=vendor -o $(BIN)/kubectl-noobaa $(REPO)/cmd/cli
.PHONY: cli

dev: operator
	WATCH_NAMESPACE=noobaa OPERATOR_NAME=noobaa-operator $(BIN)/noobaa-operator-local
.PHONY: dev

operator: vendor
	go build -mod=vendor -o $(BIN)/noobaa-operator-local $(REPO)/cmd/manager
.PHONY: operator

gen: vendor
	operator-sdk generate k8s
	operator-sdk generate openapi
.PHONY: gen

vendor:
	${GO_FLAGS} go mod vendor
.PHONY: vendor

image:
	${GO_FLAGS} ${GO_LINUX_FLAGS} go build -mod=vendor -o $(BIN)/noobaa-operator $(REPO)/cmd/manager
	docker build -f build/Dockerfile -t $(IMAGE) .
.PHONY: image

bundle:
	mkdir -p $(BUNDLE)
	cp deploy/olm-catalog/noobaa-operator/*.yaml $(BUNDLE)/
	cp deploy/olm-catalog/noobaa-operator/$(VERSION)/*.yaml $(BUNDLE)/
	cp deploy/crds/*crd.yaml $(BUNDLE)/
	( python3 -m venv $(VENV) && . $(VENV)/bin/activate && pip install operator-courier >/dev/null )
	( . $(VENV)/bin/activate && operator-courier verify --ui_validate_io $(BUNDLE) )
.PHONY: bundle

push:
	@echo TODO: To which tag do we want to push here?
	@echo       version from version/version.go? master? git describe?
	# docker push $(IMAGE)
.PHONY: push

test: vendor
	go test ./cli/... ./pkg/...
.PHONY: test

test-e2e: vendor
	operator-sdk test local ./test/e2e \
		--global-manifest deploy/cluster_role_binding.yaml \
		--debug \
		--go-test-flags "-v -parallel=1"
.PHONY: test

clean:
	rm $(BIN)/*
	rm -rf vendor/
.PHONY: clean

# Deps

install-sdk:
	@echo Installing SDK ${SDK_VERSION}
	curl https://github.com/operator-framework/operator-sdk/releases/download/${SDK_VERSION}/operator-sdk-${SDK_VERSION}-x86_64-linux-gnu -sLo ${GOPATH}/bin/operator-sdk
	chmod +x ${GOPATH}/bin/operator-sdk
.PHONY: install-sdk

#TODO scorecard 
