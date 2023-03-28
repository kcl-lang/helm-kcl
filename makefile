HELM_HOME ?= $(shell helm home)
VERSION := $(shell sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' plugin.yaml)

HELM_3_PLUGINS := $(shell bash -c 'eval $$(helm env); echo $$HELM_PLUGINS')

PKG:= kusionstack.io/helm-kcl
LDFLAGS := -X $(PKG)/cmd.Version=$(VERSION)

# Clear the "unreleased" string in BuildMetadata
LDFLAGS += -X k8s.io/helm/pkg/version.BuildMetadata=
LDFLAGS += -X k8s.io/helm/pkg/version.Version=$(shell ./scripts/dep-helm-version.sh)

GO ?= go

.PHONY: format
format:
	test -z "$$(find . -type f -o -name '*.go' -exec gofmt -d {} + | tee /dev/stderr)" || \
	test -z "$$(find . -type f -o -name '*.go' -exec gofmt -w {} + | tee /dev/stderr)"

.PHONY: install
install: build
	mkdir -p $(HELM_HOME)/plugins/helm-kcl/bin
	cp -f bin/kcl $(HELM_HOME)/plugins/helm-kcl/bin
	cp -f plugin.yaml $(HELM_HOME)/plugins/helm-kcl/

.PHONY: install/helm3
install/helm3: build
	mkdir -p $(HELM_3_PLUGINS)/helm-kcl/bin
	cp -f bin/kcl $(HELM_3_PLUGINS)/helm-kcl/bin
	cp -f plugin.yaml $(HELM_3_PLUGINS)/helm-kcl/

.PHONY: lint
lint:
	scripts/update-gofmt.sh
	scripts/verify-gofmt.sh
	scripts/verify-golint.sh
	scripts/verify-govet.sh

.PHONY: build
build: lint
	mkdir -p bin/
	go build -v -o bin/kcl -ldflags="$(LDFLAGS)"

.PHONY: test
test:
	go test -v ./...

.PHONY: bootstrap
bootstrap:
	go mod download
	command -v golint || GO111MODULE=off go get -u golang.org/x/lint/golint

.PHONY: docker-run-release
docker-run-release: export pkg=/go/src/github.com/databus23/helm-kcl
docker-run-release:
	git checkout main
	git push
	docker run -it --rm -e GITHUB_TOKEN -v $(shell pwd):$(pkg) -w $(pkg) golang:1.18.1 make bootstrap release

.PHONY: dist
dist: export COPYFILE_DISABLE=1 #teach OSX tar to not put ._* files in tar archive
dist: export CGO_ENABLED=0
dist:
	rm -rf build/kcl/* release/*
	mkdir -p build/kcl/bin release/
	cp -f README.md LICENSE plugin.yaml build/kcl
	GOOS=linux GOARCH=amd64 $(GO) build -o build/kcl/bin/kcl -trimpath -ldflags="$(LDFLAGS)"
	tar -C build/ -zcvf $(CURDIR)/release/helm-kcl-linux-amd64.tgz kcl/
	GOOS=linux GOARCH=arm64 $(GO) build -o build/kcl/bin/kcl -trimpath -ldflags="$(LDFLAGS)"
	tar -C build/ -zcvf $(CURDIR)/release/helm-kcl-linux-arm64.tgz kcl/
	GOOS=freebsd GOARCH=amd64 $(GO) build -o build/kcl/bin/kcl -trimpath -ldflags="$(LDFLAGS)"
	tar -C build/ -zcvf $(CURDIR)/release/helm-kcl-freebsd-amd64.tgz kcl/
	GOOS=darwin GOARCH=amd64 $(GO) build -o build/kcl/bin/kcl -trimpath -ldflags="$(LDFLAGS)"
	tar -C build/ -zcvf $(CURDIR)/release/helm-kcl-macos-amd64.tgz kcl/
	GOOS=darwin GOARCH=arm64 $(GO) build -o build/kcl/bin/kcl -trimpath -ldflags="$(LDFLAGS)"
	tar -C build/ -zcvf $(CURDIR)/release/helm-kcl-macos-arm64.tgz kcl/
	rm build/kcl/bin/kcl
	GOOS=windows GOARCH=amd64 $(GO) build -o build/kcl/bin/kcl.exe -trimpath -ldflags="$(LDFLAGS)"
	tar -C build/ -zcvf $(CURDIR)/release/helm-kcl-windows-amd64.tgz kcl/

.PHONY: release
release: lint dist
	scripts/release.sh v$(VERSION)

# Test for the plugin installation with `helm plugin install -v THIS_BRANCH` works
# Useful for verifying modified `install-binary.sh` still works against various environments
.PHONY: test-plugin-installation
test-plugin-installation:
	docker build -f testdata/Dockerfile.install .
