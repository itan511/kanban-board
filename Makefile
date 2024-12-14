run: build
	@./bin/kanban-board

build:
	@go build -o bin/kanban-board ./cmd/main.go
