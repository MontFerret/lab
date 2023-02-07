VERSION ?= $(shell sh versions.sh lab)
FERRET_VERSION = $(shell sh versions.sh ferret)
DIR_BIN = ./bin
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
		-ldflags "-X main.version=${VERSION} -X github.com/MontFerret/lab/runtime.version=${FERRET_VERSION}" \
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
	goimports -w -local github.com/MontFerret ./cdn ./cmd ./reporters ./runner ./runtime ./sources ./testing

lint:
	staticcheck ./... && \
	revive -config revive.toml -formatter stylish -exclude ./pkg/parser/fql/... -exclude ./vendor/... ./...

vet:
	go vet ./...

release:
ifeq ($(RELEASE_VERSION), )
	$(error "Release version is required (RELEASE_VERSION)")
else ifeq ($(FERRET_VERSION), )
	$(error "Ferret version is required")
else ifeq ($(GITHUB_TOKEN), )
	$(error "GitHub token is required (GITHUB_TOKEN)")
else
	rm -rf ./dist && \
	git tag -a v$(RELEASE_VERSION) -m "New $(RELEASE_VERSION) version" && \
	git push origin v$(RELEASE_VERSION) && \
	goreleaser
endif