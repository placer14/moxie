FROM golang:latest
MAINTAINER Mike Greenberg <mg@nobulb.com>

RUN mkdir -p /go/src/app
WORKDIR /go/src/app

ENTRYPOINT ["/bin/bash"]
