PACKAGE := github.com/hstreamdb/dev-deploy

export GO_BUILD=GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) go build -ldflags '-s -w'

all: dev-deploy

fmt:
	gofmt -s -w -l `find . -name '*.go' -print`

dev-deploy:
	$(GO_BUILD) -o bin/dev-deploy $(PACKAGE)/cmd

.PHONY: fmt, dev-deploy, all