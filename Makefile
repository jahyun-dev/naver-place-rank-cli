SHELL := /bin/sh

BINARY_NAME := naver-place-rank
DIST_DIR := dist

.PHONY: build test fmt lint tidy clean release-snapshot release check

build:
	go build -o $(BINARY_NAME) ./

test:
	go test ./...

fmt:
	gofmt -w *.go

lint:
	golangci-lint run

tidy:
	go mod tidy

clean:
	rm -rf $(BINARY_NAME) $(DIST_DIR)

release-snapshot:
	goreleaser release --snapshot --clean

release:
	goreleaser release --clean

check:
	goreleaser check
