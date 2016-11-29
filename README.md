# moxie

## Purpose

A proxy server which forks traffic based on the path of the request.

## Building a Dev Environment

`make dev` or more succinctly `make` will create a container inside
which you have full access to an isolated golang environment.

When the environment is already present, `make` will start and attach
you to the existing environment.

## Destroying a Dev Environment

`make clean` will destroy the containers and images for the golang
dev environment.


