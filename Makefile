.PHONY: build
build:
	go build -v ./cmd/gophermart

.PHONY: test
test:
	go clean -testcache
	go test -cover -v ./internal/store/sqlstore
	go test -cover -v ./pkg/...
	go test -cover -v ./internal/server


.DEFAULT_GOAL := build
