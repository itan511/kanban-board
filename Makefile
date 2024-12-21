build-db:
	docker-compose build db

build-app:
	docker-compose build app

run-db:
	docker-compose up -d db

run-app:
	docker-compose up -d app

all: build-db build-app run-db run-app

clean:
	docker-compose down -v

stop:
	docker-compose stop

local-build:
	@go build -o bin/kanban-board ./cmd/main.go

local-run: local-build
	@./bin/kanban-board