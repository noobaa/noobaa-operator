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

GO_LINUX ?= GOOS=linux GOARCH=amd64
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
	@echo "✅ all"
.PHONY: all

build: cli image gen-olm
	@echo "✅ build"
.PHONY: build

cli: gen
	operator-sdk up local --operator-flags "version"
	@echo "✅ cli"
.PHONY: cli

image: gen
	operator-sdk build $(IMAGE)
	@echo "✅ image"
.PHONY: image

vendor:
	mkdir -p $(BUNDLE)
	echo "package bundle" > $(BUNDLE)/tmp.go
	go mod vendor
	@echo "✅ vendor"
.PHONY: vendor

run: gen
	operator-sdk up local --operator-flags "operator run"
.PHONY: run

clean:
	rm -rf $(OUTPUT)
	rm -rf vendor/
	@echo "✅ clean"
.PHONY: clean

release:
	docker push $(IMAGE)
	@echo "✅ docker push"
	mkdir -p build-releases
	cp build/_output/bin/noobaa-operator build-releases/noobaa-linux-v$(VERSION)
	@echo "✅ build-releases/noobaa-linux-v$(VERSION)"
	cp build/_output/bin/noobaa-operator-local build-releases/noobaa-mac-v$(VERSION)
	@echo "✅ build-releases/noobaa-mac-v$(VERSION)"
.PHONY: release


#------------#
#- Generate -#
#------------#

gen: vendor $(BUNDLE)/deploy.go
	@echo "✅ gen"
.PHONY: gen

$(BUNDLE)/deploy.go: pkg/bundle/bundle.go $(shell find deploy/ -type f)
	mkdir -p $(BUNDLE)
	go run pkg/bundle/bundle.go deploy/ $(BUNDLE)/deploy.go

gen-api: gen
	$(TIME) operator-sdk generate k8s
	$(TIME) operator-sdk generate openapi
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
	mkdir -p $(OLM)
	cp deploy/crds/*_crd.yaml $(OLM)/
	operator-sdk up local --operator-flags "olm csv -n my-noobaa-operator" \
		> $(OLM)/noobaa-operator.v$(VERSION).clusterserviceversion.yaml
	operator-sdk up local --operator-flags "olm package -n my-noobaa-operator" \
		> $(OLM)/noobaa-operator.package.yaml
	python3 -m venv $(VENV) && \
		. $(VENV)/bin/activate && \
		pip3 install --upgrade pip && \
		pip3 install operator-courier && \
		operator-courier verify --ui_validate_io $(OLM)
	@echo "✅ gen-olm"
.PHONY: gen-olm


#-----------#
#- Testing -#
#-----------#

test: lint unittest
	@echo "✅ test"
.PHONY: test

lint: gen
	$(TIME) go run golang.org/x/lint/golint \
		-set_exit_status=1 \
		$$(go list ./... | cut -d'/' -f4- | sed 's/^\(.*\)$$/\.\/\1\//' | grep -v ./pkg/apis/noobaa/v1alpha1/)
	@echo
	$(TIME) go run golang.org/x/lint/golint \
		-set_exit_status=1 \
		$$(echo ./pkg/apis/noobaa/v1alpha1/* | tr ' ' '\n' | grep -v '/zz_generated')
	@echo "✅ lint"
.PHONY: lint

unittest: gen
	$(TIME) go test ./pkg/... ./cmd/... ./version/...
	@echo "✅ unittest"
.PHONY: unittest

test-cli:
	$(TIME) go test ./test/cli/...
	@echo "✅ test-cli"
.PHONY: test-cli

test-csv: gen-olm
	operator-sdk alpha olm install || exit 0
	kubectl create ns my-noobaa-operator || exit 0
	operator-sdk up local --operator-flags "crd create -n my-noobaa-operator"
	operator-sdk up local --operator-flags "operator install --no-deploy -n my-noobaa-operator"
	kubectl apply -f deploy/olm-catalog/operator-group.yaml
	kubectl apply -f $(OLM)/noobaa-operator.v$(VERSION).clusterserviceversion.yaml
	sleep 30
	kubectl wait pod -n my-noobaa-operator -l noobaa-operator=deployment --for condition=ready
	@echo "✅ test-csv"
.PHONY: test-csv

test-olm: gen-olm
	./test/test-olm.sh
	@echo "✅ test-olm"
.PHONY: test-olm

test-scorecard: $(OUTPUT)/olm-global-manifest.yaml gen-olm
	kubectl create ns noobaa-scorecard || exit 0
	$(TIME) operator-sdk scorecard --verbose \
		--csv-path $(OLM)/noobaa-operator.v$(VERSION).clusterserviceversion.yaml \
		--crds-dir $(OLM)/ \
		--cr-manifest deploy/crds/noobaa_v1alpha1_noobaa_cr.yaml \
		--cr-manifest deploy/crds/noobaa_v1alpha1_backingstore_cr.yaml \
		--cr-manifest deploy/crds/noobaa_v1alpha1_bucketclass_cr.yaml \
		--global-manifest $(OUTPUT)/olm-global-manifest.yaml \
		--namespace noobaa-scorecard
	@echo "✅ test-scorecard"
.PHONY: test-scorecard

$(OUTPUT)/olm-global-manifest.yaml: gen
	operator-sdk up local --operator-flags "crd yaml" > $(OUTPUT)/olm-global-manifest.yaml
	echo "---" >> $(OUTPUT)/olm-global-manifest.yaml
	cat deploy/cluster_role.yaml >> $(OUTPUT)/olm-global-manifest.yaml
	echo "---" >> $(OUTPUT)/olm-global-manifest.yaml
	yq write deploy/cluster_role_binding.yaml subjects.0.namespace noobaa-scorecard >> $(OUTPUT)/olm-global-manifest.yaml
	echo "---" >> $(OUTPUT)/olm-global-manifest.yaml
	yq write deploy/obc/storage_class.yaml provisioner "noobaa.io/noobaa-scorecard.bucket" >> $(OUTPUT)/olm-global-manifest.yaml

# TODO operator-sdk test local is not working on CI !
# test-e2e: gen
# 	operator-sdk test local ./test/e2e \
# 		--global-manifest deploy/cluster_role_binding.yaml \
# 		--debug \
# 		--go-test-flags "-v -parallel=1"
# .PHONY: test-e2e
