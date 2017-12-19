NAME     := polymerase
VERSION  := $(shell git describe --tags --exact-match 2> /dev/null || git rev-parse --short HEAD || echo "unknown")
REVISION := $(shell git rev-parse HEAD)

PROTO := protoc
MOCKGEN := mockgen

SRCS    := $(shell find . -type f -name '*.go')
PROTOSRCS := $(shell find . -type f -name '*.proto' | grep -v -e vendor)
MOCKS := pkg/storage/storage.go
LINUX_LDFLAGS := -s -w -extldflags "-static"
DARWIN_LDFLAGS := -s -w
LINKFLAGS := \
	-X "github.com/taku-k/polymerase/pkg/build.tag=$(VERSION)" \
	-X "github.com/taku-k/polymerase/pkg/build.rev=$(REVISION)"
override LINUX_LDFLAGS += $(LINKFLAGS)
override DARWIN_LDFLAGS += $(LINKFLAGS)


.DEFAULT_GOAL := bin/$(NAME)

bin/$(NAME): $(SRCS)
	go build -ldflags '$(DARWIN_LDFLAGS)' -o bin/$(NAME)

.PHONY: cross-build
cross-build: deps
	GOOS=darwin GOARCH=amd64 go build -ldflags '$(DARWIN_LDFLAGS)' -o dist/$(NAME)_darwin_amd64
	GOOS=linux GOARCH=amd64 go build -a -tags netgo -installsuffix netgo -ldflags '$(LINUX_LDFLAGS)' -o dist/$(NAME)_linux_amd64

.PHONY: linux
linux: deps
	GOOS=linux GOARCH=amd64 go build -a -tags netgo -installsuffix netgo -ldflags '$(LINUX_LDFLAGS)' -o dist/$(NAME)_linux_amd64

.PHONY: deps
deps:
	go get github.com/golang/mock/mockgen

.PHONY: proto
proto: $(PROTOSRCS)
	for src in $(PROTOSRCS); do \
	  $(PROTO) \
	    -Ipkg \
	    -Ivendor \
	    -I$$GOPATH/src \
	    $$src \
	    --gofast_out=plugins=grpc:pkg; \
	done

.PHONY: mockgen
mockgen: $(MOCKS)
	$(MOCKGEN) -source pkg/storage/storage.go -destination pkg/storage/mock.go -package storage

.PHONY: vet
vet:
	go tool vet -all -printfuncs=Wrap,Wrapf,Errorf $$(find . -maxdepth 1 -mindepth 1 -type d | grep -v -e "^\.\/\." -e vendor)

.PHONY: test
test:
	go test -cover -v ./pkg/...

.PHONY: test-race
test-race:
	echo "" > coverage.txt
	for d in $$(go list ./... | grep -v vendor); do \
	  go test -v -race -coverprofile=profile.out -covermode=atomic $$d; \
	  if [ -f profile.out ]; then \
	    cat profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi \
	done

.PHONY: test-integration
test-integration:
	./integration-test/test.sh

.PHONY: test-all
test-all: vet test test-race test-integration

.PHONY: fmt
fmt:
	gofmt -s -w $$(find . -type f -name '*.go' | grep -v -e vendor)

.PHONY: imports
imports:
	goimports -w $$(find . -type f -name '*.go' | grep -v -e vendor)

.PHONY: docker-test-base
docker-test-base:
	docker build -t takuk/polymerase-test-base -f integration-test/sut/Dockerfile.testbase .
	docker push takuk/polymerase-test-base

