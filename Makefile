# Hap - the simple and effective provisioner
# Copyright (c) 2019 GWoo (https://github.com/gwoo)

.PHONY: bin/hap bin/hap-linux-amd64 bin/hap-darwin-amd64 test
VERSION := $(shell git describe --always --dirty --tags)
VERSION_FLAGS := -ldflags "-X main.Version=$(VERSION)"

bin/hap: bin
	go build -v $(VERSION_FLAGS) -o $@ ./cmd/hap

bin/hap-linux-amd64: bin
	GOOS=linux GOARCH=amd64 go build -v $(VERSION_FLAGS) -o $@ ./cmd/hap

bin/hap-darwin-amd64: bin
	GOOS=darwin GOARCH=amd64 go build -v $(VERSION_FLAGS) -o $@ ./cmd/hap

test:
	go test -v ./...

bin:
	mkdir bin