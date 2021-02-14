# Copyright (c) 2013-2019 by Michael Dvorkin and contributors. All Rights Reserved.
# Use of this source code is governed by a MIT-style license that can
# be found in the LICENSE file.

CMD_DIR  = $(CURDIR)/cmd
COMMANDS = $(wildcard ${CMD_DIR}/*/main.go)
BIN_DIR  = $(CURDIR)/bin
BINS     = $(foreach cmd,${COMMANDS},$(notdir $(abspath $(dir ${cmd}))))
PACKAGE  = github.com/mop-tracker/mop
VERSION  = 0.2.0
LDFLAGS  = -X ${PACKAGE}/internal.Version=${VERSION}

default: vendor build

build: 
	@$Q for bin in $(BINS); do \
		echo "Building $$bin $(VERSION)"; \
		GO111MODULE=on go build \
			-tags release \
			-ldflags '$(LDFLAGS)' \
			-o $(BIN_DIR)/$$bin $(CMD_DIR)/$$bin; \
	done

.PHONY: clean
clean:
	@echo "Cleaning up"
	@rm -rf $(BIN_DIR)
	@rm -rf coverage.out

format:
	@echo "Formatting"
	@go fmt ./...

install:
	go install -x $(PACKAGE)

buildall:
	GOOS=darwin  GOARCH=amd64 go build $(GOFLAGS) -o ./bin/mop-$(VERSION)-osx-64         $(PACKAGE)
	GOOS=freebsd GOARCH=amd64 go build $(GOFLAGS) -o ./bin/mop-$(VERSION)-freebsd-64     $(PACKAGE)
	GOOS=linux   GOARCH=amd64 go build $(GOFLAGS) -o ./bin/mop-$(VERSION)-linux-64       $(PACKAGE)
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o ./bin/mop-$(VERSION)-windows-64.exe $(PACKAGE)
	GOOS=windows GOARCH=386   go build $(GOFLAGS) -o ./bin/mop-$(VERSION)-windows-32.exe $(PACKAGE)

.PHONY: run
run:
	go run ./cmd/mop/main.go

test:
	@go test ./...

.PHONY: vendor
vendor:
	@go mod tidy
	@go mod vendor

.PHONY: vet
vet:
	@echo "Running vet"
	@go vet ./...
