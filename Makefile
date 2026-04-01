VERSION ?= $(shell sh ./scripts/versions.sh lab)
FERRET_VERSION = $(shell sh ./scripts/versions.sh ferret)
DIR_BIN = ./bin
DIR_PKG = ./pkg
DIR_CMD = ./cmd
NAME = lab

default: build

build: vet test compile

install-tools:
	go install honnef.co/go/tools/cmd/staticcheck@latest && \
	go install golang.org/x/tools/cmd/goimports@latest && \
	go install github.com/mgechev/revive@latest

install:
	go get

compile:
	go build -v -o ${DIR_BIN}/${NAME} \
		-ldflags "-X main.version=${VERSION} -X github.com/MontFerret/lab/pkg/runtime.version=${FERRET_VERSION}" \
	./main.go

test:
	go test ./...

cover:
	go test -race -coverprofile=coverage.txt -covermode=atomic ... && \
	curl -s https://codecov.io/bash | bash

doc:
	godoc -http=:6060 -index

fmt:
	go fmt ./... && \
	goimports -w -local github.com/MontFerret ${DIR_PKG} ${DIR_CMD} main.go

lint:
	staticcheck ./... && \
	revive -config revive.toml -formatter stylish -exclude ./pkg/parser/fql/... -exclude ./vendor/... ./...

vet:
	go vet ./...

release:
	./scripts/release.sh $(TAG)