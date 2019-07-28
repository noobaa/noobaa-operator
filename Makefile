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

all: local build
	@echo "all - done."
.PHONY: all

build: image olm
	@echo "build - done."
.PHONY: build

image: gen
	operator-sdk build $(IMAGE)
.PHONY: image

gen: vendor $(BUNDLE)/deploy.go
.PHONY: gen

$(BUNDLE)/deploy.go: pkg/bundle/bundle.go $(shell find deploy/ -type f)
	mkdir -p $(BUNDLE)
	$(GO) run pkg/bundle/bundle.go deploy/ $(BUNDLE)/deploy.go

vendor:
	$(GO) mod vendor
.PHONY: vendor

olm:
	rm -rf $(OLM)
	mkdir -p $(OLM)
	cp deploy/olm-catalog/package/* $(OLM)/
	cp deploy/crds/*crd.yaml $(OLM)/
	python3 -m venv $(VENV)
	( \
		. $(VENV)/bin/activate && \
		pip3 install --upgrade pip && \
		pip3 install operator-courier && \
		operator-courier verify --ui_validate_io $(OLM) \
	)
	@echo "olm - ready at $(OLM)"
.PHONY: olm


#-----------#
#- Testing -#
#-----------#

test: lint unittest
	@echo "test - done."
.PHONY: test

test-integ: scorecard test-e2e
	@echo "test-integ - done."
.PHONY: test-integ

lint: gen
	go get -u golang.org/x/lint/golint
	golint -set_exit_status=0 ./cmd/... ./pkg/... ./test/... ./version/...
.PHONY: lint

unittest: gen
	go test ./cmd/... ./pkg/... ./version/...
.PHONY: unittest

test-e2e: gen
	# TODO fix test-e2e !
	# operator-sdk test local ./test/e2e \
	# 	--global-manifest deploy/cluster_role_binding.yaml \
	# 	--debug \
	# 	--go-test-flags "-v -parallel=1"
.PHONY: test-e2e

scorecard: gen
	kubectl create ns test-noobaa-operator-scorecard
	operator-sdk scorecard \
		--cr-manifest deploy/crds/noobaa_v1alpha1_noobaa_cr.yaml \
		--cr-manifest deploy/crds/noobaa_v1alpha1_backingstore_cr.yaml \
		--cr-manifest deploy/crds/noobaa_v1alpha1_bucketclass_cr.yaml \
		--csv-path deploy/olm-catalog/package/noobaa-operator.v$(VERSION).clusterserviceversion.yaml \
		--namespace test-noobaa-operator-scorecard
.PHONY: scorecard


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

local: gen
	operator-sdk up local &>/dev/null
.PHONY: local

run: gen
	operator-sdk up local --operator-flags "operator run"
.PHONY: run

gen-api: gen
	operator-sdk generate k8s
	operator-sdk generate openapi
.PHONY: gen-api

gen-api-fail-if-dirty: gen-api
	git diff -s --exit-code pkg/apis/noobaa/v1alpha1/zz_generated.deepcopy.go || (echo "Build failed: API has been changed but the deep copy functions aren't up to date. Run 'make gen-api' and update your PR." && exit 1)
	git diff -s --exit-code pkg/apis/noobaa/v1alpha1/zz_generated.openapi.go || (echo "Build failed: API has been changed but the deep copy functions aren't up to date. Run 'make gen-api' and update your PR." && exit 1)
.PHONY: gen-api-fail-if-dirty

clean:
	rm -rf $(OUTPUT)
	rm -rf vendor/
.PHONY: clean
