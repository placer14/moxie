# moxie

## Purpose

A proxy server which forks traffic based on the path of the request.

## Usage

*moxie* expects that you have the following dependancies available in
the $PATH on your host:

- bash
- gnumake (tested with 3.81)
- docker (tested with 1.12)
- docker-compose (tested with 1.9.0)

*gnumake* is used to manage the development resources and are described
below.

### Testing

`make test` will create the dev environment, and then `go get` the
test dependencies and execute all the tests from within the environment.

### Building Development

`make dev` will create a container inside which you have full access
to an isolated container environment called *dev*.

Alternatively, you can run a command inside your container directly from
the prompt by using `docker-compose run --rm --entrypoint <command> dev
<arguments>`, like this:

`docker-compose run --rm --entrypoint go dev test ./...`

or you can just type it without the `--entrypoint` since `go` is already
the default entrypoint of *dev* which you can see by examining
`environments/Dockerfile.dev`.

Note: Make targets output their commands so you can copy and run them
directly. Don't worry about making a mess as `make clean` does a good
job cleaning up.

### Building Production

`make prod` will produce a new Docker image called `production:latest` which
has the built moxie.go binary as the entrypoint. This will also run
tests and execute a container based on the image for added confidence!

Exiting from the container will cause it to be destroyed, however the
image `moxie_production:latest` will persist for further use.

Each execution of `make prod` will destroy the existing `moxie_production:latest` image
and produce a new one. (Note: Intermediate images used during the creation of
the production image are not destroyed.)

This is the default target and may be called with just `make`.

### Cleaning up

`make clean` will destroy the containers and images created by the
docker-compose and attempts to provide a clean zero-state for your host.

### Example Development Cadence

The local repository is mounted to `/go/src/moxie` within the
development environment provided with standard golang tools. A typical
development cadence may look like:

1. **Make changes**. Using your editor of choice from your host. The
working directory is mounted within *dev* and will mirror changes you
make in the filesystem.
2. **Create and attach to development environment.** `make dev` which creates
the env (if missing) and then attaches you to your newly created *dev*.
3. **Run tests.** `go test <path>` within *dev* or `make test` on local host.
4. **Or execute moxie.** `go run moxie.go` within *dev* will start moxie
on port 8080. When `make dev` attaches you to *dev*, it maps your
localhost's port 8080 to the container's port 8080. This allows you to
`curl` from inside or outside of the container.
5. **Or build a production image.** `make prod` which tests, creates and runs
server binary.

## Development Tools

There are a few tools which may be useful for use in development.

### moxie

This is manages setting up the proxyHandler and handing it off to the
golang http server. There are a few flag available to change its
behavior:

> `--port <valid port>`

define which port the proxy should bind to on the local host

> `--proxied-host <IP or FQDN>`

define a new host to recieve proxied traffic when no routes match the request

> `--proxied-port <valid port>`

define a new port to recieve proxied traffic when no routes match the request

### echohttpd

This is a dummy endpoint useful for testing moxie in a safe environment. When
run, it will listen on default port 8000 and echo the received request
in its HTTP/1.x wire representation into the body of a HTTP 200 OK response as
well as on STDOUT in the tty the process was started on.

From within *dev*:
`go run tools/echohttpd.go`

Multiple copies of this may be used by providing each copy with a unique
port to run on via the `--port` flag.

From within *dev*:
`go run tools/echohttpd.go --port 9999`

Alternatively, you can setup both containers ready to to talk to each
other.

`docker-compose -f environments/docker-compose-echo.yml up`

### Makefile

A Note: Chaining make targets together such (for example `make clean test`)
does not work. This is because the expression within `ifeq` is expanded on
inital execution and not after each target is completed.

### docker-compose

Under the hood of the Makefile, we are using docker-compose to orchestrate
Docker containers to provide our golang environment. You may run your
own commands against *dev* using the following incantations:

*Run a go subcommand* `docker-compose run dev test ./...`

The default starting command of the container (the entrypoint) is the go
binary and will accept go subcommands by default. This runs the test
subcommand on the repo mounted inside of the container.

*Start a bash prompt inside container* `docker-compose run --entrypoint /bin/bash dev`

This changes the default starting command of the container (the
entrypoint) to another command, in this case, bash.

*Run another go tool inside container* `docker-compose run --entrypoint gofmt dev moxie.go`

This will run the gofmt tool on the moxie.go file mounted within the
container.

