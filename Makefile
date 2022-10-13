.PHONY: build
build:
	go build -v ./cmd/gophermart

.PHONY: test
test:
	go test -cover -v ./...

.DEFAULT_GOAL := build
