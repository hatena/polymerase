FROM golang:1.8.3 AS build-env

ENV PKG github.com/taku-k/polymerase
WORKDIR /go/src/$PKG
ADD pkg /go/src/$PKG/pkg
ADD vendor /go/src/$PKG/vendor
ADD Makefile main.go /go/src/$PKG/
RUN go install

FROM busybox
ENV PATH /go/bin:$PATH
COPY --from=build-env /go/bin/polymerase /go/bin/polymerase
