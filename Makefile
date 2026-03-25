BINARY := wook2woke_server

ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: build run test clean docker-build docker-up docker-down

build:
	CGO_ENABLED=1 go build -o $(BINARY) .

run: build
	./$(BINARY)

test:
	go test -v ./...

clean:
	rm -f $(BINARY)

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down
