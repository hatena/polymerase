NAME     := xtralab

PROTO := protoc

SRCS    := $(shell find . -type f -name '*.go')
PROTOSRCS := $(shell find . -type f -name '*.proto')
MOCKS := pkg/storage/storage.go
LDFLAGS := -ldflags="-s -w -extldflags \"-static\""

.DEFAULT_GOAL := bin/$(NAME)

bin/$(NAME): $(SRCS)
	go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o bin/$(NAME)

.PHONY: cross-build
cross-build: deps
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)_darwin_amd64
	GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)_darwin_386
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)_linux_amd64
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)_linux_386
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)_windows_amd64.exe
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo $(LDFLAGS) -o dist/$(NAME)_windows_386.exe

.PHONY: glide
glide:
ifeq ($(shell command -v glide 2> /dev/null),)
    curl https://glide.sh/get | sh
endif

.PHONY: deps
deps: glide
	glide install

.PHONY: proto
proto: $(PROTOSRCS)
	$(PROTO) -I pkg pkg/backup/proto/backup.proto --go_out=plugins=grpc:pkg

.PHONY: mockgen
mockgen: $(MOCKS)
	$(MOCKGEN) -source pkg/storage/storage.go -destination pkg/storage/mock.go
