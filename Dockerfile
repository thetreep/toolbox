FROM        golang:1.15-alpine3.12 as base

RUN         apk -u add git openssh build-base
WORKDIR     /go/src/github.com/thetreep/toolbox
ADD         .   /go/src/github.com/thetreep/toolbox
COPY        scripts/dep .
RUN         chmod +x dep; ./dep

FROM        base as builder
RUN         make build
