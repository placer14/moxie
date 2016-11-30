all: prod

test:
ifeq ($(shell docker images -q go_dev 2> /dev/null),)
	@make create_dev
endif
	docker run --rm --volumes-from go_dev_volumes go_dev get -t ./...
	docker run --rm --volumes-from go_dev_volumes go_dev test ./...

prod:
	@echo "\n== Running tests..."
	@make test
	@echo "\n== Building binary..."
	docker run --rm --volumes-from go_dev_volumes go_dev build -o moxie moxie.go
	@echo "\n== Creating final image moxie:latest..."
	docker build -t moxie:latest -f environments/Dockerfile.prod .
	rm moxie
	@echo "\n== Complete. All systems GO! (Press Ctrl+C to exit.)\n"
	docker run --rm -p 8080:8080/tcp moxie:latest

dev:
ifeq ($(shell docker images -q go_dev 2> /dev/null),)
	@make create_dev
endif
ifeq ($(shell docker ps -q --filter="ancestor=go_dev" 2> /dev/null),)
	docker run --rm -it --volumes-from go_dev_volumes -p 8080:8080/tcp --entrypoint "/bin/bash" go_dev
else
	docker exec -it `docker ps --filter="ancestor=go_dev" -q` bash
endif

create_dev:
	@echo "\n== Go development image is not present. Creating..."
	docker build -t go_dev -f environments/Dockerfile.dev .
	@echo "\n== Creating golang dev container..."
	docker create --name go_dev_volumes -v /go -v "$(PWD)":/go/src/moxie golang:1.7
	docker run --rm --volumes-from go_dev_volumes go_dev get -t

clean:
	@echo "\n== Removing containers..."
	@docker rm -fv go_dev go_dev_volumes; true #ignore errors if containers don't exist
	@echo "\n== Removing images..."
	@docker rmi -f go_dev
	@docker rmi -f moxie
