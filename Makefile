# Hap - the simple and effective provisioner
# Copyright (c) 2015 Garrett Woodworth (https://github.com/gwoo)
.PHONY: bin/hap test

bin/hap: bin
	go build -v -o $@ ./cmd/hap

bin/hap-linux-amd64: bin
	GOOS=linux GOARCH=amd64 go build -v -o $@ ./cmd/hap

test:
	go test -v ./...

bin:
	mkdir bin