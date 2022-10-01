.PHONY: build
build:
	go build -v ./cmd/gophermart

.DEFAULT_GOAL := build
