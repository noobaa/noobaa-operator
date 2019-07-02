BIN = build/_output/bin
REPO = github.com/noobaa/noobaa-operator

all: build

build:
	go build -mod=vendor -o $(BIN)/noobaa-operator $(REPO)/cmd/manager

vendor:
	go mod vendor

image: build
	docker build -f build/Dockerfile -t noobaa-operator .

.PHONY: all build vendor image 
