all: prod

test:
	docker-compose run --rm --entrypoint go moxie get -t ./...
	docker-compose run --rm --entrypoint go moxie test ./...

prod:
	@echo "\n== Running tests..."
	@make test
	@echo "\n== Building binary..."
	docker-compose run --rm --entrypoint go moxie build -o moxie moxie.go
	@echo "\n== Creating final image moxie_production:latest..."
	docker build -f environments/Dockerfile.moxie --no-cache -t moxie_production:latest .
	rm moxie
	@echo "\n== Complete."

echo:
	@echo "\n== Building echohttpd image..."
	docker-compose run --rm --entrypoint go moxie build -o echohttpd tools/echohttpd.go
	docker build -f environments/Dockerfile.echohttpd --no-cache -t echohttpd:latest .
	rm echohttpd
	@echo "\n== Complete."

clean:
	@docker-compose down -v --rmi all
