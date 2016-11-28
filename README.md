# moxie

## Purpose

A proxy server which forks traffic based on the path of the request.

## Building a new production image

Note: Building this assumes you have go locally setup with GOPATH and
GOBIN configured in your environment.

1. Run `make build`
2. Docker image `placer14/proxy:latest` will be created and a copy of
   the binary `proxy` will be located in your local `GOBIN` path.

## Documentation

Visit https://godoc.org/github.com/placer14/proxyhandler or run

`godoc github.com/placer14/proxyhandler`


