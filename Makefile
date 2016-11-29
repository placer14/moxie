all: dev

dev:
	./bin/go_dev

clean:
	docker rm -fv go_dev go_dev_volumes; true #ignore errors if containers don't exist
	docker rmi go_dev
