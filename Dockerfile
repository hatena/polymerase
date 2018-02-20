FROM golang:latest

ENV USER root
RUN curl -sSL https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh > /wait-for-it.sh && \
    chmod +x /wait-for-it.sh

ENV PKG github.com/hatena/polymerase
WORKDIR /go/src/$PKG
ADD pkg /go/src/$PKG/pkg
ADD vendor /go/src/$PKG/vendor
ADD Makefile main.go /go/src/$PKG/
RUN go install
ENV PATH /go/bin:$PATH
