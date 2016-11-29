all: test

test:
	@make create_dev
	docker run --rm -it --volumes-from go_dev_volumes go_dev get -t ./proxyhandler
	docker run --rm -it --volumes-from go_dev_volumes go_dev test ./proxyhandler

dev:
ifeq ($(shell docker images -q go_dev 2> /dev/null),)
	@make create_dev
endif
	docker run --rm -it --volumes-from go_dev_volumes -p 8080:8080/tcp --entrypoint "/bin/bash" go_dev

create_dev:
	@echo "\nGo development image is not present. Creating..."
	docker build -t go_dev .
	@echo "\nCreating golang dev container..."
	docker create --name go_dev_volumes -v /go -v "$(PWD)":/go/src/moxie golang:1.7
	docker run --rm --volumes-from go_dev_volumes go_dev get -t

clean:
	@echo "Removing containers..."
	@docker rm -fv go_dev go_dev_volumes; true #ignore errors if containers don't exist
	@echo "Removing image..."
	@docker rmi -f go_dev
