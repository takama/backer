# Copyright 2018 Igor Dolzhikov. All rights reserved.
# Use of this source code is governed by a MIT-style
# license that can be found in the LICENSE file.

PROJECT=github.com/takama/backer

# Use the 0.0.0 tag for testing, it shouldn't clobber any release builds
RELEASE?=v0.3.4

BUILDTAGS=

all: test

GO_LIST_FILES=$(shell go list ${PROJECT}/... | grep -v vendor)

fmt:
	@echo "+ $@"
	@go list -f '{{if len .TestGoFiles}}"gofmt -s -l {{.Dir}}"{{end}}' ${GO_LIST_FILES} | xargs -L 1 sh -c

lint: bootstrap
	@echo "+ $@"
	@go list -f '{{if len .TestGoFiles}}"golint -min_confidence=0.85 {{.Dir}}/..."{{end}}' ${GO_LIST_FILES} | xargs -L 1 sh -c

vet:
	@echo "+ $@"
	@go vet ${GO_LIST_FILES}

test: fmt lint vet
	@echo "+ $@"
	@go test -v -race -cover -tags "$(BUILDTAGS) cgo" ${GO_LIST_FILES}

cover:
	@echo "+ $@"
	@> coverage.txt
	@go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}} && cat {{.Dir}}/.coverprofile  >> coverage.txt"{{end}}' ${GO_LIST_FILES} | xargs -L 1 sh -c

HAS_LINT := $(shell command -v golint;)

bootstrap:
ifndef HAS_LINT
	go get -u github.com/golang/lint/golint
endif

.PHONY: all fmt lint vet test cover bootstrap
