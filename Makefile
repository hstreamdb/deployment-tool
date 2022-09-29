PACKAGE := github.com/hstreamdb/deployment-tool

export GO111MODULE=on
GO_BUILD=CGO_ENABLED=0 GOOS=$(GOOS) go build -ldflags '-s -w -extldflags "-static"'

all: hdt

fmt:
	gofmt -s -w -l `find . -name '*.go' -print`

hdt:
	$(GO_BUILD) -o bin/hdt $(PACKAGE)/cmd

.PHONY: fmt, hdt, all