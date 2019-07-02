BIN = build/_output/bin
REPO = github.com/noobaa/noobaa-operator

all: build

build:
	go build -mod=vendor -o $(BIN)/noobaa-operator $(REPO)/cmd/manager

image: build
	docker build -f build/Dockerfile .

.PHONY: all build image
