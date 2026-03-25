.PHONY: build
build:
	@go build -o bin/promptd ./cmd/promptd/main.go

.PHONY: clean
clean:
	@rm -rf bin/
