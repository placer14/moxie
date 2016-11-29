# moxie

## Purpose

A proxy server which forks traffic based on the path of the request.

## Testing

`make test` will create the dev environment, and then `go get` the
test dependencies and execute all the tests from within the environment.

## Building Development

`make dev` will create a container inside which you have full access
to an isolated golang environment.

When the environment is already present, `make dev` will start and
attach you to the existing environment.

## Building Production

`make prod` will produce a new Docker image called `moxie:latest` which
has the built moxie.go binary as the entrypoint. This will also run
tests and execute a container based on the image for added confidence!

## Cleaning up

`make clean` will destroy the containers and images created by the
Makefile and attempts to provide a clean zero-state for your host.

## Makefile note

Chaining make targets together such (for example `make clean test`) does not
work. This is because the expression within `ifeq` is expanded on inital
execution and not after each target is completed.
