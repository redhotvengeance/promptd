.PHONY: build
build: deps
	@go build -o bin/promptd ./cmd/promptd/main.go

.PHONY: clean
clean:
	go clean
	rm -rf ./bin/
	rm -rf ./internal/libsql/data/local.db
	find ./internal/libsql/data/sql ! -name '.gitkeep' -type f -exec rm -f {} +

.PHONY: deps
deps:
	sqlc generate

.PHONY: migrate
migrate: deps
	atlas migrate diff --dev-url "sqlite://dev?mode=memory" --to "file://./internal/libsql/data/schema.sql" --dir "file://./internal/libsql/data/migrations"
	atlas migrate apply --dir "file://./internal/libsql/data/migrations" --url "sqlite://./internal/libsql/data/local.db"
