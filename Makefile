# Required build-tools:
#   - go
#   - git
#   - python3
#   - minikube
#   - operator-sdk

VERSION ?= $(shell go run cmd/version/main.go)
IMAGE ?= noobaa/noobaa-operator:$(VERSION)
REPO ?= github.com/noobaa/noobaa-operator

GO ?= CGO_ENABLED=0 GO111MODULE=on go
GO_LINUX ?= GOOS=linux GOARCH=amd64 $(GO)
GO_LIST ?= ./cmd/... ./pkg/... ./test/... ./version/...
GOHOSTOS ?= $(shell go env GOHOSTOS)

OUTPUT ?= build/_output
BIN ?= $(OUTPUT)/bin
OLM ?= $(OUTPUT)/olm
VENV ?= $(OUTPUT)/venv
BUNDLE ?= $(OUTPUT)/bundle


#------------#
#- Building -#
#------------#

all: build
	@echo "all - done."
.PHONY: all

build: cli image
	@echo "build - done."
.PHONY: build

cli: gen
	operator-sdk up local --operator-flags "version"
	@echo "cli - done."
.PHONY: cli

image: gen
	operator-sdk build $(IMAGE)
	@echo "image - done."
.PHONY: image

gen: vendor $(BUNDLE)/deploy.go
.PHONY: gen

$(BUNDLE)/deploy.go: pkg/bundle/bundle.go $(shell find deploy/ -type f)
	mkdir -p $(BUNDLE)
	$(GO) run pkg/bundle/bundle.go deploy/ $(BUNDLE)/deploy.go

vendor:
	mkdir -p $(BUNDLE)
	echo "package bundle" > $(BUNDLE)/empty.go
	$(GO) mod vendor
.PHONY: vendor


#-----------#
#- Testing -#
#-----------#

test: lint unittest
	@echo "test - done."
.PHONY: test

test-integ: scorecard test-cli
	@echo "test-integ - done."
.PHONY: test-integ

lint: gen
	go get -u golang.org/x/lint/golint
	golint -set_exit_status=0 ./cmd/... ./pkg/... ./test/... ./version/...
.PHONY: lint

unittest: gen
	go test ./pkg/...
.PHONY: unittest

test-cli: cli
	go test ./test/
.PHONY: test-cli

# TODO operator-sdk test local is not working on CI !
# test-e2e: gen
# 	operator-sdk test local ./test/e2e \
# 		--global-manifest deploy/cluster_role_binding.yaml \
# 		--debug \
# 		--go-test-flags "-v -parallel=1"
# .PHONY: test-e2e

scorecard: gen
	kubectl create ns test-noobaa-operator-scorecard
	operator-sdk scorecard \
		--cr-manifest deploy/crds/noobaa_v1alpha1_noobaa_cr.yaml \
		--cr-manifest deploy/crds/noobaa_v1alpha1_backingstore_cr.yaml \
		--cr-manifest deploy/crds/noobaa_v1alpha1_bucketclass_cr.yaml \
		--csv-path deploy/olm-catalog/package/noobaa-operator.v$(VERSION).clusterserviceversion.yaml \
		--namespace test-noobaa-operator-scorecard
.PHONY: scorecard

test-olm:
	rm -rf $(OLM)
	mkdir -p $(OLM)
	cp  deploy/olm-catalog/package/* \
		deploy/crds/*_crd.yaml \
		deploy/obc/*_crd.yaml \
		$(OLM)/
	./test/test-olm.sh
	@echo "olm - done."
.PHONY: test-olm

#-------------#
#- Releasing -#
#-------------#

push-image:
	@echo "push-image - It is too risky to push like this, because it easily overrides.
	@echo       version from version/version.go? master? git describe?
	# docker push $(IMAGE)
.PHONY: push-image


#--------------#
#- Developing -#
#--------------#

run: gen
	operator-sdk up local --operator-flags "operator run"
.PHONY: run

gen-api: gen
	operator-sdk generate k8s
	operator-sdk generate openapi
.PHONY: gen-api

gen-api-fail-if-dirty: gen-api
	git diff --exit-code || ( \
		echo "Build failed: gen-api is not up to date."; \
		echo "Run 'make gen-api' and update your PR.";  \
		exit 1; \
	)
.PHONY: gen-api-fail-if-dirty

clean:
	rm -rf $(OUTPUT)
	rm -rf vendor/
.PHONY: clean
