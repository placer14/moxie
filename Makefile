all: moxie

test:
	docker-compose run --rm --entrypoint go moxie get -t ./...
	docker-compose run --rm --entrypoint go moxie test ./...

moxie:
	@echo "\n== Building moxie...\n"
	@echo "\n== Running tests...\n"
	@make test
	@echo "\n== Building binary...\n"
	docker-compose run --rm --entrypoint go moxie build -o moxie moxie.go
	@echo "\n== Creating final image moxie_production:latest...\n"
	docker build -f environments/Dockerfile.moxie --no-cache -t moxie_production:latest .
	rm moxie
	@echo "\n== moxie_production:latest created!\n"

httpecho:
	@echo "\n== Building httpecho image...\n"
	docker-compose run --rm --entrypoint go moxie build -o httpecho tools/httpecho.go
	docker build -f environments/Dockerfile.httpecho --no-cache -t httpecho:latest .
	rm httpecho
	@echo "\n== httpecho:latest created!\n"

websocketecho:
	@echo "\n== Building websocketecho image...\n"
	@cd tools/websocket_echo_server && docker build --no-cache -t websocketecho:latest .
	@echo "\n== websocketecho:latest created!\n"

clean:
	@docker rmi websocketecho httpecho moxie_production; true
	@docker-compose down -v --rmi all
