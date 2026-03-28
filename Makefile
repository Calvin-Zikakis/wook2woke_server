BINARY := wook2woke_server

ifneq (,$(wildcard .env))
  include .env
  export
endif

WOKE_SCORE  ?= 7
DESCRIPTION ?= Test upload from Makefile
HOST        ?= http://localhost:8080

.PHONY: build run test clean docker-build docker-up docker-down upload delete-all

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

delete-all:
	curl -X DELETE $(HOST)/api/entries \
	  -H "X-API-Key: $(API_KEY)"

upload:
ifndef PHOTO
	$(error PHOTO is required. Usage: make upload PHOTO=/path/to/image.jpg)
endif
	curl -X POST $(HOST)/api/upload \
	  -H "X-API-Key: $(API_KEY)" \
	  -F "wokeScore=$(WOKE_SCORE)" \
	  -F "description=$(DESCRIPTION)" \
	  -F "photo=@$(PHOTO)"
