all: prod

test:
	docker-compose run --rm dev get -t ./...
	docker-compose run --rm dev test ./...

prod:
	@echo "\n== Running tests..."
	@make test
	@echo "\n== Building binary..."
	docker-compose run --rm dev build -o moxie moxie.go
	@echo "\n== Creating final image moxie:latest..."
	docker-compose build production
	rm moxie
	@echo "\n== Complete. All systems GO! (Press Ctrl+C to exit.)\n"
	docker-compose run --rm production

dev:
	docker-compose run --rm --entrypoint /bin/bash dev

rebuild_dev:
	docker-compose build --no-cache dev
	docker-compose run --rm dev get -t ./...

clean:
	@echo "\n== Removing containers..."
	@docker-compose down -v --rmi all
