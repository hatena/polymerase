NAME     := polymerase

PROTO := protoc
MOCKGEN := mockgen

SRCS    := $(shell find . -type f -name '*.go')
PROTOSRCS := $(shell find . -type f -name '*.proto' | grep -v -e vendor)
MOCKS := pkg/storage/storage.go
LDFLAGS := -ldflags="-s -w -extldflags \"-static\""

.DEFAULT_GOAL := bin/$(NAME)

bin/$(NAME): $(SRCS)
	go build -o bin/$(NAME)

.PHONY: cross-build
cross-build: deps
	GOOS=darwin GOARCH=amd64 go build -o dist/$(NAME)_darwin_amd64
	GOOS=linux GOARCH=amd64 go build -o dist/$(NAME)_linux_amd64

.PHONY: linux
linux: deps
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)_linux_amd64

.PHONY: glide
glide:
ifeq ($(shell command -v glide 2> /dev/null),)
    curl https://glide.sh/get | sh
endif

.PHONY: deps
deps: glide
	go get github.com/golang/mock/mockgen

.PHONY: proto
proto: $(PROTOSRCS)
	for src in $(PROTOSRCS); do \
	  $(PROTO) \
	   -Ipkg \
	   $$src \
	   --go_out=plugins=grpc:pkg; \
	done

.PHONY: mockgen
mockgen: $(MOCKS)
	$(MOCKGEN) -source pkg/storage/storage.go -destination pkg/storage/mock.go -package storage

.PHONY: vet
vet:
	go tool vet -all -printfuncs=Wrap,Wrapf,Errorf $$(find . -maxdepth 1 -mindepth 1 -type d | grep -v -e "^\.\/\." -e vendor)

.PHONY: test
test:
	go test -cover -v $$(glide nv)

.PHONY: test-race
test-race:
	go test -v -race $$(glide nv)

.PHONY: test-all
test-all: vet test-race

.PHONY: fmt
fmt:
	gofmt -s -w $$(find . -type f -name '*.go' | grep -v -e vendor)

.PHONY: imports
imports:
	goimports -w $$(find . -type f -name '*.go' | grep -v -e vendor)

