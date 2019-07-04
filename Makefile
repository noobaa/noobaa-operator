
BIN = build/_output/bin
REPO = github.com/noobaa/noobaa-operator
DEV_IMAGE = noobaa-operator
IMAGE = noobaa/noobaa-operator

# Default tasks:

all: build
	@echo "@@@ All Done."
.PHONY: all

# Developer tasks:

build: cli image
	@echo "@@@ Build Done."
.PHONY: build

cli: vendor
	go build -mod=vendor -o $(BIN)/kubectl-noobaa $(REPO)/cmd/cli
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
	go mod vendor
.PHONY: vendor

image:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o $(BIN)/noobaa-operator $(REPO)/cmd/manager
	docker build -f build/Dockerfile -t $(DEV_IMAGE) .
.PHONY: image

push:
	docker push $(IMAGE)
.PHONY: push

test:
	go test ./...
.PHONY: test

clean:
	rm $(BIN)/*
	rm -rf vendor/
	docker rmi $(IMAGE)
.PHONY: clean

