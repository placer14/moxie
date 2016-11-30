# moxie

## Purpose

A proxy server which forks traffic based on the path of the request.

## Usage

*moxie* expects that you have the following dependancies available in
the $PATH on your host:

- bash
- gnumake (tested with 3.81)
- docker (tested with 1.12)

*gnumake* is used to manage the development resources and are described
below.

### Testing

`make test` will create the dev environment, and then `go get` the
test dependencies and execute all the tests from within the environment.

### Building Development

`make dev` will create a container inside which you have full access
to an isolated golang environment called *go_dev*.

When the environment is already present, `make dev` will start and
attach you to the existing environment. Exiting all instances of from an
existing environment will destroy it, causing `make dev` to produce a
new environment on next execution.

### Building Production

`make prod` will produce a new Docker image called `moxie:latest` which
has the built moxie.go binary as the entrypoint. This will also run
tests and execute a container based on the image for added confidence!

Exiting from the container will cause it to be destroyed, however the
image `moxie:latest` will persist for further use.

Each execution of `make prod` will destroy the existing `moxie:latest` image
and produce a new one. (Note: Intermediate images used during the creation of
the production image are not destroyed.)

This is the default target and may be called with just `make`.

### Cleaning up

`make clean` will destroy the containers and images created by the
Makefile and attempts to provide a clean zero-state for your host.

### Example Development Cadence

The local repository is mounted to `/go/src/moxie` within the
development environment provided with standard golang tools. A typical
development cadence may look like:

1. **Make changes**. Using your editor of choice from your host. The
working directory is mounted within *go_dev* and will mirror changes you
make in the filesystem.
2. **Create and attach to development environment.** `make dev` which creates
the env (if missing) and then attaches you to your newly created *go_dev*.
3. **Run tests.** `go test <path>` within *go_dev* or `make test` on local host.
4. **Or execute moxie.** `go run moxie.go` within *go_dev* will start moxie
on port 8080. When `make dev` attaches you to *go_dev*, it maps your
localhost's port 8080 to the container's port 8080. This allows you to
`curl` from inside or outside of the container.
5. **Or build a production image.** `make prod` which tests, creates and runs
server binary.

## Development Tools

There are a few tools which may be useful for use in development.

### echohttpd

This is a dummy endpoint useful for testing moxie in a safe environment. When
run, it will listen on default port 8000 and echo the received request
in its HTTP/1.x wire representation into the body of a HTTP 200 OK response as
well as on STDOUT in the tty the process was started on.

From within *go_dev*:
`go run echohttpd/echohttpd.go`

Multiple copies of this may be used by providing each copy with a unique
port to run on via the `-p` flag.

From within *go_dev*:
`go run echohttpd/echohttpd.go -p 9999`

### Makefile

A Note: Chaining make targets together such (for example `make clean test`)
does not work. This is because the expression within `ifeq` is expanded on
inital execution and not after each target is completed.

### Docker

Under the hood of the Makefile, we are manipulating Docker containers to
provide our golang environment. The Makefile shows how you might
manipulate the containers to do other things that may not already be
provided by the existing make targets.
