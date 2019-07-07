
BIN = build/_output/bin
REPO = github.com/noobaa/noobaa-operator
IMAGE = noobaa/noobaa-operator:master

GO_FLAGS ?= CGO_ENABLED=0 GO111MODULE=on
GO_LINUX_FLAGS ?= GOOS=linux GOARCH=amd64

# Default tasks:

all: build
	@echo "@@@ All Done."
.PHONY: all

# Developer tasks:

lint:
	@echo Linting...
	golint -set_exit_status=1 $(shell go list ./cmd/... ./pkg/...)
.PHONY: lint

build: cli image
	@echo "@@@ Build Done."
.PHONY: build

cli: vendor
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

push:
	docker push $(IMAGE)
.PHONY: push

test: vendor
	go test ./...
.PHONY: test

clean:
	rm $(BIN)/*
	rm -rf vendor/
.PHONY: clean

# Deps

install-tools:
	go get -u golang.org/x/lint/golint
.PHONY: install-tools


install-sdk:
	@echo Installing SDK ${SDK_VERSION}
	curl https://github.com/operator-framework/operator-sdk/releases/download/${SDK_VERSION}/operator-sdk-${SDK_VERSION}-x86_64-linux-gnu -sLo ${GOPATH}/bin/operator-sdk
	chmod +x ${GOPATH}/bin/operator-sdk
.PHONY: install-sdk


#TODO scorecard 
