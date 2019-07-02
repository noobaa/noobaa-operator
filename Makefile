
BIN = build/_output/bin
REPO = github.com/noobaa/noobaa-operator
IMAGE = noobaa/noobaa-operator:master

all: cli image

build: cli operator

image:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o $(BIN)/noobaa-operator $(REPO)/cmd/manager
	docker build -f build/Dockerfile -t $(IMAGE) .

cli: vendor
	go build -mod=vendor -o $(BIN)/noobaa $(REPO)/cmd/cli

operator: vendor
	go build -mod=vendor -o $(BIN)/noobaa-operator-local $(REPO)/cmd/manager

vendor:
	go mod vendor

test:
	go test ./...

clean:
	rm $(BIN)/*
	rm -rf vendor/
	docker rmi $(IMAGE)

.PHONY: \
	all \
	build \
	image \
	cli \
	operator \
	vendor \
	test \
	clean
