# moxie

## Purpose

A proxy server which forks traffic based on the path of the request.

## Testing

`make test` will create the dev environment, and then `go get` the
test dependencies and execute all the tests from within the environment.

## Building Development

`make dev` will create a container inside which you have full access
to an isolated golang environment called `go_dev`.

When the environment is already present, `make dev` will start and
attach you to the existing environment.

## Development Cycle

The local repository is mounted to `/go/src/moxie` within the
development environment provided with standard golang tools. A typical
development cadence may look like:

1. Make changes to local repository.
2. `make dev` which creates (if missing) and attaches you to go_dev
3. `make test` from host or `go test <path>` within go_dev
4. `go run moxie.go` within go_dev will start moxie for testing
5. (optionally) `make prod` which tests, creates and runs server binary

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
