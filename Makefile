# Hap - the simple and effective provisioner
# Copyright (c) 2014 Garrett Woodworth (https://github.com/gwoo)
.PHONY: bin/hap test

bin/hap: bin
	go build -v -o $@ ./cmd/hap

test:
	go test -v ./...

bin:
	mkdir bin