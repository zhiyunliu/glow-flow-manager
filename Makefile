.PHONY: all test test-web lint build build-web clean install start help

export GIN_MODE=release
GLUE_GLOBAL := github.com/zhiyunliu/glue
GO_BIN ?= go
CI_BIN := bin/manager-glow
LINT_BIN ?= golangci-lint
OPT_DIR := /opt/manager-glow
OPT_ETC := $(OPT_DIR)/etc
WEB_DIR := ./web

GIT_REV := $(shell git rev-list --tags --max-count=1 2>/dev/null)
GIT_TAG := $(shell git describe --tags $(GIT_REV) 2>/dev/null)
GIT_COMMIT := $(shell git log --pretty=format:'%h' -1 2>/dev/null)
ifeq ($(GIT_TAG),)
	VERSION=dev
else
	VERSION=$(GIT_TAG)
endif
BUILD_TIME := $(shell date +%Y.%m.%d.%H:%M:%S)
PKG_VERSION := $(shell cat version.txt)

LDFLAGS := -X "$(GLUE_GLOBAL)/global.GitCommit=$(GIT_COMMIT)" \
-X "$(GLUE_GLOBAL)/global.BuildTime=$(BUILD_TIME)" \
-X "$(GLUE_GLOBAL)/global.PkgVersion=$(PKG_VERSION)" \
-X "$(GLUE_GLOBAL)/global.Version=$(VERSION)"

all: build-web build

test:
	$(GO_BIN) vet ./...
	$(GO_BIN) test -count=1 ./...

test-web:
	cd $(WEB_DIR) && npm test

lint:
	$(LINT_BIN) run ./...

build-web:
	cd $(WEB_DIR) && npm ci && npm run build

build:
	$(GO_BIN) build -ldflags '$(LDFLAGS)' -o $(CI_BIN) .

clean:
	rm -f $(CI_BIN)

install:
	mkdir -p $(DESTDIR)$(OPT_DIR)/bin
	mkdir -p $(DESTDIR)$(OPT_ETC)
	mkdir -p $(DESTDIR)$(OPT_DIR)/log
	install -m 755 -D $(CI_BIN) $(DESTDIR)$(OPT_DIR)/$(CI_BIN)
	cp *.json $(DESTDIR)$(OPT_ETC)/

start:
	$(OPT_DIR)/$(CI_BIN) run -C $(OPT_ETC)/config.json

help:
	@echo make: build web and service
	@echo make build-web: build the embedded admin UI
	@echo make build: build the service binary
	@echo make test: go vet and unit tests
	@echo make test-web: frontend unit tests
	@echo make install: deploy to $(OPT_DIR)
	@echo make start: start service
